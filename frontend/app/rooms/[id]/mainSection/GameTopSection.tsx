import { useRoomContext } from "@/contexts/RoomContext";
import { useRequiredUser } from "@/contexts/UserContext";
import { RoomHost, RoomPlayer } from "@/types/room";
import BoardPanel from "./BoardPanel";
import CorrectAnswerDemoPanel from "./CorrectAnswerDemoPanel";
import FinalRoundBoardPanel from "./FinalRoundBoardPanel";
import QuestionPanel from "./QuestionPanel";
import RevealingQuestionPanel from "./RevealingQuestionPanel";
import RoundDemoPanel from "./RoundDemoPanel";
import TextPanel from "./TextPanel";

export default function GameTopSection() {
  const user = useRequiredUser();
  const {
    room,
    roundDemo,
    onRoundDemoDone,
    questionDemo,
    correctAnswerDemo,
    selectQuestion,
    removeFinalRoundCategory,
  } = useRoomContext();

  const isHost = user.id === room.host?.id;

  if (questionDemo)
    return (
      <TextPanel
        topText={`${questionDemo.category}, ${questionDemo.value}`}
        bottomText={
          { regular: "Regular", catInBag: "Cat in bag!", auction: "Auction!" }[
            questionDemo.type
          ]
        }
        commentText={questionDemo.categoryComment}
      />
    );
  if (correctAnswerDemo)
    return <CorrectAnswerDemoPanel correctAnswerDemo={correctAnswerDemo} />;
  if (roundDemo)
    return <RoundDemoPanel roundDemo={roundDemo} onFinish={onRoundDemoDone} />;

  switch (room.state) {
    case "waiting_for_start":
      return <TextPanel topText="Waiting for start" />;

    case "selecting_question":
      return (
        <BoardPanel
          currentRoundQuestions={room.currentRoundQuestions!}
          selectQuestion={selectQuestion}
          canSelectQuestion={isHost || user.id === room.currentPlayer}
        />
      );

    case "revealing_question":
      return (
        <RevealingQuestionPanel
          attachment={room.currentQuestion!.attachment}
          attachmentEndsAt={room.currentQuestion!.attachmentRevealEndsAt}
          attachmentLastProgress={
            room.currentQuestion!.attachmentRevealLastProgress
          }
          text={room.currentQuestion!.text}
          textEndsAt={room.currentQuestion!.timerStartsAt}
          textLastProgress={room.currentQuestion!.textRevealLastProgress}
          paused={room.pausedState.paused}
          category={room.currentQuestion!.category}
          value={room.currentQuestion!.value}
        />
      );

    case "showing_question":
      return (
        <QuestionPanel
          attachment={room.currentQuestion!.attachment}
          attachmentLastProgress={
            room.currentQuestion!.attachmentRevealLastProgress
          }
          text={room.currentQuestion!.text}
          textLastProgress={room.currentQuestion!.textRevealLastProgress}
          questionType={room.currentQuestion!.type}
          timeBar={{
            progress: room.currentQuestion!.timerLastProgress,
            endsAt: new Date(room.currentQuestion!.timerEndsAt).getTime(),
            paused: room.pausedState.paused,
          }}
          category={room.currentQuestion!.category}
          value={room.currentQuestion!.value}
        />
      );

    case "answering":
      return (
        <QuestionPanel
          attachment={room.currentQuestion!.attachment}
          attachmentLastProgress={
            room.currentQuestion!.attachmentRevealLastProgress
          }
          text={room.currentQuestion!.text}
          textLastProgress={room.currentQuestion!.textRevealLastProgress}
          questionType={room.currentQuestion!.type}
          timeBar={{
            progress: 1,
            endsAt: new Date(room.answeringPlayer!.timerEndsAt).getTime(),
            paused: room.pausedState.paused,
          }}
          category={room.currentQuestion!.category}
          value={room.currentQuestion!.value}
        />
      );

    case "passing":
      return (
        <TextPanel
          topText="Cat in bag!"
          bottomText="Pass the question to another player"
          timeBar={{
            progress: 1,
            endsAt: new Date(room.currentQuestion!.passingEndsAt).getTime(),
            paused: room.pausedState.paused,
          }}
        />
      );

    case "betting":
      return (
        <TextPanel
          topText="Auction!"
          bottomText="Place your bet"
          timeBar={{
            progress: 1,
            endsAt: new Date(room.currentQuestion!.bettingEndsAt).getTime(),
            paused: room.pausedState.paused,
          }}
        />
      );

    case "selecting_final_round_category":
      return (
        <FinalRoundBoardPanel
          availableCategories={room.finalRoundState?.availableCategories!}
          canRemoveCategory={user.id === room.currentPlayer || isHost}
          removeCategory={removeFinalRoundCategory}
        />
      );

    case "final_round_betting":
      return (
        <TextPanel
          topText="Final round!"
          bottomText={room.finalRoundState!.question!.category}
          commentText="Place your bet"
        />
      );

    case "showing_final_round_question":
      return (
        <QuestionPanel
          attachment={room.finalRoundState!.question!.attachment}
          attachmentLastProgress={0}
          text={room.finalRoundState!.question!.text}
          textLastProgress={1}
          questionType="final"
          timeBar={{
            progress: 1,
            endsAt: new Date(room.finalRoundState!.timerEndsAt!).getTime(),
            paused: room.pausedState.paused,
          }}
          category={room.finalRoundState!.question!.category}
        />
      );

    case "validating_final_round_answers":
      return (
        <QuestionPanel
          attachment={room.finalRoundState!.question!.attachment}
          attachmentLastProgress={1}
          text={room.finalRoundState!.question!.text}
          textLastProgress={1}
          questionType="regular"
          timeBar={{
            progress: 0,
            endsAt: 0,
          }}
          category={room.finalRoundState!.question!.category}
        />
      );

    case "game_over":
      return <TextPanel topText="Game over!" />;

    default:
      return <TextPanel topText={`Unexpected room state: ${room.state}`} />;
  }
}
