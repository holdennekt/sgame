import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { ChatMessage, isChatMessage } from "@/components/Message";
import { isError } from "@/middleware";
import {
  isRoomModerator,
  isRoomPlayer,
  RoomModerator,
  RoomPlayer,
  RoundDemo,
  QuestionDemo,
  CorrectAnswerDemo,
  isRoundDemo,
  isQuestionDemo,
  isCorrectAnswerDemo,
} from "@/types/room";
import { joinRoom, leaveRoom } from "@/app/api";
import { useWebSocket } from "./useWebSocket";

export function useRoom(
  initialRoom: RoomModerator | RoomPlayer,
  userId: string,
  isSpectator = false,
  password?: string
) {
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

  const wsPath = (() => {
    const base = `/api/ws/room/${initialRoom.id}`;
    if (isSpectator && password) {
      return `${base}?password=${encodeURIComponent(password)}`;
    }
    return base;
  })();

  const { wsConn, handlers } = useWebSocket(wsPath, "room");

  const setError = (msg: string) =>
    setLastError((prev) => ({ msg, count: (prev?.count ?? 0) + 1 }));

  const send = (event: string, payload?: unknown) => {
    if (isSpectator) return;
    wsConn.current?.send(JSON.stringify({ event, payload }));
  };

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
    if (!isRoomModerator(payload) && !isRoomPlayer(payload)) return;
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
  const startAnswer = () => send("start_answer");
  const submitTypedAnswer = (answer: string) =>
    send("submit_answer", { answer });
  const banPlayer = (playerId: string) => send("ban_player", { playerId });
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
  const skipRound = () => send("skip_round");
  const changeScore = (playerId: string, score: number) =>
    send("change_score", { playerId, score });

  const leave = async () => {
    if (isSpectator) {
      router.push("/");
      return;
    }
    try {
      await leaveRoom(room.id);
      router.push("/");
    } catch (e) {
      if (isError(e)) setError(e.error);
    }
  };

  const joinAsPlayer = async () => {
    try {
      await joinRoom(room.id, undefined);
      router.push(`/rooms/${room.id}`);
    } catch (e) {
      if (isError(e)) setError(e.error);
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
      joinAsPlayer,
      startAnswer,
      submitTypedAnswer,
      banPlayer,
      selectQuestion,
      passQuestion,
      placeBet,
      placeFinalRoundBet,
      submitFinalRoundAnswer,
      removeFinalRoundCategory,
      validateAnswer,
      validateFinalRoundAnswer,
      skipQuestion,
      skipRound,
      changeScore,
    },
  };
}
