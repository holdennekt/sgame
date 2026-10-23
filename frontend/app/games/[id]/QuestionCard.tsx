"use client";

import { FiHelpCircle, FiImage, FiMusic, FiVideo } from "react-icons/fi";
import { AttachmentKind, QuestionEndReason } from "@/types/game_event";
import { QuestionType } from "@/types/pack";
import EventRow, { Chip } from "./EventRow";
import { QuestionBlock, involves } from "./helpers";
import { useLog } from "./LogContext";
import { PlayerBadge } from "./PlayerBadge";

const TYPE_LABELS: Record<QuestionType, string> = {
  regular: "Regular",
  auction: "Auction",
  catInBag: "Cat in bag",
};

const ATTACHMENT_ICONS: Record<AttachmentKind, JSX.Element> = {
  image: <FiImage size={13} />,
  audio: <FiMusic size={13} />,
  video: <FiVideo size={13} />,
};

export function AttachmentBadge({ kind }: { kind: AttachmentKind | null }) {
  if (!kind) return null;
  return (
    <span className="inline-flex items-center gap-1 text-xs text-on-surface-muted">
      {ATTACHMENT_ICONS[kind]}
      {kind}
    </span>
  );
}

const OUTCOMES: Record<Exclude<QuestionEndReason, "answered">, string> = {
  nobody_answered: "Nobody answered",
  no_bets: "Nobody bet",
  skipped: "Skipped by host",
};

export default function QuestionCard({
  question,
  number,
}: {
  question: QuestionBlock;
  number: number;
}) {
  const { filter } = useLog();
  const { selected, items, ended } = question;
  const d = selected.data;

  const dimmed =
    filter !== null &&
    ![selected, ...items, ...(ended ? [ended] : [])].some((e) =>
      involves(e, filter)
    );

  return (
    <div
      className={`border border-border rounded-lg bg-surface overflow-hidden transition-opacity duration-150 ${
        dimmed ? "opacity-40" : ""
      }`}
    >
      <div className="px-3 sm:px-4 py-3 flex flex-col gap-2 border-b border-border">
        <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
          <span className="text-xs text-on-surface-muted tabular-nums">
            Q{number}
          </span>
          <span className="text-sm font-semibold text-on-surface">
            {d.category}
          </span>
          <Chip tone="primary">{d.value}</Chip>
          {d.type !== "regular" && (
            <Chip tone="warning">{TYPE_LABELS[d.type]}</Chip>
          )}
          <AttachmentBadge kind={d.attachmentType} />
          {selected.actor && (
            <span className="ml-auto inline-flex items-center gap-1.5 text-xs text-on-surface-muted">
              picked by <PlayerBadge id={selected.actor} />
            </span>
          )}
        </div>
        {d.text && (
          <p className="text-sm text-on-surface break-words">{d.text}</p>
        )}
      </div>

      <div className="px-3 sm:px-4 py-2 flex flex-col">
        {items.map((e) => (
          <EventRow key={e.seq} event={e} />
        ))}
        {items.length === 0 && (
          <p className="py-1 text-xs text-on-surface-muted">No activity</p>
        )}
      </div>

      {ended && (
        <div className="px-3 sm:px-4 py-2.5 border-t border-border bg-surface-raised/50 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
          <span className="inline-flex items-center gap-1.5">
            <FiHelpCircle size={14} className="text-on-surface-muted" />
            {ended.data.answers?.length ? (
              <span className="text-on-surface">
                {ended.data.answers.join(" / ")}
              </span>
            ) : (
              <span className="text-on-surface-muted">No answer recorded</span>
            )}
          </span>
          <span className="ml-auto text-xs text-on-surface-muted inline-flex items-center gap-1.5">
            {ended.data.reason === "answered" ? (
              <>
                Answered by{" "}
                {ended.data.winner && <PlayerBadge id={ended.data.winner} />}
              </>
            ) : (
              OUTCOMES[ended.data.reason]
            )}
          </span>
        </div>
      )}
    </div>
  );
}
