import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { ChatMessage, isChatMessage } from "@/components/Message";
import { isError } from "@/middleware";
import {
  isRoomHost,
  isRoomPlayer,
  RoomHost,
  RoomPlayer,
  RoundDemo,
  QuestionDemo,
  CorrectAnswerDemo,
  isRoundDemo,
  isQuestionDemo,
  isCorrectAnswerDemo,
} from "@/types/room";
import { leaveRoom } from "@/app/actions";
import { useWebSocket } from "./useWebSocket";

export function useRoom(initialRoom: RoomHost | RoomPlayer, userId: string) {
  const router = useRouter();
  const [room, setRoom] = useState(initialRoom);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [lastError, setLastError] = useState<{
    msg: string;
    count: number;
  } | null>(null);
  const [roundDemo, setRoundDemo] = useState<RoundDemo | null>(null);
  const [questionDemo, setQuestionDemo] = useState<QuestionDemo | null>(null);
  const [correctAnswerDemo, setCorrectAnswerDemo] =
    useState<CorrectAnswerDemo | null>(null);
  const delayedRoundDemoRef = useRef<RoundDemo | null>(null);

  const { wsConn, handlers } = useWebSocket(
    `/api/ws/room/${initialRoom.id}`,
    "room"
  );

  const setError = (msg: string) =>
    setLastError((prev) => ({ msg, count: (prev?.count ?? 0) + 1 }));

  const send = (event: string, payload?: unknown) =>
    wsConn.current?.send(JSON.stringify({ event, payload }));

  useEffect(() => {
    if (!questionDemo && !correctAnswerDemo && delayedRoundDemoRef.current) {
      setRoundDemo(delayedRoundDemoRef.current);
      delayedRoundDemoRef.current = null;
    }
  }, [questionDemo, correctAnswerDemo]);

  handlers.set("error", (payload) => {
    if (!isError(payload)) return;
    setError(payload.error);
  });
  handlers.set("chat", (payload) => {
    if (!isChatMessage(payload)) return;
    setChatMessages((msgs) => [...msgs, payload]);
  });
  handlers.set("room_updated", (payload) => {
    if (!isRoomHost(payload) && !isRoomPlayer(payload)) return;
    setRoom(payload);
  });
  handlers.set("room_deleted", () => router.push("/"));
  handlers.set("round_demo", (payload) => {
    if (!isRoundDemo(payload)) return;
    if (correctAnswerDemo) {
      delayedRoundDemoRef.current = payload;
      return;
    }
    setRoundDemo(payload);
  });
  handlers.set("question_demo", (payload) => {
    if (!isQuestionDemo(payload)) return;
    setQuestionDemo(payload);
    setTimeout(() => setQuestionDemo(null), payload.duration * 1000);
  });
  handlers.set("correct_answer_demo", (payload) => {
    if (!isCorrectAnswerDemo(payload)) return;
    setCorrectAnswerDemo(payload);
    setTimeout(() => setCorrectAnswerDemo(null), payload.duration * 1000);
  });

  const sendChatMessage = (text: string) => send("chat", { text });
  const startGame = () => send("start_game");
  const togglePause = () => send(room.pausedState.paused ? "unpause" : "pause");
  const selectQuestion = (q: { category: string; index: number }) =>
    send("select_question", q);
  const submitAnswer = () => send("submit_answer");
  const passQuestion = (passTo: string) => send("pass_question", { passTo });
  const placeBet = (amount: number) => send("place_bet", { amount });
  const placeFinalRoundBet = (amount: number) =>
    send("place_final_round_bet", { amount });
  const submitFinalRoundAnswer = (answer: string) =>
    send("submit_final_round_answer", { answer });
  const removeFinalRoundCategory = (category: string) =>
    send("remove_final_round_category", { category });
  const validateAnswer = (isCorrect: boolean) =>
    send("validate_answer", { isCorrect });
  const validateFinalRoundAnswer = (isCorrect: boolean) =>
    send("validate_final_round_answer", { isCorrect });
  const skipQuestion = () => send("skip_question");
  const changeScore = (playerId: string, score: number) =>
    send("change_score", { playerId, score });

  const leave = async () => {
    try {
      await leaveRoom(room.id);
      router.push("/");
    } catch (error) {
      if (error instanceof Error) setError(error.message);
    }
  };

  return {
    lastError,
    chat: {
      messages: chatMessages,
      send: sendChatMessage,
    },
    game: {
      room,
      roundDemo,
      questionDemo,
      correctAnswerDemo,
      onRoundDemoDone: () => setRoundDemo(null),
      startGame,
      togglePause,
      leave,
      submitAnswer,
      selectQuestion,
      passQuestion,
      placeBet,
      placeFinalRoundBet,
      submitFinalRoundAnswer,
      removeFinalRoundCategory,
      validateAnswer,
      validateFinalRoundAnswer,
      skipQuestion,
      changeScore,
    },
  };
}
