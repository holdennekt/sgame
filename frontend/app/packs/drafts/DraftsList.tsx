"use client";

import { useState, useRef, useEffect } from "react";
import { useDebouncedCallback } from "use-debounce";
import {
  useInfiniteQuery,
  useQueryClient,
  InfiniteData,
} from "@tanstack/react-query";
import { useVirtualizer } from "@tanstack/react-virtual";
import { ToastContainer, toast } from "react-toastify";
import { getDrafts, deleteDraft } from "@/app/api";
import { isError } from "@/middleware";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { SearchResponse } from "@/types/search";
import { PackDraft, countIncompleteQuestions } from "@/types/pack_draft";
import { IoIosSearch } from "react-icons/io";
import { FiTrash2, FiEdit2, FiArrowLeft } from "react-icons/fi";

function DraftCard({
  draft,
  onDelete,
}: {
  draft: PackDraft;
  onDelete: () => Promise<void>;
}) {
  const incomplete = countIncompleteQuestions(draft);

  return (
    <div className="bg-surface border border-border rounded-md p-4 flex gap-4 hover:border-primary transition-colors duration-150">
      <div className="flex-1 min-w-0 flex flex-col gap-1.5">
        <div className="flex items-center gap-2">
          <Link
            className="text-base font-semibold text-on-surface hover:text-primary transition-colors duration-150 truncate"
            href={`/packs/drafts/${draft.id}`}
            title={draft.name || "Untitled draft"}
          >
            {draft.name || (
              <em className="opacity-50 font-normal">Untitled draft</em>
            )}
          </Link>
          {incomplete > 0 && (
            <span className="shrink-0 text-xs font-medium px-1.5 py-0.5 rounded bg-warning/20 text-warning">
              {incomplete} missing {incomplete === 1 ? "answer" : "answers"}
            </span>
          )}
        </div>

        <div className="flex flex-wrap items-center gap-x-3 gap-y-0.5 text-xs text-on-surface-muted">
          <span>
            {draft.rounds.length}{" "}
            {draft.rounds.length === 1 ? "round" : "rounds"}
          </span>
          <span>Created {new Date(draft.createdAt).toLocaleDateString()}</span>
          {draft.updatedAt !== draft.createdAt && (
            <span>
              Updated {new Date(draft.updatedAt).toLocaleDateString()}
            </span>
          )}
        </div>
      </div>

      <div className="flex items-center gap-1.5 shrink-0">
        <Link
          href={`/packs/drafts/${draft.id}`}
          className="h-8 w-8 inline-flex items-center justify-center rounded-lg border border-border text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150"
          title="Edit draft"
        >
          <FiEdit2 size={13} />
        </Link>
        <button
          onClick={onDelete}
          className="h-8 w-8 inline-flex items-center justify-center rounded-lg border border-border text-on-surface-muted hover:text-danger hover:border-danger transition-colors duration-150"
          title="Delete draft"
        >
          <FiTrash2 size={13} />
        </button>
      </div>
    </div>
  );
}

export default function DraftsList() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [filter, setFilter] = useState("");
  const [debouncedFilter, setDebouncedFilter] = useState("");
  const scrollRef = useRef<HTMLDivElement>(null);

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPending } =
    useInfiniteQuery<SearchResponse<PackDraft>>({
      queryKey: ["drafts", debouncedFilter],
      queryFn: ({ pageParam }) =>
        getDrafts(debouncedFilter, pageParam as number),
      initialPageParam: 1,
      getNextPageParam: (lastPage) =>
        lastPage.hasNext ? lastPage.page + 1 : undefined,
    });

  const debounceFilter = useDebouncedCallback((value: string) => {
    setDebouncedFilter(value);
  }, 400);

  const drafts = data?.pages.flatMap((p) => p.items) ?? [];

  const handleDelete = async (draftId: string) => {
    if (!confirm("Delete this draft? This cannot be undone.")) return;
    try {
      await deleteDraft(draftId);
      queryClient.setQueryData(
        ["drafts", debouncedFilter],
        (old: InfiniteData<SearchResponse<PackDraft>>) => ({
          ...old,
          pages: old.pages.map((page) => ({
            ...page,
            items: page.items.filter((d) => d.id !== draftId),
          })),
        })
      );
    } catch (e) {
      toast.error(isError(e) ? e.error : "Delete failed", {
        containerId: "drafts",
      });
    }
  };

  const virtualizer = useVirtualizer({
    count: drafts.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => 88,
    overscan: 5,
    measureElement: (el) => el.getBoundingClientRect().height,
  });

  const virtualItems = virtualizer.getVirtualItems();
  const lastVirtualItem = virtualItems[virtualItems.length - 1];

  useEffect(() => {
    if (!lastVirtualItem) return;
    if (
      lastVirtualItem.index >= drafts.length - 1 &&
      hasNextPage &&
      !isFetchingNextPage
    ) {
      fetchNextPage();
    }
  }, [
    lastVirtualItem?.index,
    drafts.length,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
  ]);

  return (
    <>
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded-md bg-surface text-on-surface p-4 border border-border gap-3">
          <div className="flex items-center gap-3 shrink-0">
            <div className="flex items-center gap-2 shrink-0">
              <Link
                href="/packs"
                className="flex items-center gap-1 text-sm text-on-surface-muted hover:text-on-surface transition-colors duration-150"
              >
                <FiArrowLeft size={14} />
                Packs
              </Link>
              <span className="text-on-surface-muted">/</span>
              <span className="text-sm font-medium text-on-surface">
                Drafts
              </span>
            </div>
            <div className="ml-auto relative flex-1 sm:flex-none">
              <div className="pointer-events-none absolute inset-y-0 left-2.5 flex items-center text-on-surface-muted">
                <IoIosSearch size={16} />
              </div>
              <input
                className="h-9 w-full sm:w-52 pl-8 pr-3 rounded-lg border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring transition-[border-color] duration-150"
                type="text"
                placeholder="Search drafts..."
                value={filter}
                onChange={(e) => {
                  setFilter(e.target.value);
                  debounceFilter(e.target.value.trim());
                }}
              />
            </div>
          </div>

          <div ref={scrollRef} className="flex-1 overflow-y-auto">
            {isPending ? (
              <div className="w-full h-full flex items-center justify-center text-on-surface-muted select-none">
                <p className="text-sm">Loading…</p>
              </div>
            ) : drafts.length ? (
              <div
                style={{
                  height: `${virtualizer.getTotalSize()}px`,
                  position: "relative",
                }}
              >
                {virtualItems.map((virtualItem) => (
                  <div
                    key={virtualItem.key}
                    data-index={virtualItem.index}
                    ref={virtualizer.measureElement}
                    style={{
                      position: "absolute",
                      top: 0,
                      left: 0,
                      width: "100%",
                      transform: `translateY(${virtualItem.start}px)`,
                      paddingBottom: "8px",
                    }}
                  >
                    <DraftCard
                      draft={drafts[virtualItem.index]}
                      onDelete={() =>
                        handleDelete(drafts[virtualItem.index].id)
                      }
                    />
                  </div>
                ))}
              </div>
            ) : (
              <div className="w-full h-full flex items-center justify-center text-on-surface-muted select-none">
                <p className="text-sm">No drafts yet</p>
              </div>
            )}
            {isFetchingNextPage && (
              <p className="py-2 text-center text-xs text-on-surface-muted">
                Loading…
              </p>
            )}
          </div>
        </div>
      </main>

      <ToastContainer
        containerId="drafts"
        position="bottom-left"
        theme="colored"
      />
    </>
  );
}
