import {
  PackPreview,
  HiddenQuestion,
  HiddenFinalRoundQuestion,
  FinalRoundQuestion,
  Question,
  QuestionType,
  PrivacyType,
  Comment,
  Attachment,
} from "./pack";
import { Host, Player } from "./user";

export interface GameHistoryEntry {
  id: string;
  name: string;
  packPreview: PackPreview;
  players: Player[];
  finishedAt: string;
}

export interface CreateRoomRequest {
  name: string;
  packId: string;
  options: RoomOptions;
}

export interface CreateRoomResponse {
  id: string;
}

export interface Room {
  id: string;
  name: string;
  createdBy: string;
  options: RoomOptions;
  packPreview: PackPreview;
  host: Host | null;
  players: Player[];
  banList: string[];
  state: RoomState;
  currentRoundName: string | null;
  currentRoundQuestions: CurrentRoundQuestions | null;
  currentPlayer: string | null;
  currentQuestion: CurrentQuestion | null;
  answeringPlayer: AnsweringPlayer | null;
  allowedToAnswer: string[];
  finalRoundState: FinalRoundState | null;
  pausedState: PausedState;
}

const dummyRoom: Room = {
  id: "",
  name: "",
  createdBy: "",
  options: {
    maxPlayers: 0,
    type: "public",
    password: null,
    questionThinkingTime: 10,
    answerThinkingTime: 5,
    questionThinkingTimeFinal: 60,
    readingSymbolsPerSecond: 30,
    falseStartAllowed: true,
  },
  packPreview: { id: "", name: "" },
  host: null,
  players: [],
  banList: [],
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
    bettingEndsAt: null,
    timerEndsAt: null,
  },
  pausedState: {
    paused: false,
    pausedAt: null,
  },
};

export const isRoom = (obj: unknown): obj is Room => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyRoom).every((key) => Object.hasOwn(obj, key));
};

export type RoomState =
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

export interface RoomOptions {
  maxPlayers: number;
  type: PrivacyType;
  password: string | null;
  questionThinkingTime: number;
  answerThinkingTime: number;
  questionThinkingTimeFinal: number;
  readingSymbolsPerSecond: number;
  falseStartAllowed: boolean;
}

export interface BoardQuestion {
  index: number;
  value: number;
  hasBeenPlayed: boolean;
}

export interface CategoryQuestions {
  category: string;
  questions: BoardQuestion[];
}
export type CurrentRoundQuestions = CategoryQuestions[];

export interface CurrentQuestion extends Question {
  attachmentRevealEndsAt: string;
  attachmentRevealLastProgress: number;
  textRevealLastProgress: number;
  timerStartsAt: string;
  timerEndsAt: string;
  timerLastProgress: number;
  bettingEndsAt: string;
  passingEndsAt: string;
}

export interface AnsweringPlayer {
  id: string;
  timerStartsAt: string;
  timerEndsAt: string;
}

export interface FinalRoundState {
  availableCategories: Record<string, boolean>;
  question: FinalRoundQuestion | null;
  players: string[];
  playersAnswers: Record<string, string>;
  bettingEndsAt: string | null;
  timerEndsAt: string | null;
}

export interface PausedState {
  paused: boolean;
  pausedAt: string | null;
}

export interface RoomLobby {
  id: string;
  name: string;
  packPreview: PackPreview;
  host: Host | null;
  players: Player[];
  maxPlayers: number;
  type: PrivacyType;
  status: string;
}

const dummyRoomLobby: RoomLobby = {
  id: "1",
  name: "xyz",
  packPreview: { id: "1", name: "wtf" },
  players: [],
  maxPlayers: 3,
  type: "public",
  status: "Idle",
  host: null,
};

export const isRoomLobby = (obj: unknown): obj is RoomLobby => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyRoomLobby).every((key) => Object.hasOwn(obj, key));
};

export interface RoomHost {
  id: string;
  name: string;
  packPreview: PackPreview;
  host: Host | null;
  players: Player[];
  state: RoomState;
  currentRoundName: string | null;
  currentRoundQuestions: CurrentRoundQuestions | null;
  currentPlayer: string | null;
  currentQuestion: CurrentQuestion | null;
  answeringPlayer: AnsweringPlayer | null;
  allowedToAnswer: string[];
  finalRoundState: FinalRoundState | null;
  pausedState: PausedState;
}

const dummyRoomHost: RoomHost = {
  id: "",
  name: "",
  packPreview: { id: "", name: "" },
  host: null,
  players: [],
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
    bettingEndsAt: null,
    timerEndsAt: null,
  },
  pausedState: {
    paused: false,
    pausedAt: null,
  },
};

export const isRoomHost = (obj: unknown): obj is RoomHost => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyRoomHost).every((key) => Object.hasOwn(obj, key));
};

export interface RoomPlayer {
  id: string;
  name: string;
  packPreview: PackPreview;
  host: Host | null;
  players: Player[];
  state: RoomState;
  currentRoundName: string | null;
  currentRoundQuestions: CurrentRoundQuestions | null;
  currentPlayer: string | null;
  currentQuestion: HiddenCurrentQuestion | null;
  answeringPlayer: AnsweringPlayer | null;
  allowedToAnswer: string[];
  finalRoundState: HiddenFinalRoundState | null;
  pausedState: PausedState;
}

const dummyRoomPlayer: RoomPlayer = {
  id: "",
  name: "",
  packPreview: { id: "", name: "" },
  host: null,
  players: [],
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
    bettingEndsAt: null,
    timerEndsAt: null,
  },
  pausedState: {
    paused: false,
    pausedAt: null,
  },
};

export const isRoomPlayer = (obj: unknown): obj is RoomPlayer => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyRoomPlayer).every((key) => Object.hasOwn(obj, key));
};

export interface HiddenCurrentQuestion extends HiddenQuestion {
  type: QuestionType;
  text: string | null;
  attachment: Attachment | null;
  attachmentRevealEndsAt: string;
  attachmentRevealLastProgress: number;
  textRevealLastProgress: number;
  timerStartsAt: string;
  timerEndsAt: string;
  timerLastProgress: number;
  bettingEndsAt: string;
  passingEndsAt: string;
}

export interface HiddenFinalRoundState {
  availableCategories: Record<string, boolean>;
  question: HiddenFinalRoundQuestion | null;
  players: string[];
  playersAnswers: Record<string, boolean>;
  bettingEndsAt: string | null;
  timerEndsAt: string | null;
}

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
  return Object.keys(dummyRoundDemo).every((key) => Object.hasOwn(obj, key));
};

export type QuestionDemo = {
  category: string;
  categoryComment: string | null;
  value: number;
  type: QuestionType;
  duration: number;
};

const dummyQuestionDemo: QuestionDemo = {
  category: "",
  categoryComment: null,
  value: 0,
  type: "regular",
  duration: 0,
};

export const isQuestionDemo = (obj: unknown): obj is QuestionDemo => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyQuestionDemo).every((key) => Object.hasOwn(obj, key));
};

export type CorrectAnswerDemo = {
  answers: string[];
  comment: Comment | null;
  duration: number;
};

const dummyCorrectAnswerDemo: CorrectAnswerDemo = {
  answers: [],
  comment: null,
  duration: 0,
};

export const isCorrectAnswerDemo = (obj: unknown): obj is CorrectAnswerDemo => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyCorrectAnswerDemo).every((key) =>
    Object.hasOwn(obj, key)
  );
};
