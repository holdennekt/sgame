"use client";

import { ReactNode } from "react";
import {
  FiAlertTriangle,
  FiArrowRight,
  FiCheck,
  FiEdit3,
  FiPause,
  FiPlay,
  FiSkipForward,
  FiUserMinus,
  FiUserX,
  FiWifi,
  FiWifiOff,
  FiX,
  FiZap,
} from "react-icons/fi";
import { GameEvent } from "@/types/game_event";
import {
  formatBuzz,
  formatRelative,
  formatSigned,
  involves,
  isSystem,
} from "./helpers";
import { useLog } from "./LogContext";
import { PlayerBadge, PlayerName } from "./PlayerBadge";

type Tone = "neutral" | "success" | "danger" | "warning" | "primary";

const TONES: Record<Tone, string> = {
  neutral: "bg-surface-raised text-on-surface-muted",
  success: "bg-green-500/15 text-green-700 dark:text-green-400",
  danger: "bg-danger/15 text-danger",
  warning: "bg-amber-500/15 text-amber-700 dark:text-amber-400",
  primary: "bg-primary/15 text-primary",
};

export function Chip({
  tone = "neutral",
  children,
}: {
  tone?: Tone;
  children: ReactNode;
}) {
  return (
    <span
      className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-xs font-medium ${TONES[tone]}`}
    >
      {children}
    </span>
  );
}

function Quote({ children }: { children: ReactNode }) {
  return (
    <span className="px-2 py-0.5 rounded-md bg-surface-raised text-on-surface break-words">
      {children}
    </span>
  );
}

function Verdict({
  data,
}: {
  data: Extract<GameEvent, { type: "answer_validated" }>["data"];
}) {
  if (data.source === "moderator") return <Chip>by host</Chip>;
  if (data.source === "timeout")
    return <Chip tone="warning">time ran out</Chip>;
  if (data.source === "system") return <Chip>auto</Chip>;
  if (data.method === "error") {
    return <Chip tone="warning">AI error, counted wrong</Chip>;
  }
  const pct =
    data.confidence !== undefined
      ? ` ${Math.round(data.confidence * 100)}%`
      : "";
  return (
    <Chip tone="primary">
      {data.method === "exact" ? "AI · exact match" : `AI${pct}`}
    </Chip>
  );
}

function Validation({
  data,
  showAnswer,
}: {
  data: Extract<GameEvent, { type: "answer_validated" }>["data"];
  showAnswer: boolean;
}) {
  return (
    <>
      <PlayerBadge id={data.player} />
      {showAnswer && data.answer && <Quote>{data.answer}</Quote>}
      <Chip tone={data.correct ? "success" : "danger"}>
        {data.correct ? <FiCheck size={12} /> : <FiX size={12} />}
        {data.correct ? "Correct" : "Wrong"} {formatSigned(data.delta)}
      </Chip>
      <span className="text-xs text-on-surface-muted tabular-nums">
        score {data.score}
      </span>
      <Verdict data={data} />
    </>
  );
}

const betText = (amount: number, auto: boolean) => (
  <>
    bet <b className="tabular-nums">{amount}</b>
    {auto && <Chip>auto</Chip>}
  </>
);

function content(e: GameEvent): { body: ReactNode; slim?: ReactNode } {
  switch (e.type) {
    case "buzzed": {
      const { text, early } = formatBuzz(e.data.msFromTimerStart);
      return {
        body: (
          <>
            <FiZap className="text-amber-500" size={14} />
            <PlayerBadge id={e.data.player} />
            <span>buzzed in</span>
            <Chip tone={early ? "warning" : "neutral"}>{text}</Chip>
          </>
        ),
      };
    }
    case "false_start":
      return {
        body: (
          <>
            <FiAlertTriangle className="text-danger" size={14} />
            <PlayerBadge id={e.data.player} />
            <span>false started</span>
          </>
        ),
      };
    case "passed":
      return {
        body: (
          <>
            <PlayerBadge id={e.data.from} />
            <FiArrowRight className="text-on-surface-muted" size={14} />
            <PlayerBadge id={e.data.to} />
            {e.data.auto && <Chip>auto</Chip>}
          </>
        ),
      };
    case "bet_placed":
    case "final_bet_placed":
      return {
        body: (
          <>
            <PlayerBadge id={e.data.player} />
            <span className="inline-flex items-center gap-1.5">
              {betText(e.data.amount, e.data.auto)}
            </span>
          </>
        ),
      };
    case "auction_won":
      return {
        body: (
          <>
            <PlayerBadge id={e.data.player} />
            <span>
              won the auction at <b className="tabular-nums">{e.data.amount}</b>
            </span>
          </>
        ),
      };
    case "answer_submitted":
    case "final_answer_submitted":
      return {
        body: (
          <>
            <PlayerBadge id={e.data.player} />
            <span>answered</span>
            <Quote>{e.data.answer}</Quote>
          </>
        ),
      };
    case "answer_validated":
      return { body: <Validation data={e.data} showAnswer={false} /> };
    case "final_answer_validated":
      return { body: <Validation data={e.data} showAnswer /> };
    case "final_category_removed":
      return {
        body: (
          <>
            {e.actor && !isSystem(e.actor) && <PlayerBadge id={e.actor} />}
            <span>
              removed category <b>{e.data.category}</b>
            </span>
          </>
        ),
      };
    case "final_round_started":
      return {
        body: (
          <span>
            Final round: {e.data.players.length} player
            {e.data.players.length === 1 ? "" : "s"} qualified
          </span>
        ),
      };
    case "final_question_started":
      return {
        body: (
          <span className="inline-flex flex-wrap items-center gap-x-2 gap-y-1">
            Final question for
            {e.data.players.map((id) => (
              <PlayerBadge key={id} id={id} />
            ))}
          </span>
        ),
      };
    case "final_category_chosen":
      return {
        body: (
          <span>
            Category <b>{e.data.category}</b> chosen
          </span>
        ),
      };
    case "paused":
      return {
        slim: (
          <>
            <FiPause size={12} /> <PlayerName id={e.actor} /> paused the game
          </>
        ),
        body: null,
      };
    case "unpaused":
      return {
        slim: (
          <>
            <FiPlay size={12} />{" "}
            {e.data.auto ? (
              "Game was unpaused automatically"
            ) : (
              <>
                <PlayerName id={e.actor} /> unpaused the game
              </>
            )}
          </>
        ),
        body: null,
      };
    case "score_changed":
      return {
        slim: (
          <>
            <FiEdit3 size={12} /> <PlayerName id={e.actor} /> changed{" "}
            <PlayerName id={e.data.player} /> score {e.data.from} → {e.data.to}
          </>
        ),
        body: null,
      };
    case "player_banned":
      return {
        slim: (
          <>
            <FiUserX size={12} /> <PlayerName id={e.actor} /> banned{" "}
            <b>{e.data.name || "a player"}</b>
          </>
        ),
        body: null,
      };
    case "round_skipped":
      return {
        slim: (
          <>
            <FiSkipForward size={12} /> <PlayerName id={e.actor} /> skipped the
            round
          </>
        ),
        body: null,
      };
    case "player_disconnected":
      return {
        slim: (
          <>
            <FiWifiOff size={12} /> <PlayerName id={e.data.player} />{" "}
            disconnected
          </>
        ),
        body: null,
      };
    case "player_reconnected":
      return {
        slim: (
          <>
            <FiWifi size={12} /> <PlayerName id={e.data.player} /> reconnected
          </>
        ),
        body: null,
      };
    case "player_left":
      return {
        slim: (
          <>
            <FiUserMinus size={12} /> <PlayerName id={e.data.player} /> left
          </>
        ),
        body: null,
      };
    case "log_truncated":
      return {
        slim: (
          <>The log reached its size limit; later events are not recorded</>
        ),
        body: null,
      };
    default:
      return { body: null };
  }
}

export default function EventRow({ event }: { event: GameEvent }) {
  const { start, filter } = useLog();
  const { body, slim } = content(event);
  if (!body && !slim) return null;

  const dimmed = filter !== null && !involves(event, filter);
  return (
    <div
      className={`flex gap-3 items-start transition-opacity duration-150 ${
        slim ? "py-0.5" : "py-1"
      } ${dimmed ? "opacity-25" : ""}`}
    >
      <span className="w-11 shrink-0 pt-0.5 text-[11px] tabular-nums text-on-surface-muted">
        {formatRelative(event.at, start)}
      </span>
      {slim ? (
        <div className="min-w-0 flex-1 flex flex-wrap items-center gap-x-1.5 gap-y-0.5 text-xs text-on-surface-muted">
          {slim}
        </div>
      ) : (
        <div className="min-w-0 flex-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-on-surface">
          {body}
        </div>
      )}
    </div>
  );
}
