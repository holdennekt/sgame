"use client";

import { useEffect, useRef, useState } from "react";
import { isError, User } from "../../../middleware";
import { WsMessage, WsMessageHandler } from "../lobby/Lobby";
import { toast, ToastContainer } from "react-toastify";
import Chat from "../Chat";
import { ChatMessage, isChatMessage } from "../Message";
import { getAvatar } from "../lobby/Room";
import Link from "next/link";
import ControlButtons from "./ControlButtons";
import { useRouter } from "next/navigation";
import { PackPreview } from "../lobby/NewRoomModal";
import { leaveRoom } from "@/app/actions";
import { Answer, Question, QuestionType } from "../pack/PackEditor";
import { Player } from "./mainSection/PlayerTable";
import MainPanel from "./mainSection/MainPanel";

export type Host = User & { isConnected: boolean };
export type BoardQuestion = {
  index: number;
  value: number;
  hasBeenPlayed: boolean;
};

type RoomState =
  | "waiting_for_start"
  | "selecting_question"
  | "revealing_question"
  | "showing_question"
  | "answering"
  | "betting"
  | "passing"
  | "selecting_final_round_category"
  | "final_round_betting"
  | "showing_final_round_question"
  | "validating_final_round_answers"
  | "game_over";

export type CurrentQuestion = Question & {
  timerLastProgress: number;
  timerStartsAt: Date;
  timerEndsAt: Date;
};
export type HiddenCurrentQuestion = HiddenQuestion & {
  type: QuestionType;
  text: string;
  timerLastProgress: number;
  timerStartsAt: Date;
  timerEndsAt: Date;
};

export type Room = {
  id: string;
  name: string;
  packPreview: PackPreview;
  host: Host | null;
  players: Player[];
  state: RoomState;
  currentRoundName: string | null;
  currentRoundQuestions: { [key: string]: BoardQuestion[] } | null;
  currentPlayer: string | null;
  currentQuestion: CurrentQuestion | HiddenCurrentQuestion | null;
  answeringPlayer: {
    id: string;
    timerStartsAt: Date;
    timerEndsAt: Date;
  } | null;
  allowedToAnswer: string[];
  finalRoundState: FinalRoundState | HiddenFinalRoundState | null;
  pausedState: {
    paused: boolean;
    pausedAt: Date | null;
  };
};

const dummyRoom: Room = {
  id: "",
  name: "",
  packPreview: { id: "", name: "" },
  players: [],
  host: null,
  state: "waiting_for_start",
  currentRoundName: null,
  currentRoundQuestions: null,
  currentPlayer: null,
  currentQuestion: null,
  answeringPlayer: null,
  allowedToAnswer: [],
  finalRoundState: {
    availableCategories: {},
    question: null,
    players: [],
    playersAnswers: {},
    timerEndsAt: null,
  },
  pausedState: {
    paused: false,
    pausedAt: null,
  },
};

export const isRoom = (obj: unknown): obj is Room => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyRoom).every(key => Object.hasOwn(obj, key));
};

export type HiddenQuestion = {
  index: number;
  value: number;
  text: string;
  attachment: {
    mediaType: "image" | "audio" | "video";
    contentUrl: string;
  } | null;
};

export type FinalRoundState = {
  availableCategories: { [key: string]: boolean } | null;
  question: FinalRoundQuestion | null;
  players: string[];
  playersAnswers: { [key: string]: string } | null;
  timerEndsAt: Date | null;
};

export type HiddenFinalRoundState = {
  availableCategories: { [key: string]: boolean } | null;
  question: HiddenFinalRoundQuestion | null;
  players: string[];
  playersAnswers: { [key: string]: boolean } | null;
  timerEndsAt: Date | null;
};

type HiddenFinalRoundQuestion = {
  text: string;
  attachment: {
    mediaType: "image" | "audio" | "video";
    contentUrl: string;
  } | null;
};

type FinalRoundQuestion = HiddenFinalRoundQuestion & Answer;

export type RoundDemo = {
  name: string;
  categories: string[];
};

const dummyRoundDemo: RoundDemo = {
  name: "round 1",
  categories: ["1", "2", "3"],
};

export const isRoundDemo = (obj: unknown): obj is RoundDemo => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyRoundDemo).every(key => Object.hasOwn(obj, key));
};

export type CorrectAnswerDemo = {
  answers: string[];
  comment: string | null;
  duration: number;
};

const dummyCorrectAnswerDemo: CorrectAnswerDemo = {
  answers: ["some answer", "another one"],
  comment: "here is further explanation of answer",
  duration: 5,
};

export const isCorrectAnswerDemo = (obj: unknown): obj is CorrectAnswerDemo => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyCorrectAnswerDemo).every(key =>
    Object.hasOwn(obj, key)
  );
};

export default function RoomPage({
  user,
  initialRoom,
}: {
  user: User;
  initialRoom: Room;
}) {
  const router = useRouter();
  const [room, setRoom] = useState(initialRoom);

  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);

  const isHost = user.id === room.host?.id;

  const wsConn = useRef<WebSocket | null>(null);
  const mainContainer = useRef<HTMLDivElement>(null);
  const answerButton = useRef<HTMLDivElement>(null);

  const handlers = new Map<string, WsMessageHandler>();
  handlers.set("error", payload => {
    if (!isError(payload)) return;
    toast.error(payload.error, { containerId: "room" });
    mainContainer.current?.focus();
  });
  handlers.set("chat", payload => {
    if (!isChatMessage(payload)) return;
    setChatMessages(chatMessages => [...chatMessages, payload]);
  });
  handlers.set("room_updated", payload => {
    if (!isRoom(payload)) return;
    setRoom(payload);
  });
  handlers.set("room_deleted", () => router.push("/"));

  useEffect(() => {
    let isMounted = true;

    mainContainer.current?.focus();

    const connectWebsocket = () => {
      wsConn.current = new WebSocket(
        `ws://${window.location.host}/api/ws/room/${room.id}`
      );

      wsConn.current.onmessage = (ev: MessageEvent<string>) => {
        const message: WsMessage = JSON.parse(ev.data, (key, val) =>
          (key.endsWith("At") && val ? new Date(val as string) : val)
        );
        console.log("incoming message", message);
        const handler = handlers.get(message.event);
        if (handler) handler(message.payload);
      };

      wsConn.current.onclose = () => {
        toast.error("Disconnected from server. Trying to recoonect in 3s", {
          containerId: "room",
        });

        if (isMounted) setTimeout(connectWebsocket, 3000);
      };
    };

    connectWebsocket();

    return () => {
      isMounted = false;
      wsConn.current?.close();
    };
  }, [room.id]);

  const sendChatMessage = (text: string) => {
    wsConn.current?.send(JSON.stringify({ event: "chat", payload: { text } }));
  };

  const startGame = () => {
    if (!isHost) return;
    wsConn.current?.send(JSON.stringify({ event: "start_game" }));
  };

  const togglePause = () => {
    if (!isHost) return;
    wsConn.current?.send(JSON.stringify({ event: "toggle_pause" }));
  };

  const leave = async () => {
    try {
      await leaveRoom(room.id);
      router.push("/");
    } catch (error) {
      if (error instanceof Error) {
        toast.error(error.message, { containerId: "room" });
        mainContainer.current?.focus();
      }
    }
  };

  const submitAnswer = () => {
    answerButton.current?.blur();
    mainContainer.current?.focus();
    if (
      isHost ||
      (room.state !== "revealing_question" && room.state !== "showing_question")
    )
      return;
    wsConn.current?.send(JSON.stringify({ event: "submit_answer" }));
  };

  return (
    <>
      <main
        className="flex flex-col-reverse md:flex-row gap-2 flex-1 min-w-0 min-h-0 p-2 focus:outline-none"
        tabIndex={-1}
        ref={mainContainer}
        onKeyDown={e => {
          if (e.code !== "Space") return;
          answerButton.current?.focus();
        }}
        onKeyUp={e => {
          if (e.code !== "Space") return;
          submitAnswer();
        }}
      >
        <MainPanel
          user={user}
          room={room}
          wsConn={wsConn}
          handlers={handlers}
          answerButton={answerButton}
          submitAnswer={submitAnswer}
        />
        <div className="flex-1 flex flex-col gap-2">
          <div className="rounded surface">
            <div className="w-full flex p-2">
              <div className="flex-[1_0_auto]">
                <p className="text-lg font-semibold">{room.name}</p>
                <p className="text-sm font-normal">
                  Pack:{" "}
                  <Link
                    className="pack-link"
                    href={`/packs/${room.packPreview.id}`}
                    target="_blank"
                  >
                    {room.packPreview.name}
                  </Link>
                </p>
              </div>
              <div
                className={`flex-auto max-w-20 flex flex-col justify-center items-center${
                  room.host?.isConnected ? "" : " opacity-50"
                }`}
              >
                <div className="w-10">{getAvatar(room.host)}</div>
                <p
                  className="w-full text-center text-sm truncate text-white"
                  title={room.host?.name}
                >
                  {room.host?.name}
                </p>
              </div>
            </div>
            <ControlButtons
              isHost={isHost}
              isGameStarted={room.state !== "waiting_for_start"}
              isPaused={room.pausedState.paused}
              start={startGame}
              togglePause={togglePause}
              leave={leave}
            />
          </div>
          <div className="flex min-w-0 min-h-0 flex-1">
            <Chat
              user={user}
              messages={chatMessages}
              sendMessage={sendChatMessage}
            />
          </div>
        </div>
      </main>
      <ToastContainer
        containerId="room"
        position="bottom-left"
        theme="colored"
      />
    </>
  );
}
