"use client";

import { HiddenPack } from "@/types/pack";
import Link from "next/link";
import { FaLock, FaGlobe } from "react-icons/fa6";
import { FiTrash2 } from "react-icons/fi";

export function PackCard({
  pack,
  isOwn,
  onPlay,
  onDelete,
}: {
  pack: HiddenPack;
  isOwn: boolean;
  onPlay: () => void;
  onDelete?: () => Promise<void>;
}) {
  return (
    <div className="flex items-center gap-3 px-4 py-3 border border-border rounded-lg hover:border-primary/60 hover:bg-surface-raised/40 transition-colors duration-150">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5 min-w-0">
          {isOwn ? (
            <Link
              href={`/packs/${pack.id}`}
              className="font-medium text-on-surface hover:text-primary transition-colors truncate"
            >
              {pack.name}
            </Link>
          ) : (
            <span className="font-medium text-on-surface truncate">
              {pack.name}
            </span>
          )}
          <span className="text-on-surface-muted shrink-0">
            {pack.type === "public" ? (
              <FaGlobe size={10} />
            ) : (
              <FaLock size={10} />
            )}
          </span>
        </div>
        <p className="text-xs text-on-surface-muted mt-0.5">
          {pack.rounds.length} round{pack.rounds.length !== 1 ? "s" : ""}
          {" · "}
          {pack.rounds.reduce((s, r) => s + r.categories.length, 0)} categories
        </p>
      </div>
      <div className="flex items-center gap-2 shrink-0">
        {isOwn && (
          <>
            <Link
              href={`/packs/${pack.id}?edit=true`}
              className="px-2.5 py-1.5 rounded-md text-xs font-medium border border-border text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150"
            >
              Edit
            </Link>
            {onDelete && (
              <button
                onClick={onDelete}
                className="h-7 w-7 inline-flex items-center justify-center rounded-md text-on-surface-muted hover:text-danger hover:bg-surface-raised transition-colors duration-150"
                title="Delete pack"
              >
                <FiTrash2 size={13} />
              </button>
            )}
          </>
        )}
        <button
          onClick={onPlay}
          className="px-2.5 py-1.5 rounded-md text-xs font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
        >
          Play
        </button>
      </div>
    </div>
  );
}
