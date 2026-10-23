"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { FiArrowLeft, FiChevronDown } from "react-icons/fi";
import { getGameEvents } from "@/app/api";
import { GameEvent } from "@/types/game_event";
import EventRow from "./EventRow";
import FinalRound from "./FinalRound";
import { buildPlayers, buildTimeline, formatDuration } from "./helpers";
import { LogContext, useLog } from "./LogContext";
import { PlayerAvatar, PlayerName } from "./PlayerBadge";
import QuestionCard from "./QuestionCard";
import ScoreChart from "./ScoreChart";

function Message({ title, text }: { title: string; text: string }) {
  return (
    <div className="flex flex-col items-center gap-1 py-16 text-center">
      <h2 className="text-base font-semibold text-on-surface">{title}</h2>
      <p className="text-sm text-on-surface-muted">{text}</p>
    </div>
  );
}

function Standings({
  standings,
}: {
  standings: { player: string; score: number }[];
}) {
  const { players } = useLog();
  return (
    <div className="flex flex-col gap-1.5">
      {standings.map((s, i) => (
        <div
          key={s.player}
          className={`flex items-center gap-3 px-3 py-2 rounded-lg border ${
            i === 0
              ? "border-primary/50 bg-primary/10"
              : "border-border bg-surface"
          }`}
        >
          <span
            className={`w-6 text-center text-sm font-bold ${
              i === 0 ? "text-primary" : "text-on-surface-muted"
            }`}
          >
            {i + 1}
          </span>
          <PlayerAvatar id={s.player} size={26} />
          <span className="flex-1 min-w-0 truncate text-sm">
            <PlayerName id={s.player} />
            {players.get(s.player)?.isGuest && (
              <span className="ml-1.5 text-xs text-on-surface-muted">
                guest
              </span>
            )}
          </span>
          <span
            className={`text-sm font-semibold tabular-nums ${
              i === 0 ? "text-primary" : "text-on-surface"
            }`}
          >
            {s.score}
          </span>
        </div>
      ))}
    </div>
  );
}

function FilterChips({
  ids,
  filter,
  onChange,
}: {
  ids: string[];
  filter: string | null;
  onChange: (id: string | null) => void;
}) {
  const { players } = useLog();
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {ids.map((id) => {
        const active = filter === id;
        const color = players.get(id)?.color;
        return (
          <button
            key={id}
            onClick={() => onChange(active ? null : id)}
            className={`inline-flex items-center gap-1.5 pl-1 pr-2.5 py-1 rounded-full border text-xs font-medium transition-colors duration-150 ${
              active
                ? "text-on-surface"
                : "border-border text-on-surface-muted hover:text-on-surface"
            }`}
            style={
              active
                ? { borderColor: color, backgroundColor: `${color}22` }
                : undefined
            }
          >
            <PlayerAvatar id={id} size={18} />
            {players.get(id)?.name ?? "Unknown"}
          </button>
        );
      })}
    </div>
  );
}

function Log({ events }: { events: GameEvent[] }) {
  const [filter, setFilter] = useState<string | null>(null);
  const [collapsed, setCollapsed] = useState<Set<number>>(new Set());

  const players = useMemo(() => buildPlayers(events), [events]);
  const timeline = useMemo(() => buildTimeline(events), [events]);
  const ctx = useMemo(
    () => ({ players, start: timeline.start, filter }),
    [players, timeline.start, filter]
  );

  const playerIds = useMemo(() => {
    const fromStandings = timeline.standings.map((s) => s.player);
    if (fromStandings.length) return fromStandings;
    return [...players.values()]
      .filter((p) => p.role === "player")
      .map((p) => p.id);
  }, [timeline.standings, players]);

  const toggle = (index: number) =>
    setCollapsed((prev) => {
      const next = new Set(prev);
      if (!next.delete(index)) next.add(index);
      return next;
    });

  const questionNumbers = useMemo(() => {
    const numbers = new Map<number, number>();
    for (const round of timeline.rounds) {
      for (const entry of round.entries) {
        if (entry.kind === "question") {
          numbers.set(entry.question.selected.seq, numbers.size + 1);
        }
      }
    }
    return numbers;
  }, [timeline.rounds]);

  return (
    <LogContext.Provider value={ctx}>
      <div className="flex flex-col gap-5">
        <div className="flex flex-col gap-1">
          <h1 className="text-xl font-bold text-on-background">
            {timeline.pack ?? "Game log"}
          </h1>
          <p className="text-sm text-on-surface-muted">
            {formatDuration(timeline.end - timeline.start)} ·{" "}
            {timeline.questionCount} question
            {timeline.questionCount === 1 ? "" : "s"} · {playerIds.length}{" "}
            player{playerIds.length === 1 ? "" : "s"}
          </p>
        </div>

        {timeline.standings.length > 0 && (
          <Standings standings={timeline.standings} />
        )}

        <ScoreChart events={events} playerIds={playerIds} />

        <FilterChips ids={playerIds} filter={filter} onChange={setFilter} />

        {timeline.rounds.map((round, index) => {
          const isCollapsed = collapsed.has(index);
          return (
            <section key={index} className="flex flex-col gap-3">
              <button
                onClick={() => toggle(index)}
                className="flex items-center gap-2 text-left group"
                aria-expanded={!isCollapsed}
              >
                <FiChevronDown
                  size={16}
                  className={`text-on-surface-muted transition-transform duration-150 ${
                    isCollapsed ? "-rotate-90" : ""
                  }`}
                />
                <h2 className="text-base font-semibold text-on-surface">
                  {round.name}
                </h2>
                <div className="flex-1 h-px bg-border" />
              </button>
              {!isCollapsed &&
                round.entries.map((entry) =>
                  entry.kind === "question" ? (
                    <QuestionCard
                      key={entry.question.selected.seq}
                      question={entry.question}
                      number={questionNumbers.get(entry.question.selected.seq)!}
                    />
                  ) : (
                    <EventRow key={entry.event.seq} event={entry.event} />
                  )
                )}
            </section>
          );
        })}

        {timeline.final.length > 0 && <FinalRound events={timeline.final} />}
      </div>
    </LogContext.Provider>
  );
}

export default function GameLog({
  id,
  backHref,
}: {
  id: string;
  backHref: string;
}) {
  const { data, error, isPending } = useQuery({
    queryKey: ["gameEvents", id],
    queryFn: () => getGameEvents(id),
    retry: false,
    staleTime: Infinity,
  });

  const status = (error as { status?: number } | null)?.status;

  return (
    <div className="max-w-3xl mx-auto px-4 py-5 sm:py-8 flex flex-col gap-5">
      <Link
        href={backHref}
        className="inline-flex items-center gap-1.5 self-start text-sm text-on-surface-muted hover:text-on-surface transition-colors duration-150"
      >
        <FiArrowLeft size={14} />
        Back to history
      </Link>

      {isPending ? (
        <p className="py-16 text-center text-sm text-on-surface-muted">
          Loading…
        </p>
      ) : error ? (
        status === 404 ? (
          <Message
            title="No log recorded"
            text="This game has no event log. Games played before logging was added don't have one."
          />
        ) : status === 403 ? (
          <Message
            title="No access"
            text="Only participants of a game can view its log."
          />
        ) : (
          <Message
            title="Could not load the log"
            text="Something went wrong. Try again later."
          />
        )
      ) : data.events.length === 0 ? (
        <Message title="No log recorded" text="This game has no events." />
      ) : (
        <Log events={data.events} />
      )}
    </div>
  );
}
