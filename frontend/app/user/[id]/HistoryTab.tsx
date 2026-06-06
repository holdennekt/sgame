"use client";

import { useEffect, useRef } from "react";
import { useInfiniteQuery } from "@tanstack/react-query";
import { getGameHistory } from "@/app/api";
import { GameHistoryEntry } from "@/types/room";
import { SearchResponse } from "@/types/search";
import { HistoryCard } from "./HistoryCard";

export default function HistoryTab() {
  const sentinelRef = useRef<HTMLDivElement>(null);

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPending } =
    useInfiniteQuery<SearchResponse<GameHistoryEntry>>({
      queryKey: ["gameHistory"],
      queryFn: ({ pageParam }) => getGameHistory(pageParam as number),
      initialPageParam: 1,
      getNextPageParam: (lastPage) =>
        lastPage.hasNext ? lastPage.page + 1 : undefined,
    });

  const onIntersect = useRef<() => void>(() => {});
  onIntersect.current = () => {
    if (hasNextPage) fetchNextPage();
  };

  useEffect(() => {
    const obs = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) onIntersect.current();
      },
      { rootMargin: "100px" }
    );
    if (sentinelRef.current) obs.observe(sentinelRef.current);
    return () => obs.disconnect();
  }, []);

  const history = data?.pages.flatMap((p) => p.items) ?? [];

  return (
    <div className="flex flex-col gap-2">
      {isPending ? (
        <p className="py-8 text-center text-sm text-on-surface-muted">
          Loading…
        </p>
      ) : history.length ? (
        <>
          {history.map((entry) => (
            <HistoryCard key={entry.id} entry={entry} />
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
          No games played yet
        </p>
      )}
    </div>
  );
}
