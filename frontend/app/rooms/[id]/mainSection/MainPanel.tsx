"use client";

import { CurrentQuestion, FinalRoundState } from "@/types/room";
import { FinalRoundQuestion } from "@/types/pack";
import { useRoomContext } from "@/contexts/RoomContext";
import { useRequiredUser } from "@/contexts/UserContext";
import PlayerCard from "./PlayerCard";
import GameTopSection from "./GameTopSection";
import GameBottomSection from "./GameBottomSection";
import ValidateAnswerPanel from "./ValidateAnswerModal";
import { FiSkipForward, FiPause } from "react-icons/fi";

const SKIPPABLE_STATES = new Set([
  "revealing_question",
  "showing_question",
  "answering",
  "passing",
  "betting",
]);

export default function MainPanel() {
  const user = useRequiredUser();
  const {
    room,
    validateAnswer,
    validateFinalRoundAnswer,
    skipQuestion,
    skipRound,
  } = useRoomContext();
  const isHost = user.id === room.host?.id;
  const player = room.players.find((p) => p.id === user.id);
  const canSkip = isHost && SKIPPABLE_STATES.has(room.state);
  const canSkipRound = isHost && room.state === "selecting_question";

  return (
    <div className="flex-[3_0_0%] flex flex-col gap-2 min-w-0 min-h-0">
      <div className="relative flex-1 min-h-0 w-full bg-surface border border-border rounded-md p-2">
        {room.pausedState.paused && (
          <div className="absolute top-0 inset-x-0 z-10 flex items-center justify-center gap-1.5 py-1 rounded-t-md bg-amber-400/15 text-amber-500 text-xs font-semibold tracking-wide">
            <FiPause size={11} />
            PAUSED
          </div>
        )}
        <GameTopSection />
      </div>

      {room.players.length > 0 && (
        <div className="w-full flex flex-wrap justify-around gap-2 bg-surface border border-border rounded-md p-2">
          {room.players.map((p, index) => (
            <PlayerCard key={index} player={p} />
          ))}
        </div>
      )}

      {player && <GameBottomSection />}

      {isHost &&
        (room.state === "answering" ||
          room.state === "validating_final_round_answers") && (
          <ValidateAnswerPanel
            disabled={room.pausedState.paused}
            question={
              room.state === "answering"
                ? (room.currentQuestion as CurrentQuestion)
                : (room.finalRoundState?.question as FinalRoundQuestion)
            }
            playerAnswer={
              room.state === "validating_final_round_answers"
                ? (room.finalRoundState as FinalRoundState)?.playersAnswers[
                    room.currentPlayer!
                  ]
                : undefined
            }
            validateAnswer={
              room.state === "answering"
                ? validateAnswer
                : validateFinalRoundAnswer
            }
          />
        )}

      {canSkip && (
        <button
          className="w-full h-12 inline-flex items-center justify-center gap-1.5 rounded-md text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 disabled:opacity-50 disabled:cursor-default disabled:hover:bg-primary"
          disabled={room.pausedState.paused}
          onClick={skipQuestion}
        >
          <FiSkipForward size={12} />
          Skip question
        </button>
      )}

      {canSkipRound && (
        <button
          className="w-full h-12 inline-flex items-center justify-center gap-1.5 rounded-md text-sm font-medium bg-surface border border-border text-on-surface hover:bg-surface-raised transition-colors duration-150 disabled:opacity-50 disabled:cursor-default disabled:hover:bg-surface"
          disabled={room.pausedState.paused}
          onClick={skipRound}
        >
          <FiSkipForward size={12} />
          Skip round
        </button>
      )}
    </div>
  );
}
