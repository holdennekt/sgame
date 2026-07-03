/**
 * sgame load test
 *
 * Usage:
 *   k6 run -e BASE_URL=http://localhost:8080 -e PACK_ID=<id> -e USERS=100 loadtest/k6.js
 *
 * USERS controls the number of players (not counting the host/moderator). Rooms hold
 * up to 10 players, so USERS=100 creates 10 rooms, USERS=500 → 50 rooms, etc. Each
 * room gets one extra VU for the host on top of its players.
 *
 * Each group of 11 VUs maps to one room:
 *   - VU 1, 12, 23 … → host (uses pre-created session from setup)
 *   - VU 2-11, 13-22 … → players (guest accounts)
 *
 * The host starts the game and skips every question so the rooms cycle quickly
 * and keep the pub/sub + WS infrastructure under load the whole duration.
 */

import ws from "k6/ws";
import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Counter, Gauge } from "k6/metrics";

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------
const BASE_URL = (__ENV.BASE_URL || "http://localhost:8080").replace(/\/$/, "");
const WS_BASE = BASE_URL.replace(/^http/, "ws");
const PACK_ID = __ENV.PACK_ID; // required — must be a valid pack ID in the DB
const TOTAL_USERS = parseInt(__ENV.USERS || "100");
const PLAYERS_PER_ROOM = 10;
const ROOMS_COUNT = Math.ceil(TOTAL_USERS / PLAYERS_PER_ROOM);
// +1 per room for the host/moderator VU, which doesn't occupy a player slot.
const VUS_PER_ROOM = PLAYERS_PER_ROOM + 1;
const TOTAL_VUS = ROOMS_COUNT * VUS_PER_ROOM;
// Safety cap for the whole run, in case a room's game hangs and never reaches
// game_over — VUs otherwise stay connected until their game finishes naturally.
const MAX_DURATION = __ENV.MAX_DURATION || "5m";

if (!PACK_ID) {
  throw new Error("PACK_ID env var is required. Set it to a valid pack ID.");
}

// ---------------------------------------------------------------------------
// Custom metrics
// ---------------------------------------------------------------------------
const wsConnectTime = new Trend("ws_connect_time_ms", true);
const wsEventsReceived = new Counter("ws_events_received");
const activeRooms = new Gauge("active_rooms");
// Gap between consecutive room_updated broadcasts on a socket. Since the game
// mostly serializes state changes (one actor advances state, everyone reacts to
// the resulting room_updated before the next action is sent), this approximates
// end-to-end event processing latency without needing to correlate a specific
// send to a specific broadcast.
const wsRoomUpdatedGap = new Trend("ws_room_updated_gap_ms", true);

const ROOM_UPDATED_TRANSITIONS = [
  "selecting_question->revealing_question", // Regular selected, or Auction w/ no eligible bettors
  "selecting_question->passing", // CatInBag, 2+ eligible players
  "selecting_question->answering", // CatInBag, exactly 1 eligible player
  "selecting_question->betting", // Auction, 1+ eligible bettor
  // Round boundary: EndQuestion() (sync, inside ValidateAnswer) lands on
  // selecting_question, then StartNextRegularRound (async, via the QuestionEnded
  // internal event) advances to the next round and re-enters the same state
  "selecting_question->selecting_question",
  "selecting_question->selecting_final_round_category", // round exhausted, final round has 2+ categories
  // StartFinalRound sets SelectingFinalRoundCategory then, in the same transaction,
  // immediately calls chooseFinalRoundCategory if only one category exists
  "selecting_question->final_round_betting", // round exhausted, final round has exactly 1 category
  "selecting_question->game_over", // round exhausted, nobody eligible for final round
  "revealing_question->showing_question", // reveal timer elapses, nobody answered early
  "revealing_question->answering", // start_answer during reveal (falseStartAllowed:true)
  "showing_question->answering", // start_answer after reveal
  "answering->selecting_question", // correct, or last remaining player wrong
  "answering->revealing_question", // wrong, remaining players, still within reveal window
  "answering->showing_question", // wrong, remaining players, past reveal window
  "passing->answering", // pass_question, or auto pass timer
  "betting->answering", // all bets placed, or auto bet timer
  "selecting_final_round_category->selecting_final_round_category", // remove category, not last
  "selecting_final_round_category->final_round_betting", // remove category, last one
  "final_round_betting->showing_final_round_question", // at least one player bet > 0
  "final_round_betting->game_over", // nobody placed a positive bet
  "showing_final_round_question->validating_final_round_answers", // all answers submitted
  "validating_final_round_answers->validating_final_round_answers", // more players left
  "validating_final_round_answers->game_over", // last player validated
];

// No real bar (p95<999999 always passes) — these exist purely to make k6
// print a per-transition breakdown of ws_room_updated_gap_ms in the summary.
function roomUpdatedTransitionThresholds() {
  return Object.fromEntries(
    ROOM_UPDATED_TRANSITIONS.map((t) => [
      `ws_room_updated_gap_ms{transition:${t}}`,
      ["p(95)<999999"],
    ]),
  );
}

// ---------------------------------------------------------------------------
// k6 options
// ---------------------------------------------------------------------------
export const options = {
  scenarios: {
    game_session: {
      executor: "per-vu-iterations",
      vus: TOTAL_VUS,
      iterations: 1,
      maxDuration: MAX_DURATION,
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<1000"],
    http_req_failed: ["rate<0.05"],
    ws_connect_time_ms: ["p(95)<2000"],
    ...roomUpdatedTransitionThresholds(),
  },
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
function jsonHeaders(cookie) {
  const h = { "Content-Type": "application/json" };
  if (cookie) h["Cookie"] = `sessionId=${cookie}`;
  return h;
}

function extractSession(res) {
  const cookies = res.cookies["sessionId"];
  return cookies && cookies.length > 0 ? cookies[0].value : null;
}

function sendWS(socket, event, payload) {
  socket.send(JSON.stringify({ event, payload: payload || {} }));
}

// ---------------------------------------------------------------------------
// setup() — runs once; creates host accounts + rooms
// ---------------------------------------------------------------------------
export function setup() {
  if (!PACK_ID) throw new Error("PACK_ID is required");

  const rooms = [];

  for (let i = 0; i < ROOMS_COUNT; i++) {
    const login = `lt${Date.now().toString(36)}${i}`;
    const password = "Loadtest123!";

    const regRes = http.post(
      `${BASE_URL}/api/register`,
      JSON.stringify({ login, password }),
      { headers: { "Content-Type": "application/json" } },
    );
    check(regRes, { [`room ${i}: host registered`]: (r) => r.status === 200 });
    const hostCookie = extractSession(regRes);
    if (!hostCookie) {
      console.error(`room ${i}: failed to register host`);
      continue;
    }
    const hostUserId = regRes.json("userId");

    // Create room
    const createRes = http.post(
      `${BASE_URL}/api/rooms/`,
      JSON.stringify({
        name: `LT Room ${i}`,
        packId: PACK_ID,
        options: {
          maxPlayers: PLAYERS_PER_ROOM,
          type: "public",
          readingSymbolsPerSecond: 50,
          questionThinkingTime: 5,
          answerThinkingTime: 5,
          questionThinkingTimeFinal: 10,
          falseStartAllowed: true,
          timeToBet: 3,
          timeToPass: 3,
        },
      }),
      { headers: jsonHeaders(hostCookie) },
    );
    check(createRes, { [`room ${i}: created`]: (r) => r.status === 201 });
    const roomId = createRes.json("id");
    if (!roomId) {
      console.error(`room ${i}: failed to create room`, createRes.body);
      continue;
    }

    const hostJoinRes = http.patch(
      `${BASE_URL}/api/rooms/${roomId}/join`,
      null,
      {
        headers: jsonHeaders(hostCookie),
      },
    );
    check(hostJoinRes, {
      [`room ${i}: host joined as moderator`]: (r) => r.status === 200,
    });

    rooms.push({ roomId, hostCookie, hostUserId });
  }

  activeRooms.add(rooms.length);
  console.log(`setup: created ${rooms.length} rooms for ${TOTAL_USERS} users`);
  return rooms;
}

// ---------------------------------------------------------------------------
// VU function
// ---------------------------------------------------------------------------
export default function (rooms) {
  if (!rooms || rooms.length === 0) {
    console.error("no rooms from setup, aborting VU");
    return;
  }

  const vuIndex = __VU - 1; // 0-based
  const roomIndex = Math.floor(vuIndex / VUS_PER_ROOM) % rooms.length;
  const roleInRoom = vuIndex % VUS_PER_ROOM; // 0 = host, 1-10 = player

  const room = rooms[roomIndex];

  if (roleInRoom === 0) {
    runAsHost(room.hostCookie, room.roomId);
  } else {
    runAsPlayer(room.roomId, roleInRoom);
  }
}

// ---------------------------------------------------------------------------
// Host behaviour
// ---------------------------------------------------------------------------
function runAsHost(cookie, roomId) {
  // Wait briefly so player VUs have time to join before the game starts
  sleep(4);

  const wsUrl = `${WS_BASE}/api/ws/room/${roomId}`;
  const startMs = Date.now();

  const res = ws.connect(
    wsUrl,
    { headers: { Cookie: `sessionId=${cookie}` } },
    (socket) => {
      wsConnectTime.add(Date.now() - startMs);

      let lastRoomUpdatedAt = null;
      let lastRoomUpdatedState = null;

      socket.on("open", () => {
        // Small extra delay so lagging player VUs can join
        sleep(1);
        sendWS(socket, "start_game");
      });

      socket.on("message", (data) => {
        const msg = JSON.parse(data);
        wsEventsReceived.add(1);

        if (msg.event !== "room_updated") return;

        const room = msg.payload;
        if (!room) return;

        const now = Date.now();
        if (lastRoomUpdatedAt !== null) {
          wsRoomUpdatedGap.add(now - lastRoomUpdatedAt, {
            transition: `${lastRoomUpdatedState}->${room.state}`,
          });
        }
        lastRoomUpdatedAt = now;
        lastRoomUpdatedState = room.state;

        switch (room.state) {
          case "selecting_question":
            hostSelectQuestion(socket, room);
            break;

          case "passing": {
            // Cat in the bag — instantly pass to a random connected player
            // other than the one who selected it, so the game keeps moving.
            const candidates = (room.players || []).filter(
              (p) => p.isConnected && p.id !== room.currentPlayer,
            );
            if (candidates.length > 0) {
              const target =
                candidates[Math.floor(Math.random() * candidates.length)];
              sendWS(socket, "pass_question", { passTo: target.id });
            }
            break;
          }

          case "answering":
            // Validate the answering player as correct
            sendWS(socket, "validate_answer", { isCorrect: true });
            break;

          case "betting":
            // Players place their own bets (see runAsPlayer); host has nothing to do.
            break;

          case "selecting_final_round_category": {
            const remaining = Object.entries(
              room.finalRoundState.availableCategories,
            )
              .filter(([key, isAvailable]) => isAvailable)
              .map(([key, isAvailable]) => key);
            sendWS(socket, "remove_final_round_category", {
              category: remaining[0],
            });
            break;
          }

          case "final_round_betting":
            // Let timers handle it
            break;

          case "showing_final_round_question":
            // Let players submit; host just waits
            break;

          case "validating_final_round_answers":
            if (room.currentPlayer) {
              sendWS(socket, "validate_final_round_answer", {
                isCorrect: true,
              });
            }
            break;

          case "game_over":
            socket.close();
            break;
        }
      });

      socket.on("error", (e) => {
        console.warn(`host ws error room=${roomId}: ${e}`);
      });
    },
  );

  check(res, { "host ws connected": (r) => r && r.status === 101 });
}

// ---------------------------------------------------------------------------
// Player behaviour
// ---------------------------------------------------------------------------
function runAsPlayer(roomId, playerIndex) {
  // Create guest session
  const guestRes = http.post(
    `${BASE_URL}/api/guest`,
    JSON.stringify({ name: `LT_P${__VU}_${__ITER}` }),
    { headers: { "Content-Type": "application/json" } },
  );
  check(guestRes, { "guest created": (r) => r.status === 200 });
  const cookie = extractSession(guestRes);
  if (!cookie) return;

  const userId = guestRes.json("userId");

  // Join the room
  const joinRes = http.patch(`${BASE_URL}/api/rooms/${roomId}/join`, null, {
    headers: jsonHeaders(cookie),
  });
  // 409 Conflict means already joined or room full — treat as non-fatal
  check(joinRes, {
    "player joined": (r) => r.status === 200 || r.status === 409,
  });
  if (joinRes.status !== 200) return;

  const wsUrl = `${WS_BASE}/api/ws/room/${roomId}`;
  const startMs = Date.now();

  const res = ws.connect(
    wsUrl,
    { headers: { Cookie: `sessionId=${cookie}` } },
    (socket) => {
      wsConnectTime.add(Date.now() - startMs);

      let lastRoomUpdatedAt = null;
      let lastRoomUpdatedState = null;

      socket.on("message", (data) => {
        const msg = JSON.parse(data);
        wsEventsReceived.add(1);

        if (msg.event !== "room_updated") return;

        const room = msg.payload;
        if (!room) return;

        const now = Date.now();
        if (lastRoomUpdatedAt !== null) {
          wsRoomUpdatedGap.add(now - lastRoomUpdatedAt, {
            transition: `${lastRoomUpdatedState}->${room.state}`,
          });
        }
        lastRoomUpdatedAt = now;
        lastRoomUpdatedState = room.state;

        switch (room.state) {
          case "revealing_question":
            // Submit answer if this player is allowed
            if (room.allowedToAnswer && room.allowedToAnswer.includes(userId)) {
              sendWS(socket, "start_answer");
            }
            break;

          case "betting": {
            // Auction question — bet immediately
            const me =
              room.players && room.players.find((p) => p.id === userId);
            if (me && me.betAmount == null) {
              sendWS(socket, "place_bet", { amount: Math.ceil(me.score / 2) });
            }
            break;
          }

          case "final_round_betting": {
            const me =
              room.players && room.players.find((p) => p.id === userId);
            if (me && me.betAmount == null) {
              sendWS(socket, "place_final_round_bet", {
                amount: Math.ceil(me.score / 2),
              });
            }
            break;
          }

          case "showing_final_round_question":
            if (room.finalRoundState) {
              const myAnswer =
                room.finalRoundState.playersAnswers &&
                room.finalRoundState.playersAnswers[userId];
              if (!myAnswer) {
                sendWS(socket, "submit_final_round_answer", {
                  answer: "load test answer",
                });
              }
            }
            break;

          case "game_over":
            socket.close();
            break;
        }
      });

      socket.on("error", (e) => {
        console.warn(`player ws error room=${roomId}: ${e}`);
      });
    },
  );

  check(res, { "player ws connected": (r) => r && r.status === 101 });
}

// ---------------------------------------------------------------------------
// hostSelectQuestion — picks the first unplayed question and selects it
// ---------------------------------------------------------------------------
function hostSelectQuestion(socket, room) {
  const questions = room.currentRoundQuestions;
  if (!questions) return;

  for (const cat of questions) {
    const idx = cat.questions.findIndex((q) => !q.hasBeenPlayed);
    if (idx !== -1) {
      sendWS(socket, "select_question", { category: cat.category, index: idx });
      return;
    }
  }
  // All questions in this round played — server will auto-advance to next round or final
}

// ---------------------------------------------------------------------------
// teardown — deletes host accounts (guest players expire via idle TTL)
// ---------------------------------------------------------------------------
export function teardown(rooms) {
  if (!rooms || rooms.length === 0) return;

  let deleted = 0;
  for (const room of rooms) {
    if (!room.hostUserId) continue;
    const res = http.del(`${BASE_URL}/api/users/${room.hostUserId}`, null, {
      headers: jsonHeaders(room.hostCookie),
    });
    if (res.status === 204) {
      deleted++;
    } else {
      console.warn(
        `teardown: failed to delete host ${room.hostUserId}: ${res.status}`,
      );
    }
  }
  console.log(`teardown: deleted ${deleted}/${rooms.length} host accounts`);
}
