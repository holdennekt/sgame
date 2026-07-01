"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { toast } from "react-toastify";
import { User } from "@/middleware";
import { RoomLobby } from "@/types/room";
import { FaLock, FaGlobe, FaEye } from "react-icons/fa6";
import { TbRobot } from "react-icons/tb";
import { getAvatar } from "@/components/UserAvatar";
import { joinRoom } from "@/app/api";
import { isError } from "@/middleware";

export default function RoomLobbyCard({
  room,
  openPasswordModal,
  openSpectateModal,
}: {
  room: RoomLobby;
  openPasswordModal: (roomId: string) => void;
  openSpectateModal: (roomId: string) => void;
}) {
  const router = useRouter();
  const [joining, setJoining] = useState(false);

  const playersSlots = new Array<User | null>(room.maxPlayers).fill(null);
  for (const [index, player] of room.players.entries()) {
    playersSlots[index] = player;
  }

  const isPlaying = room.status === "playing";

  const handleJoin = async () => {
    setJoining(true);
    try {
      await joinRoom(room.id, undefined);
      router.push(`/rooms/${room.id}`);
    } catch (e) {
      if (isError(e)) toast.error(e.error, { containerId: "lobby" });
    } finally {
      setJoining(false);
    }
  };

  return (
    <div className="p-3 flex flex-col gap-2.5 border-b border-border hover:bg-surface-raised transition-colors duration-150">
      {/* Header */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1.5 min-w-0">
          <p
            className="text-sm font-semibold truncate text-on-surface"
            title={room.name}
          >
            {room.name}
          </p>
          <span className="shrink-0 text-on-surface-muted">
            {room.type === "public" ? (
              <FaGlobe size={11} />
            ) : (
              <FaLock size={11} />
            )}
          </span>
        </div>
        <span className="text-xs shrink-0 text-on-surface-muted tabular-nums">
          {room.players.length}/{room.maxPlayers}
        </span>
      </div>

      {/* Pack */}
      <p className="text-xs text-on-surface-muted">
        <Link
          className="text-on-surface hover:text-primary transition-colors duration-150"
          href={`/packs/${room.packPreview.id}`}
          target="_blank"
        >
          {room.packPreview.name}
        </Link>
      </p>

      {/* Avatars */}
      <div className="flex items-center gap-1 overflow-x-auto">
        {room.options.aiHost ? (
          <div className="w-7 h-7 rounded border-2 border-primary shrink-0 flex items-center justify-center bg-primary/10 text-primary">
            <TbRobot size={16} />
          </div>
        ) : (
          <div className="w-7 h-7 rounded overflow-hidden border-2 border-primary shrink-0">
            {getAvatar(room.moderator)}
          </div>
        )}
        <div className="w-px h-5 bg-border shrink-0 mx-1" />
        {playersSlots.map((player, index) => (
          <div
            key={index}
            className="w-7 h-7 rounded overflow-hidden border border-border shrink-0"
          >
            {getAvatar(player)}
          </div>
        ))}
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between">
        <span
          className={`text-xs font-medium ${
            isPlaying ? "text-secondary" : "text-yellow-500"
          }`}
        >
          {room.status}
        </span>
        <div className="flex items-center gap-2">
          {room.type === "public" ? (
            <Link
              className="inline-flex items-center gap-1 justify-center px-3 py-1 rounded-lg text-xs font-medium bg-surface-raised text-on-surface border border-border hover:bg-border transition-colors duration-150"
              href={`/rooms/${room.id}?spectate=true`}
            >
              <FaEye size={11} />
              Watch
            </Link>
          ) : (
            <button
              className="inline-flex items-center gap-1 justify-center px-3 py-1 rounded-lg text-xs font-medium bg-surface-raised text-on-surface border border-border hover:bg-border transition-colors duration-150"
              onClick={() => openSpectateModal(room.id)}
            >
              <FaEye size={11} />
              Watch
            </button>
          )}
          {room.type === "public" ? (
            <button
              className="inline-flex items-center justify-center px-3 py-1 rounded-lg text-xs font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 disabled:opacity-50"
              onClick={handleJoin}
              disabled={joining}
            >
              {joining ? "Joining…" : "Join"}
            </button>
          ) : (
            <button
              className="inline-flex items-center justify-center px-3 py-1 rounded-lg text-xs font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
              onClick={() => openPasswordModal(room.id)}
            >
              Join
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
