import { useRoomContext } from "@/contexts/RoomContext";
import { useRequiredUser } from "@/contexts/UserContext";
import BettingPanel from "./BettingPanel";
import FinalRoundAnswerPanel from "./FinalRoundAnswerPanel";
import { Player } from "@/types/user";

export default function GameBottomSection() {
  const user = useRequiredUser();
  const {
    room,
    placeBet,
    placeFinalRoundBet,
    submitFinalRoundAnswer,
    answerButton,
    submitAnswer,
  } = useRoomContext();

  const player = room.players.find((p) => p.id === user.id);
  if (!player) return null;

  const allowedToAnswer = room.allowedToAnswer?.includes(user.id) ?? false;

  switch (room.state) {
    case "betting":
    case "final_round_betting":
      return (
        <BettingPanel
          player={player}
          allowedToBet={player.score > 0 && !player.betAmount}
          placeBet={room.state === "betting" ? placeBet : placeFinalRoundBet}
        />
      );

    case "showing_final_round_question":
      return (
        <FinalRoundAnswerPanel
          allowedToAnswer={allowedToAnswer}
          submitFinalRoundAnswer={submitFinalRoundAnswer}
        />
      );

    case "validating_final_round_answers":
      return null;

    default:
      return (
        <div
          className="w-full h-12 rounded-md bg-primary text-on-primary hover:bg-primary-hover focus:outline-none transition-colors duration-150"
          tabIndex={-1}
          ref={answerButton}
          onClick={submitAnswer}
        />
      );
  }
}
