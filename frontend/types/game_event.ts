import { QuestionType } from "./pack";

export const SYSTEM_ACTOR = "SYSTEM";

export type AttachmentKind = "image" | "audio" | "video";
export type ValidationSource = "moderator" | "ai" | "timeout" | "system";
export type VerdictMethod = "exact" | "ai" | "error";
export type QuestionEndReason =
  "answered" | "nobody_answered" | "no_bets" | "skipped";

interface BaseEvent<T extends string, D> {
  seq: number;
  at: string;
  type: T;
  actor?: string;
  data: D;
}

interface Validation {
  player: string;
  answer: string;
  correct: boolean;
  delta: number;
  score: number;
  source: ValidationSource;
  method?: VerdictMethod;
  confidence?: number;
}

export type GameEvent =
  | BaseEvent<
      "player_joined",
      {
        player: string;
        name: string;
        avatar: string | null;
        isGuest: boolean;
        role: "player" | "moderator";
      }
    >
  | BaseEvent<"player_left", { player: string }>
  | BaseEvent<"player_disconnected", { player: string }>
  | BaseEvent<"player_reconnected", { player: string }>
  | BaseEvent<
      "game_started",
      {
        pack: string;
        players: string[];
        aiHost: boolean;
        falseStart: boolean;
      }
    >
  | BaseEvent<"round_started", { round: string; index: number }>
  | BaseEvent<"round_skipped", { round: string | null }>
  | BaseEvent<
      "question_selected",
      {
        round: string;
        category: string;
        value: number;
        type: QuestionType;
        text: string | null;
        attachmentType: AttachmentKind | null;
      }
    >
  | BaseEvent<"passed", { from: string; to: string; auto: boolean }>
  | BaseEvent<"bet_placed", { player: string; amount: number; auto: boolean }>
  | BaseEvent<"auction_won", { player: string; amount: number }>
  | BaseEvent<
      "buzzed",
      {
        player: string;
        phase: "revealing" | "showing";
        msFromTimerStart: number;
      }
    >
  | BaseEvent<"false_start", { player: string }>
  | BaseEvent<"answer_submitted", { player: string; answer: string }>
  | BaseEvent<"answer_validated", Validation>
  | BaseEvent<
      "question_ended",
      { reason: QuestionEndReason; answers?: string[]; winner?: string }
    >
  | BaseEvent<"final_round_started", { players: string[] }>
  | BaseEvent<"final_category_removed", { category: string }>
  | BaseEvent<
      "final_category_chosen",
      {
        category: string;
        text: string | null;
        attachmentType: AttachmentKind | null;
        answers?: string[];
      }
    >
  | BaseEvent<
      "final_bet_placed",
      { player: string; amount: number; auto: boolean }
    >
  | BaseEvent<"final_question_started", { players: string[] }>
  | BaseEvent<"final_answer_submitted", { player: string; answer: string }>
  | BaseEvent<"final_answer_validated", Validation>
  | BaseEvent<"score_changed", { player: string; from: number; to: number }>
  | BaseEvent<"player_banned", { player: string; name: string }>
  | BaseEvent<"paused", Record<string, never>>
  | BaseEvent<"unpaused", { auto: boolean }>
  | BaseEvent<"game_ended", { standings: { player: string; score: number }[] }>
  | BaseEvent<"log_truncated", Record<string, never>>;

export type GameEventType = GameEvent["type"];

export interface GameEventsResponse {
  events: GameEvent[];
}
