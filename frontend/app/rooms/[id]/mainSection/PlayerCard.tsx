import React, { useRef, useState } from "react";
import { getAvatar } from "@/components/UserAvatar";
import { useRoomContext } from "@/contexts/RoomContext";
import { useRequiredUser } from "@/contexts/UserContext";
import { Player } from "@/types/user";

export default function PlayerCard({ player }: { player: Player }) {
  const user = useRequiredUser();
  const { room, passQuestion, changeScore } = useRoomContext();
  const isHost = user.id === room.host?.id;
  const [editingScore, setEditingScore] = useState(false);
  const [scoreInput, setScoreInput] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const isYellow =
    (room.state === "selecting_question" && player.id === room.currentPlayer) ||
    (room.state === "selecting_final_round_category" &&
      player.id === room.currentPlayer) ||
    (room.state === "passing" &&
      player.id !== room.currentPlayer &&
      player.isConnected) ||
    ((room.state === "betting" || room.state === "final_round_betting") &&
      player.score > 0 &&
      player.betAmount === null) ||
    (room.state === "validating_final_round_answers" &&
      player.id === room.currentPlayer);

  const isOrange =
    room.state === "answering" && player.id === room.answeringPlayer?.id;

  const isBettingState =
    room.state === "betting" || room.state === "final_round_betting";
  const hasBet = isBettingState && player.betAmount !== null;

  const canBePassedTo =
    room.state === "passing" &&
    (isHost || user.id === room.currentPlayer) &&
    player.id !== room.currentPlayer;

  const startEditingScore = (e: React.MouseEvent) => {
    e.stopPropagation();
    setScoreInput(String(player.score));
    setEditingScore(true);
    setTimeout(() => {
      inputRef.current?.select();
    }, 0);
  };

  const commitScore = () => {
    const parsed = parseInt(scoreInput, 10);
    if (!isNaN(parsed) && parsed !== player.score) {
      changeScore(player.id, parsed);
    }
    setEditingScore(false);
  };

  const handleScoreKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") commitScore();
    if (e.key === "Escape") setEditingScore(false);
  };

  const borderCls = isOrange
    ? "border-2 border-orange-400"
    : isYellow
    ? "border-2 border-yellow-400"
    : "border-2 border-border";

  return (
    <div
      className={`min-w-6 max-w-16 flex-1 flex flex-col rounded-md overflow-hidden transition-all duration-150 ${borderCls}${
        player.isConnected ? "" : " opacity-40"
      }${canBePassedTo ? " cursor-pointer hover:opacity-75" : ""}`}
      onClick={canBePassedTo ? () => passQuestion(player.id) : undefined}
    >
      <div className="w-full aspect-square relative">
        {getAvatar(player)}
        {hasBet && (
          <div className="absolute inset-0 flex items-center justify-center bg-black/50">
            <span className="text-white text-sm font-bold">✓</span>
          </div>
        )}
      </div>
      <div className="w-full px-1 py-0.5 text-center bg-surface-raised">
        <p
          className="text-[10px] truncate text-on-surface-muted leading-tight"
          title={player.name}
        >
          {player.name}
        </p>
        {isHost && editingScore ? (
          <input
            ref={inputRef}
            className="w-full text-xs font-extrabold text-center bg-surface border border-primary rounded text-on-surface leading-tight outline-none px-0.5"
            value={scoreInput}
            onChange={(e) => setScoreInput(e.target.value)}
            onBlur={commitScore}
            onKeyDown={handleScoreKeyDown}
            onClick={(e) => e.stopPropagation()}
          />
        ) : (
          <p
            className={`text-xs font-extrabold text-on-surface leading-tight${
              isHost ? " cursor-pointer hover:text-primary" : ""
            }`}
            onDoubleClick={isHost ? startEditingScore : undefined}
            title={isHost ? "Double-click to edit score" : undefined}
          >
            {player.score}
          </p>
        )}
        {isOrange && player.betAmount && (
          <p className="text-[10px] font-bold text-orange-400 leading-tight">
            {player.betAmount}
          </p>
        )}
      </div>
    </div>
  );
}
