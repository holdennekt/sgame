"use client";

import { createContext, useContext } from "react";
import { LogPlayer } from "./helpers";

export interface LogContextValue {
  players: Map<string, LogPlayer>;
  start: number;
  filter: string | null;
}

export const LogContext = createContext<LogContextValue>({
  players: new Map(),
  start: 0,
  filter: null,
});

export const useLog = () => useContext(LogContext);
