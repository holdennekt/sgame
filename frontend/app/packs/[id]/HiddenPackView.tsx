import { HiddenPack } from "@/types/pack";
import Link from "next/link";
import { FiArrowLeft, FiGlobe } from "react-icons/fi";

export default function HiddenPackView({
  pack,
  backHref,
}: {
  pack: HiddenPack;
  backHref: string;
}) {
  return (
    <div className="flex flex-col gap-5 h-full overflow-y-auto">
      <div className="flex items-center gap-3">
        <Link
          href={backHref}
          className="flex items-center justify-center w-8 h-8 rounded-lg border border-border text-on-surface-muted hover:bg-surface-raised transition-colors duration-150 shrink-0"
        >
          <FiArrowLeft size={16} />
        </Link>
        <div className="flex-1 min-w-0">
          <h1 className="text-lg font-semibold text-on-surface truncate">
            {pack.name}
          </h1>
          <p className="text-xs text-on-surface-muted">
            by {pack.createdBy.name}
          </p>
        </div>
        <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-surface-raised border border-border text-xs font-medium text-on-surface-muted shrink-0">
          <FiGlobe size={11} />
          Public
        </div>
      </div>

      <div className="flex flex-col gap-3">
        {pack.rounds.map((round, ri) => (
          <div
            key={ri}
            className="rounded-lg border border-border bg-surface-raised overflow-hidden"
          >
            <div className="px-4 py-2.5 border-b border-border bg-surface">
              <span className="text-sm font-medium text-on-surface">
                {round.name}
              </span>
              <span className="ml-2 text-xs text-on-surface-muted">
                {round.categories.length}{" "}
                {round.categories.length === 1 ? "category" : "categories"}
              </span>
            </div>
            <div className="flex flex-wrap gap-2 p-3">
              {round.categories.map((cat, ci) => (
                <span
                  key={ci}
                  className="px-2.5 py-1 rounded-md bg-surface border border-border text-xs text-on-surface-muted"
                >
                  {cat.name}
                </span>
              ))}
            </div>
          </div>
        ))}

        {pack.finalRound.categories.length > 0 && (
          <div className="rounded-lg border border-border bg-surface-raised overflow-hidden">
            <div className="px-4 py-2.5 border-b border-border bg-surface">
              <span className="text-sm font-medium text-on-surface">
                Final Round
              </span>
              <span className="ml-2 text-xs text-on-surface-muted">
                {pack.finalRound.categories.length}{" "}
                {pack.finalRound.categories.length === 1
                  ? "category"
                  : "categories"}
              </span>
            </div>
            <div className="flex flex-wrap gap-2 p-3">
              {pack.finalRound.categories.map((cat, ci) => (
                <span
                  key={ci}
                  className="px-2.5 py-1 rounded-md bg-surface border border-border text-xs text-on-surface-muted"
                >
                  {cat.name}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
