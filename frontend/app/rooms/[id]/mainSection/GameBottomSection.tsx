import { useRef, useState } from "react";
import { useRoomContext } from "@/contexts/RoomContext";
import { useRequiredUser } from "@/contexts/UserContext";
import BettingPanel from "./BettingPanel";
import FinalRoundAnswerPanel from "./FinalRoundAnswerPanel";

export default function GameBottomSection() {
  const user = useRequiredUser();
  const {
    room,
    placeBet,
    placeFinalRoundBet,
    submitFinalRoundAnswer,
    submitTypedAnswer,
    answerButton,
    startAnswer,
  } = useRoomContext();
  const [typedAnswer, setTypedAnswer] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  const player = room.players.find((p) => p.id === user.id);
  if (!player) return null;

  const allowedToAnswer = room.allowedToAnswer?.includes(user.id) ?? false;

  if (
    room.state === "answering" &&
    room.options.aiHost &&
    room.answeringPlayer?.id === user.id
  ) {
    if (room.answeringPlayer.answer) {
      return (
        <div className="w-full h-12 flex items-center justify-center gap-2 rounded-md bg-surface border border-border text-sm text-on-surface-muted">
          <span className="animate-pulse">Validating answer…</span>
        </div>
      );
    }

    const handleSubmit = () => {
      const answer = typedAnswer.trim();
      if (!answer) return;
      submitTypedAnswer(answer);
      setTypedAnswer("");
    };

    return (
      <div className="w-full flex gap-2">
        <input
          ref={inputRef}
          autoFocus
          className="flex-1 h-12 px-3 rounded-md bg-surface border border-border text-sm text-on-surface outline-none focus:border-primary transition-colors"
          placeholder="Type your answer…"
          value={typedAnswer}
          onChange={(e) => setTypedAnswer(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") handleSubmit();
          }}
        />
        <button
          className="h-12 px-4 rounded-md bg-primary text-on-primary text-sm font-medium hover:bg-primary-hover transition-colors disabled:opacity-50 disabled:cursor-default disabled:hover:bg-primary"
          onClick={handleSubmit}
          disabled={!typedAnswer.trim()}
        >
          Submit
        </button>
      </div>
    );
  }

  switch (room.state) {
    case "betting":
    case "final_round_betting":
      return (
        <BettingPanel
          player={player}
          allowedToBet={
            !room.pausedState.paused && player.score > 0 && !player.betAmount
          }
          placeBet={room.state === "betting" ? placeBet : placeFinalRoundBet}
        />
      );

    case "showing_final_round_question":
      return (
        <FinalRoundAnswerPanel
          allowedToAnswer={!room.pausedState.paused && allowedToAnswer}
          submitFinalRoundAnswer={submitFinalRoundAnswer}
        />
      );

    case "validating_final_round_answers":
      return null;

    default: {
      const questionOpen =
        room.state === "revealing_question" ||
        room.state === "showing_question";
      const canAnswer =
        questionOpen && allowedToAnswer && !room.pausedState.paused;
      const label = canAnswer
        ? "Answer"
        : room.pausedState.paused
          ? "Game paused"
          : questionOpen
            ? "You can't answer this question"
            : "Waiting for a question";

      return (
        <div
          role="button"
          aria-disabled={!canAnswer}
          aria-label={canAnswer ? "Answer (Space)" : label}
          className={`w-full h-12 flex items-center justify-center gap-2 rounded-md text-sm font-medium focus:outline-none transition-[background-color,transform,filter] duration-150 ${
            canAnswer
              ? "bg-primary text-on-primary hover:bg-primary-hover focus:brightness-75 focus:scale-[0.99] cursor-pointer"
              : "bg-surface-raised text-on-surface-muted cursor-default"
          }`}
          tabIndex={-1}
          ref={answerButton}
          onClick={startAnswer}
        >
          {label}
          {canAnswer && (
            <kbd className="px-1.5 py-0.5 rounded border border-on-primary/40 text-xs font-normal">
              Space
            </kbd>
          )}
        </div>
      );
    }
  }
}
