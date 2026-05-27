"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useInfiniteQuery } from "@tanstack/react-query";
import { useDebouncedCallback } from "use-debounce";
import { IoIosSearch, IoIosAdd } from "react-icons/io";
import { getPacksCreatedBy, deletePack } from "@/app/actions";
import { isError } from "@/middleware";
import { toast } from "react-toastify";
import { HiddenPack, PackPreview } from "@/types/pack";
import { SearchResponse } from "@/types/search";
import { PackCard } from "./PackCard";

export default function PacksTab({
  userId,
  isOwn,
  onPlay,
}: {
  userId: string;
  isOwn: boolean;
  onPlay: (pack: PackPreview) => void;
}) {
  const router = useRouter();
  const [filter, setFilter] = useState("");
  const [debouncedFilter, setDebouncedFilter] = useState("");
  const sentinelRef = useRef<HTMLDivElement>(null);

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isPending,
    refetch,
  } = useInfiniteQuery<SearchResponse<HiddenPack>>({
    queryKey: ["packs", userId, debouncedFilter],
    queryFn: async ({ pageParam }) => {
      const result = await getPacksCreatedBy(
        userId,
        debouncedFilter,
        pageParam as number,
      );
      if (isError(result)) throw new Error(result.error);
      return result;
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.hasNext ? lastPage.page + 1 : undefined,
  });

  const debounceFilter = useDebouncedCallback((value: string) => {
    setDebouncedFilter(value);
  }, 400);

  const onIntersect = useRef<() => void>(() => {});
  onIntersect.current = () => {
    if (hasNextPage) fetchNextPage();
  };

  useEffect(() => {
    const obs = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) onIntersect.current();
      },
      { rootMargin: "100px" },
    );
    if (sentinelRef.current) obs.observe(sentinelRef.current);
    return () => obs.disconnect();
  }, []);

  const packs = data?.pages.flatMap((p) => p.items) ?? [];

  const handleDelete = async (packId: string) => {
    if (!confirm("Delete this pack?")) return;
    const result = await deletePack(packId);
    if (isError(result)) {
      toast.error(result.error, { containerId: "profile" });
      return;
    }
    refetch();
  };

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <div className="pointer-events-none absolute inset-y-0 left-2 flex items-center text-on-surface-muted">
            <IoIosSearch size={14} />
          </div>
          <input
            className="h-8 w-full pl-7 pr-2.5 rounded-md border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring transition-[border-color] duration-150"
            placeholder="Search packs…"
            value={filter}
            onChange={(e) => {
              setFilter(e.target.value);
              debounceFilter(e.target.value.trim());
            }}
          />
        </div>
        {isOwn && (
          <button
            onClick={() => router.push("/packs/new")}
            className="inline-flex items-center gap-1 px-3 py-1.5 rounded-md text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 shrink-0"
          >
            <IoIosAdd size={16} />
            New
          </button>
        )}
      </div>

      <div className="flex flex-col gap-2">
        {isPending ? (
          <p className="py-8 text-center text-sm text-on-surface-muted">
            Loading…
          </p>
        ) : packs.length ? (
          <>
            {packs.map((pack, i) => (
              <PackCard
                key={pack.id ?? i}
                pack={pack}
                isOwn={isOwn}
                onPlay={() => onPlay({ id: pack.id, name: pack.name })}
                onDelete={isOwn ? () => handleDelete(pack.id) : undefined}
              />
            ))}
            <div ref={sentinelRef} className="h-1" />
            {isFetchingNextPage && (
              <p className="py-2 text-center text-xs text-on-surface-muted">
                Loading…
              </p>
            )}
          </>
        ) : (
          <p className="py-8 text-center text-sm text-on-surface-muted">
            No packs found
          </p>
        )}
      </div>
    </div>
  );
}
