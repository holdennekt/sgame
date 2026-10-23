"use client";

import { FiAward, FiHelpCircle } from "react-icons/fi";
import { GameEvent } from "@/types/game_event";
import EventRow from "./EventRow";
import { AttachmentBadge } from "./QuestionCard";

export default function FinalRound({ events }: { events: GameEvent[] }) {
  const chosen = events.find((e) => e.type === "final_category_chosen");

  return (
    <div className="border border-primary/40 rounded-lg bg-surface overflow-hidden">
      <div className="px-3 sm:px-4 py-3 flex flex-col gap-2 border-b border-border">
        <div className="flex items-center gap-2 text-sm font-semibold text-on-surface">
          <FiAward className="text-primary" size={16} />
          Final round
          {chosen?.type === "final_category_chosen" && (
            <span className="font-normal text-on-surface-muted">
              · {chosen.data.category}
            </span>
          )}
          {chosen?.type === "final_category_chosen" && (
            <AttachmentBadge kind={chosen.data.attachmentType} />
          )}
        </div>
        {chosen?.type === "final_category_chosen" && chosen.data.text && (
          <p className="text-sm text-on-surface break-words">
            {chosen.data.text}
          </p>
        )}
      </div>
      <div className="px-3 sm:px-4 py-2 flex flex-col">
        {events.map((e) => (
          <EventRow key={e.seq} event={e} />
        ))}
      </div>
      {chosen?.type === "final_category_chosen" &&
      chosen.data.answers?.length ? (
        <div className="px-3 sm:px-4 py-2.5 border-t border-border bg-surface-raised/50 flex items-center gap-1.5 text-sm text-on-surface">
          <FiHelpCircle size={14} className="text-on-surface-muted" />
          {chosen.data.answers.join(" / ")}
        </div>
      ) : null}
    </div>
  );
}
