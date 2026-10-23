import { GameEvent, SYSTEM_ACTOR } from "@/types/game_event";

export interface LogPlayer {
  id: string;
  name: string;
  avatar: string | null;
  isGuest: boolean;
  role: "player" | "moderator";
  color: string;
}

export type QuestionSelected = Extract<
  GameEvent,
  { type: "question_selected" }
>;
export type QuestionEnded = Extract<GameEvent, { type: "question_ended" }>;

export interface QuestionBlock {
  selected: QuestionSelected;
  items: GameEvent[];
  ended: QuestionEnded | null;
}

export type RoundEntry =
  | { kind: "question"; question: QuestionBlock }
  | { kind: "note"; event: GameEvent };

export interface RoundBlock {
  name: string;
  at: string;
  entries: RoundEntry[];
}

export interface Timeline {
  start: number;
  end: number;
  pack: string | null;
  standings: { player: string; score: number }[];
  rounds: RoundBlock[];
  final: GameEvent[];
  questionCount: number;
}

const COLORS = [
  "#6366f1",
  "#14b8a6",
  "#f59e0b",
  "#ec4899",
  "#22c55e",
  "#06b6d4",
  "#a855f7",
  "#ef4444",
  "#84cc16",
  "#f97316",
];
const MODERATOR_COLOR = "#71717a";

export const buildPlayers = (events: GameEvent[]): Map<string, LogPlayer> => {
  const players = new Map<string, LogPlayer>();
  let colorIndex = 0;
  for (const e of events) {
    if (e.type === "player_joined") {
      const known = players.get(e.data.player);
      const isModerator = e.data.role === "moderator";
      players.set(e.data.player, {
        id: e.data.player,
        name: e.data.name,
        avatar: e.data.avatar,
        isGuest: e.data.isGuest,
        role: known?.role === "player" ? "player" : e.data.role,
        color:
          known?.color ??
          (isModerator
            ? MODERATOR_COLOR
            : COLORS[colorIndex++ % COLORS.length]),
      });
    }
  }
  return players;
};

const FINAL_TYPES = new Set<GameEvent["type"]>([
  "final_round_started",
  "final_category_removed",
  "final_category_chosen",
  "final_bet_placed",
  "final_question_started",
  "final_answer_submitted",
  "final_answer_validated",
]);

const LOBBY_TYPES = new Set<GameEvent["type"]>([
  "player_joined",
  "player_left",
  "game_started",
]);

export const buildTimeline = (events: GameEvent[]): Timeline => {
  const timeline: Timeline = {
    start: events.length ? Date.parse(events[0].at) : 0,
    end: events.length ? Date.parse(events[events.length - 1].at) : 0,
    pack: null,
    standings: [],
    rounds: [],
    final: [],
    questionCount: 0,
  };
  let round: RoundBlock | null = null;
  let question: QuestionBlock | null = null;
  let inFinal = false;

  for (const e of events) {
    if (e.type === "game_started") timeline.pack = e.data.pack;
    if (e.type === "game_ended") timeline.standings = e.data.standings;
    if (LOBBY_TYPES.has(e.type) || e.type === "game_ended") continue;

    if (e.type === "round_started") {
      round = { name: e.data.round, at: e.at, entries: [] };
      timeline.rounds.push(round);
      continue;
    }
    if (FINAL_TYPES.has(e.type)) inFinal = true;

    if (inFinal) {
      timeline.final.push(e);
      continue;
    }
    if (e.type === "question_selected") {
      if (!round) {
        round = { name: e.data.round, at: e.at, entries: [] };
        timeline.rounds.push(round);
      }
      question = { selected: e, items: [], ended: null };
      round.entries.push({ kind: "question", question });
      timeline.questionCount++;
      continue;
    }
    if (e.type === "question_ended" && question) {
      question.ended = e;
      question = null;
      continue;
    }
    if (question) {
      question.items.push(e);
    } else if (round) {
      round.entries.push({ kind: "note", event: e });
    }
  }
  return timeline;
};

export const involves = (e: GameEvent, playerId: string): boolean => {
  if (e.actor === playerId) return true;
  const d = e.data as Record<string, unknown>;
  return ["player", "from", "to", "winner"].some((k) => d[k] === playerId);
};

export const isSystem = (actor?: string) => actor === SYSTEM_ACTOR;

export const formatRelative = (at: string, start: number): string => {
  const sec = Math.max(0, Math.round((Date.parse(at) - start) / 1000));
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `+${m}:${s.toString().padStart(2, "0")}`;
};

export const formatDuration = (ms: number): string => {
  const sec = Math.max(0, Math.round(ms / 1000));
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  const pad = (n: number) => n.toString().padStart(2, "0");
  return h ? `${h}:${pad(m)}:${pad(s)}` : `${m}:${pad(s)}`;
};

export const formatSigned = (n: number): string =>
  n >= 0 ? `+${n}` : `−${Math.abs(n)}`;

export const formatBuzz = (ms: number): { text: string; early: boolean } => {
  const sec = (Math.abs(ms) / 1000).toFixed(1);
  return ms < 0
    ? { text: `${sec}s early`, early: true }
    : { text: `${sec}s`, early: false };
};
