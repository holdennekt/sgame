"use client";

import React, { useState, useRef, useEffect } from "react";
import { useDebouncedCallback } from "use-debounce";
import { useInfiniteQuery } from "@tanstack/react-query";
import { useVirtualizer } from "@tanstack/react-virtual";
import { ToastContainer } from "react-toastify";
import { getPacks, deletePack } from "@/app/actions";
import { isError } from "@/middleware";
import { toast } from "react-toastify";
import NewRoomModal from "@/components/NewRoomModal";
import { useRouter } from "next/navigation";
import { useRequiredUser } from "@/contexts/UserContext";
import Link from "next/link";
import { SearchResponse } from "@/types/search";
import { HiddenPack, PackPreview } from "@/types/pack";
import { FaLock, FaGlobe } from "react-icons/fa6";
import { FiTrash2 } from "react-icons/fi";
import { IoIosSearch, IoIosAdd } from "react-icons/io";

function PackCard({ pack, onPlay, onDelete }: { pack: HiddenPack; onPlay: () => void; onDelete: () => Promise<void> }) {
  const user = useRequiredUser();
  const isOwn = user.id === pack.createdBy.id;

  return (
    <div className="bg-surface border border-border rounded-md p-4 flex gap-4 hover:border-primary transition-colors duration-150">
      <div className="flex-1 min-w-0 flex flex-col gap-2">
        <div>
          <div className="flex items-center gap-1.5">
            {isOwn ? (
              <Link
                className="text-base font-semibold text-on-surface hover:text-primary transition-colors duration-150 truncate"
                href={`/packs/${pack.id}`}
                title={pack.name}
              >
                {pack.name}
              </Link>
            ) : (
              <span className="text-base font-semibold text-on-surface truncate" title={pack.name}>
                {pack.name}
              </span>
            )}
            <span className="shrink-0 text-on-surface-muted">
              {pack.type === "public" ? <FaGlobe size={11} /> : <FaLock size={11} />}
            </span>
          </div>
          <p className="text-xs text-on-surface-muted mt-0.5">
            by{" "}
            <Link
              className="text-on-surface hover:text-primary transition-colors duration-150"
              href={`/user/${pack.createdBy.id}`}
            >
              {pack.createdBy.name}
            </Link>
          </p>
        </div>

        <div className="flex flex-col gap-0.5">
          {pack.rounds.map((round, ri) => (
            <div key={ri} className="flex items-center gap-x-1.5 text-[12px] text-on-surface-muted flex-wrap">
              <span className="font-medium text-on-surface">{round.name}</span>
              {round.categories.map((c, ci) => (
                <React.Fragment key={ci}>
                  <span>·</span>
                  <span>{c.name}</span>
                </React.Fragment>
              ))}
            </div>
          ))}
          {pack.finalRound.categories.length > 0 && (
            <div className="flex items-center gap-x-1.5 text-[12px] text-on-surface-muted flex-wrap">
              <span className="font-medium text-on-surface">Final</span>
              {pack.finalRound.categories.map((c, ci) => (
                <React.Fragment key={ci}>
                  <span>·</span>
                  <span>{c.name}</span>
                </React.Fragment>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="flex flex-col items-end justify-between shrink-0 gap-2">
        <button
          className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
          onClick={onPlay}
        >
          Play
        </button>
        {isOwn && (
          <div className="flex items-center gap-1.5">
            <Link
              className="inline-flex items-center justify-center px-3 py-1.5 rounded-lg text-xs font-medium border border-border text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150"
              href={`/packs/${pack.id}?edit=true`}
            >
              Edit
            </Link>
            <button
              onClick={onDelete}
              className="h-8 w-8 inline-flex items-center justify-center rounded-lg border border-border text-on-surface-muted hover:text-danger hover:border-danger transition-colors duration-150"
              title="Delete pack"
            >
              <FiTrash2 size={13} />
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default function PacksList() {
  const router = useRouter();
  const [filter, setFilter] = useState("");
  const [debouncedFilter, setDebouncedFilter] = useState("");
  const [newRoomModal, setNewRoomModal] = useState<{ isOpen: boolean; pack?: PackPreview }>({
    isOpen: false,
  });
  const scrollRef = useRef<HTMLDivElement>(null);

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPending, refetch } =
    useInfiniteQuery<SearchResponse<HiddenPack>>({
      queryKey: ["packs", debouncedFilter],
      queryFn: async ({ pageParam }) => {
        const result = await getPacks(debouncedFilter, pageParam as number);
        if (isError(result)) throw new Error(result.error);
        return result;
      },
      initialPageParam: 1,
      getNextPageParam: (lastPage) => lastPage.hasNext ? lastPage.page + 1 : undefined,
    });

  const debounceFilter = useDebouncedCallback((value: string) => {
    setDebouncedFilter(value);
  }, 400);

  const packs = data?.pages.flatMap((p) => p.items) ?? [];

  const handleDelete = async (packId: string) => {
    if (!confirm("Delete this pack?")) return;
    const result = await deletePack(packId);
    if (isError(result)) { toast.error(result.error, { containerId: "packs" }); return; }
    refetch();
  };

  const virtualizer = useVirtualizer({
    count: packs.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => 110,
    overscan: 5,
    measureElement: (el) => el.getBoundingClientRect().height,
  });

  const virtualItems = virtualizer.getVirtualItems();
  const lastVirtualItem = virtualItems[virtualItems.length - 1];

  useEffect(() => {
    if (!lastVirtualItem) return;
    if (lastVirtualItem.index >= packs.length - 1 && hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [lastVirtualItem?.index, packs.length, hasNextPage, isFetchingNextPage, fetchNextPage]);

  return (
    <>
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded-md bg-surface text-on-surface p-4 border border-border gap-3">
          <div className="flex items-center gap-2 shrink-0">
            <div className="relative w-full sm:w-auto">
              <div className="pointer-events-none absolute inset-y-0 left-2.5 flex items-center text-on-surface-muted">
                <IoIosSearch size={16} />
              </div>
              <input
                className="h-9 w-full sm:w-52 pl-8 pr-3 rounded-lg border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring transition-[border-color] duration-150"
                type="text"
                placeholder="Search packs..."
                value={filter}
                onChange={(e) => {
                  setFilter(e.target.value);
                  debounceFilter(e.target.value.trim());
                }}
              />
            </div>
            <button
              className="ml-auto inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 shrink-0"
              onClick={() => router.push("/packs/new")}
            >
              <IoIosAdd size={16} />
              New pack
            </button>
          </div>

          <div ref={scrollRef} className="flex-1 overflow-y-auto">
            {isPending ? (
              <div className="w-full h-full flex items-center justify-center text-on-surface-muted select-none">
                <p className="text-sm">Loading…</p>
              </div>
            ) : packs.length ? (
              <div style={{ height: `${virtualizer.getTotalSize()}px`, position: "relative" }}>
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
                    <PackCard
                      pack={packs[virtualItem.index]}
                      onPlay={() => setNewRoomModal({
                        isOpen: true,
                        pack: { id: packs[virtualItem.index].id, name: packs[virtualItem.index].name },
                      })}
                      onDelete={() => handleDelete(packs[virtualItem.index].id)}
                    />
                  </div>
                ))}
              </div>
            ) : (
              <div className="w-full h-full flex items-center justify-center text-on-surface-muted select-none">
                <p className="text-sm">No packs found</p>
              </div>
            )}
            {isFetchingNextPage && (
              <p className="py-2 text-center text-xs text-on-surface-muted">Loading…</p>
            )}
          </div>
        </div>
      </main>

      <NewRoomModal
        isOpen={newRoomModal.isOpen}
        close={() => setNewRoomModal({ isOpen: false })}
        fixedPack={newRoomModal.pack}
      />
      <ToastContainer containerId="packs" position="bottom-left" theme="colored" />
    </>
  );
}
