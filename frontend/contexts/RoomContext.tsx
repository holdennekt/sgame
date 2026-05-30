"use client";

import { createContext, RefObject, useContext } from "react";
import React from "react";
import {
  RoomHost,
  RoomPlayer,
  RoundDemo,
  QuestionDemo,
  CorrectAnswerDemo,
} from "@/types/room";

export type RoomContextValue = {
  room: RoomHost | RoomPlayer;
  roundDemo: RoundDemo | null;
  questionDemo: QuestionDemo | null;
  correctAnswerDemo: CorrectAnswerDemo | null;
  onRoundDemoDone: () => void;
  answerButton: RefObject<HTMLDivElement>;
  submitAnswer: () => void;
  selectQuestion: (q: { category: string; index: number }) => void;
  passQuestion: (passTo: string) => void;
  placeBet: (amount: number) => void;
  placeFinalRoundBet: (amount: number) => void;
  submitFinalRoundAnswer: (answer: string) => void;
  removeFinalRoundCategory: (category: string) => void;
  validateAnswer: (isCorrect: boolean) => void;
  validateFinalRoundAnswer: (isCorrect: boolean) => void;
  skipQuestion: () => void;
  changeScore: (playerId: string, score: number) => void;
};

const RoomContext = createContext<RoomContextValue | null>(null);

export function RoomProvider({
  children,
  value,
}: {
  children: React.ReactNode;
  value: RoomContextValue;
}) {
  return <RoomContext.Provider value={value}>{children}</RoomContext.Provider>;
}

export function useRoomContext() {
  const ctx = useContext(RoomContext);
  if (!ctx) throw new Error("useRoomContext must be used within RoomProvider");
  return ctx;
}
