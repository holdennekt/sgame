"use client";

import { useMemo } from "react";
import { GameEvent } from "@/types/game_event";
import { useLog } from "./LogContext";

interface Point {
  t: number;
  score: number;
}

const W = 640;
const H = 190;
const PAD = { top: 12, right: 12, bottom: 18, left: 40 };

export default function ScoreChart({
  events,
  playerIds,
}: {
  events: GameEvent[];
  playerIds: string[];
}) {
  const { players, start, filter } = useLog();

  const { series, end, min, max } = useMemo(() => {
    const series = new Map<string, Point[]>(
      playerIds.map((id) => [id, [{ t: 0, score: 0 }]])
    );
    let end = 0;
    for (const e of events) {
      const t = Date.parse(e.at) - start;
      end = Math.max(end, t);
      let player: string | null = null;
      let score = 0;
      if (
        e.type === "answer_validated" ||
        e.type === "final_answer_validated"
      ) {
        player = e.data.player;
        score = e.data.score;
      } else if (e.type === "score_changed") {
        player = e.data.player;
        score = e.data.to;
      }
      if (player === null || !series.has(player)) continue;
      series.get(player)!.push({ t, score });
    }
    const all = [...series.values()].flat().map((p) => p.score);
    return {
      series,
      end: Math.max(end, 1),
      min: Math.min(0, ...all),
      max: Math.max(1, ...all),
    };
  }, [events, playerIds, start]);

  const hasChanges = [...series.values()].some((s) => s.length > 1);
  if (!hasChanges) return null;

  const x = (t: number) => PAD.left + (t / end) * (W - PAD.left - PAD.right);
  const y = (s: number) =>
    PAD.top + (1 - (s - min) / (max - min)) * (H - PAD.top - PAD.bottom);

  const ticks = [min, (min + max) / 2, max].map(Math.round);

  return (
    <div className="border border-border rounded-lg bg-surface p-3 sm:p-4">
      <h3 className="text-sm font-semibold text-on-surface mb-2">
        Score progression
      </h3>
      <svg
        viewBox={`0 0 ${W} ${H}`}
        className="w-full h-auto"
        role="img"
        aria-label="Score progression per player"
      >
        {ticks.map((tick) => (
          <g key={tick}>
            <line
              x1={PAD.left}
              x2={W - PAD.right}
              y1={y(tick)}
              y2={y(tick)}
              className="stroke-border"
              strokeWidth={1}
              strokeDasharray={tick === 0 ? undefined : "3 4"}
            />
            <text
              x={PAD.left - 6}
              y={y(tick) + 4}
              textAnchor="end"
              className="fill-on-surface-muted"
              fontSize={11}
            >
              {tick}
            </text>
          </g>
        ))}
        {[...series.entries()].map(([id, points]) => {
          const last = points[points.length - 1];
          const path = points
            .map((p, i) =>
              i === 0 ? `M${x(p.t)},${y(p.score)}` : `H${x(p.t)}V${y(p.score)}`
            )
            .join("");
          const dim = filter !== null && filter !== id;
          return (
            <g
              key={id}
              style={{ opacity: dim ? 0.15 : 1 }}
              className="transition-opacity duration-150"
            >
              <path
                d={`${path}H${x(end)}`}
                fill="none"
                stroke={players.get(id)?.color ?? "currentColor"}
                strokeWidth={2}
                strokeLinejoin="round"
              />
              <circle
                cx={x(end)}
                cy={y(last.score)}
                r={3}
                fill={players.get(id)?.color ?? "currentColor"}
              />
            </g>
          );
        })}
      </svg>
    </div>
  );
}
