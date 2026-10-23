"use client";

import { getAvatar } from "@/components/UserAvatar";
import { isSystem } from "./helpers";
import { useLog } from "./LogContext";

export function PlayerAvatar({ id, size = 20 }: { id: string; size?: number }) {
  const { players } = useLog();
  const player = players.get(id);
  return (
    <div
      className="rounded-full overflow-hidden shrink-0 flex items-center justify-center"
      style={{
        width: size,
        height: size,
        boxShadow: `0 0 0 2px ${player?.color ?? "var(--border)"}`,
      }}
    >
      {getAvatar(
        player
          ? { id, name: player.name || "?", avatar: player.avatar }
          : { id, name: "?", avatar: null }
      )}
    </div>
  );
}

export function PlayerName({
  id,
  className = "",
}: {
  id?: string;
  className?: string;
}) {
  const { players } = useLog();
  if (!id) return null;
  if (isSystem(id)) {
    return <span className={`text-on-surface-muted ${className}`}>System</span>;
  }
  const player = players.get(id);
  return (
    <span
      className={`font-medium ${className}`}
      style={{ color: player?.color }}
    >
      {player?.name ?? "Unknown"}
    </span>
  );
}

export function PlayerBadge({ id }: { id: string }) {
  return (
    <span className="inline-flex items-center gap-1.5 align-middle">
      <PlayerAvatar id={id} size={18} />
      <PlayerName id={id} className="text-sm" />
    </span>
  );
}
