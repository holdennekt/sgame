import React from "react";
import { getAvatar } from "../../lobby/Room";
import { User } from "@/middleware";
import { Room } from "../Room";

export type Player = User & {
  score: number;
  betAmount: number | null;
  isConnected: boolean;
};

export default function PlayerTable({
  user,
  room,
  player,
  passQuestion,
}: {
  user: User;
  room: Room;
  player: Player;
  passQuestion: (passTo: string) => void;
}) {
  const borderClass = (() => {
    if (room.state === "selecting_question") {
      if (player.id === room.currentPlayer) return "border-2 border-yellow-400";
    }

    if (room.state === "answering") {
      if (player.id === room.answeringPlayer?.id)
        return "border-2 border-orange-600";
    }

    if (room.state === "passing") {
      if (player.id !== room.currentPlayer && player.isConnected)
        return "border-2 border-yellow-400";
    }

    if (room.state === "betting" || room.state === "final_round_betting") {
      if (player.score > 0 && player.betAmount === null)
        return "border-2 border-yellow-400";
    }

    if (room.state === "validating_final_round_answers") {
      if (player.id === room.currentPlayer)
        return "border-2 border-yellow-400";
    }

    return "border border-white";
  })();

  const canBePassedTo =
    (user.id === room.host?.id || user.id === room.currentPlayer) &&
    player.id !== room.currentPlayer;

  return (
    <div
      className={`min-w-8 max-w-24 flex-1 border-collapse${
        player.isConnected ? "" : " opacity-50"
      }${canBePassedTo ? " hover:opacity-75" : ""}`}
      onClick={canBePassedTo ? () => passQuestion(player.id) : undefined}
    >
      <div className={`w-full aspect-square ${borderClass}`}>
        {getAvatar(player)}
      </div>
      <div className={`w-full ${borderClass} rounded-b`}>
        <p
          className="w-full text-center text-base truncate hidden md:block px-1"
          title={player.name}
        >
          {player.name}
        </p>
        <p className="w-full text-center text-sm font-extrabold truncate">
          {player.score}
        </p>
        {room.state === "answering" &&
          player.id === room.answeringPlayer?.id &&
          player.betAmount && (
            <p className="w-full text-center text-sm font-bold truncate">
              {player.betAmount}
            </p>
          )}
      </div>
    </div>
  );
}
