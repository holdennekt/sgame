"use client";

import { CurrentQuestion, FinalRoundState } from "@/types/room";
import { FinalRoundQuestion } from "@/types/pack";
import { useRoomContext } from "@/contexts/RoomContext";
import { useRequiredUser } from "@/contexts/UserContext";
import PlayerCard from "./PlayerCard";
import GameTopSection from "./GameTopSection";
import GameBottomSection from "./GameBottomSection";
import ValidateAnswerModal from "./ValidateAnswerModal";
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
    <>
      <div className="flex-[3_0_0%] flex flex-col gap-2 min-w-0 min-h-0">
        <div className="relative flex-1 min-h-0 w-full bg-surface border border-border rounded-md p-2">
          <GameTopSection />
          {room.pausedState.paused && (
            <div className="absolute inset-0 flex items-center justify-center bg-background/60 backdrop-blur-sm rounded-md z-10">
              <div className="flex flex-col items-center gap-2 text-on-surface">
                <FiPause size={32} />
                <p className="text-lg font-semibold">Paused</p>
              </div>
            </div>
          )}
        </div>

        {room.players.length > 0 && (
          <div className="w-full flex flex-wrap justify-around gap-2 bg-surface border border-border rounded-md p-2">
            {room.players.map((p, index) => (
              <PlayerCard key={index} player={p} />
            ))}
          </div>
        )}

        {player && <GameBottomSection />}

        {canSkip && (
          <button
            className="w-full h-12 inline-flex items-center justify-center gap-1.5 rounded-md text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
            onClick={skipQuestion}
          >
            <FiSkipForward size={12} />
            Skip question
          </button>
        )}

        {canSkipRound && (
          <button
            className="w-full h-12 inline-flex items-center justify-center gap-1.5 rounded-md text-sm font-medium bg-surface border border-border text-on-surface hover:bg-surface-raised transition-colors duration-150"
            onClick={skipRound}
          >
            <FiSkipForward size={12} />
            Skip round
          </button>
        )}
      </div>

      {isHost && (
        <ValidateAnswerModal
          isOpen={
            room.state === "answering" ||
            room.state === "validating_final_round_answers"
          }
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
    </>
  );
}
