"use client";

import { useRef } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { RoomLobby } from "@/types/room";
import RoomLobbyCard from "./Room";

export default function RoomsList({
  rooms,
  openPasswordModal,
}: {
  rooms: RoomLobby[];
  openPasswordModal: (roomId: string) => void;
}) {
  const scrollRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: rooms.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => 130,
    overscan: 5,
    measureElement: (el) => el.getBoundingClientRect().height,
  });

  if (!rooms.length) {
    return (
      <div className="flex-auto flex flex-col justify-center items-center gap-2 opacity-50">
        <p className="text-sm text-on-surface-muted">
          No rooms yet. Create one!
        </p>
      </div>
    );
  }

  return (
    <div ref={scrollRef} className="flex-auto min-h-0 overflow-y-auto">
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          position: "relative",
        }}
      >
        {virtualizer.getVirtualItems().map((virtualItem) => (
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
            }}
          >
            <RoomLobbyCard
              room={rooms[virtualItem.index]}
              openPasswordModal={openPasswordModal}
            />
          </div>
        ))}
      </div>
    </div>
  );
}
