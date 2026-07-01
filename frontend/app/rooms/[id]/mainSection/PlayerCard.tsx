import React, { useEffect, useState } from "react";
import { getAvatar } from "@/components/UserAvatar";
import { useRoomContext } from "@/contexts/RoomContext";
import { useRequiredUser } from "@/contexts/UserContext";
import { Player } from "@/types/user";

export default function PlayerCard({ player }: { player: Player }) {
  const user = useRequiredUser();
  const { room, passQuestion, changeScore, banPlayer } = useRoomContext();
  const isModerator = room.moderator?.id === user.id;
  const [editingScore, setEditingScore] = useState(false);
  const [scoreInput, setScoreInput] = useState("");
  const [menuPos, setMenuPos] = useState<{ x: number; y: number } | null>(null);

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
    (isModerator || user.id === room.currentPlayer) &&
    player.id !== room.currentPlayer;

  const startEditingScore = () => {
    setScoreInput(String(player.score));
    setEditingScore(true);
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

  const handleContextMenu = (e: React.MouseEvent) => {
    if (!isModerator) return;
    e.preventDefault();
    setMenuPos({ x: e.clientX, y: e.clientY });
  };

  useEffect(() => {
    if (!menuPos) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setMenuPos(null);
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [menuPos]);

  const borderCls = isOrange
    ? "border-2 border-orange-400"
    : isYellow
    ? "border-2 border-yellow-400"
    : "border-2 border-border";

  return (
    <>
      <div
        className={`min-w-6 max-w-16 flex-1 flex flex-col rounded-md overflow-hidden transition-all duration-150 ${borderCls}${
          player.isConnected ? "" : " opacity-40"
        }${canBePassedTo ? " cursor-pointer hover:opacity-75" : ""}`}
        onClick={canBePassedTo ? () => passQuestion(player.id) : undefined}
        onContextMenu={handleContextMenu}
      >
        <div className="w-full aspect-square relative">
          {getAvatar(player)}
          {hasBet && (
            <div className="absolute inset-0 flex items-center justify-center bg-black/50">
              <span className="text-white text-sm font-bold">✓</span>
            </div>
          )}
          {isBettingState && isYellow && (
            <div className="absolute inset-0 flex items-center justify-center">
              <span className="text-yellow-300 text-sm font-bold drop-shadow">
                ?
              </span>
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
          {isModerator && editingScore ? (
            <input
              className="w-full text-xs font-extrabold text-center bg-surface border border-primary rounded text-on-surface leading-tight outline-none px-0.5"
              value={scoreInput}
              onChange={(e) => setScoreInput(e.target.value)}
              onBlur={commitScore}
              onKeyDown={handleScoreKeyDown}
              onClick={(e) => e.stopPropagation()}
              autoFocus
            />
          ) : (
            <p
              className={`text-xs font-extrabold text-on-surface leading-tight border border-transparent${
                isModerator ? " cursor-pointer hover:text-primary" : ""
              }`}
              onDoubleClick={
                isModerator
                  ? (e) => {
                      e.stopPropagation();
                      startEditingScore();
                    }
                  : undefined
              }
              title={isModerator ? "Double-click to edit score" : undefined}
            >
              {player.score}
            </p>
          )}
          {isOrange && player.betAmount != null && (
            <p className="text-[10px] font-bold text-orange-400 leading-tight">
              {player.betAmount}
            </p>
          )}
        </div>
      </div>

      {menuPos && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setMenuPos(null)}
          />
          <div
            className="fixed z-50 bg-surface border border-border rounded-lg shadow-lg py-1 min-w-[140px]"
            style={{ top: menuPos.y, left: menuPos.x }}
          >
            <button
              className="w-full text-left px-3 py-2 text-sm hover:bg-surface-raised transition-colors"
              onClick={() => {
                startEditingScore();
                setMenuPos(null);
              }}
            >
              Change score
            </button>
            {player.id !== user.id && (
              <button
                className="w-full text-left px-3 py-2 text-sm text-danger hover:bg-surface-raised transition-colors"
                onClick={() => {
                  banPlayer(player.id);
                  setMenuPos(null);
                }}
              >
                Ban
              </button>
            )}
          </div>
        </>
      )}
    </>
  );
}
