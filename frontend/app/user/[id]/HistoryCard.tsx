"use client";

import Link from "next/link";
import { GameHistoryEntry } from "@/types/room";

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function HistoryCard({ entry }: { entry: GameHistoryEntry }) {
  const sorted = [...entry.players].sort((a, b) => b.score - a.score);
  return (
    <div className="px-4 py-3 border border-border rounded-lg">
      <div className="flex items-center justify-between gap-2 mb-2">
        <div>
          <p className="font-medium text-on-surface text-sm">{entry.name}</p>
          <p className="text-xs text-on-surface-muted">
            {entry.packPreview.name}
          </p>
        </div>
        <span className="text-xs text-on-surface-muted shrink-0">
          {formatDate(entry.finishedAt)}
        </span>
      </div>
      <div className="flex flex-wrap gap-x-4 gap-y-1">
        {sorted.map((p, i) => (
          <div key={p.id} className="flex items-center gap-1.5 text-xs">
            <span
              className={`font-bold ${
                i === 0 ? "text-primary" : "text-on-surface-muted"
              }`}
            >
              #{i + 1}
            </span>
            {p.isGuest ? (
              <span
                className={
                  i === 0
                    ? "text-on-surface font-medium"
                    : "text-on-surface-muted"
                }
              >
                {p.name}
              </span>
            ) : (
              <Link
                href={`/user/${p.id}`}
                className={`hover:text-primary transition-colors duration-150 ${
                  i === 0
                    ? "text-on-surface font-medium"
                    : "text-on-surface-muted"
                }`}
              >
                {p.name}
              </Link>
            )}
            <span
              className={
                i === 0 ? "text-primary font-semibold" : "text-on-surface-muted"
              }
            >
              {p.score}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
