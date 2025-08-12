import { User } from "@/middleware";
import BettingPanel from "./BettingPanel";
import BoardPanel from "./BoardPanel";
import FinalRoundBoardPanel from "./FinalRoundBoardPanel";
import PlayerTable from "./PlayerTable";
import {
  CorrectAnswerDemo,
  CurrentQuestion,
  FinalRoundState,
  isCorrectAnswerDemo,
  isRoundDemo,
  Room,
  RoundDemo,
} from "../Room";
import RoundDemoPanel from "./RoundDemoPanel";
import {
  MutableRefObject,
  RefObject,
  useEffect,
  useRef,
  useState,
} from "react";
import { WsMessageHandler } from "../../lobby/Lobby";
import ValidateAnswerModal from "./ValidateAnswerModal";
import { FinalRoundQuestion } from "../../pack/PackEditor";
import TextPanel from "./TextPanel";
import FinalRoundAnswerPanel from "./FinalRoundAnswerPanel";
import QuestionPanel from "./QuestionPanel";

export default function MainPanel({
  user,
  room,
  wsConn,
  handlers,
  answerButton,
  submitAnswer,
}: {
  user: User;
  room: Room;
  wsConn: MutableRefObject<WebSocket | null>;
  handlers: Map<string, WsMessageHandler>;
  answerButton: RefObject<HTMLDivElement>;
  submitAnswer: () => void;
}) {
  const [roundDemo, setRoundDemo] = useState<RoundDemo | null>(null);
  const delayedRoundDemoRef = useRef<RoundDemo | null>(null);
  const [correctAnswerDemo, setCorrectAnswerDemo] =
    useState<CorrectAnswerDemo | null>(null);

  handlers.set("round_demo", (payload) => {
    if (!isRoundDemo(payload)) return;
    if (correctAnswerDemo) {
      delayedRoundDemoRef.current = payload;
      return;
    }
    setRoundDemo(payload);
  });

  handlers.set("correct_answer_demo", (payload) => {
    if (!isCorrectAnswerDemo(payload)) return;
    setCorrectAnswerDemo(payload);
    setTimeout(() => setCorrectAnswerDemo(null), payload.duration * 1000);
  });

  useEffect(() => {
    if (!correctAnswerDemo && delayedRoundDemoRef.current) {
      setRoundDemo(delayedRoundDemoRef.current);
      delayedRoundDemoRef.current = null;
    }
  }, [correctAnswerDemo]);

  const isHost = user.id === room.host?.id;
  const player = room.players.find((player) => player.id === user.id);

  const selectQuestion = (question: { category: string; index: number }) => {
    wsConn.current?.send(
      JSON.stringify({ event: "select_question", payload: question })
    );
  };

  const passQuestion = (passTo: string) => {
    wsConn.current?.send(
      JSON.stringify({ event: "pass_question", payload: { passTo } })
    );
  };

  const placeBet = (amount: number) => {
    wsConn.current?.send(
      JSON.stringify({ event: "place_bet", payload: { amount } })
    );
  };

  const validateAnswer = (isCorrect: boolean) => {
    wsConn.current?.send(
      JSON.stringify({ event: "validate_answer", payload: { isCorrect } })
    );
  };

  const removeFinalRoundCategory = (category: string) => {
    wsConn.current?.send(
      JSON.stringify({
        event: "remove_final_round_category",
        payload: { category },
      })
    );
  };

  const placeFinalRoundBet = (amount: number) => {
    wsConn.current?.send(
      JSON.stringify({ event: "place_final_round_bet", payload: { amount } })
    );
  };

  const submitFinalRoundAnswer = (answer: string) => {
    wsConn.current?.send(
      JSON.stringify({
        event: "submit_final_round_answer",
        payload: { answer },
      })
    );
  };

  const validateFinalRoundAnswer = (isCorrect: boolean) => {
    wsConn.current?.send(
      JSON.stringify({
        event: "validate_final_round_answer",
        payload: { isCorrect },
      })
    );
  };

  const getMainTopSection = (room: Room) => {
    switch (room.state) {
      case "waiting_for_start":
        return <TextPanel topText="Waiting for start" />;
      case "selecting_question":
        if (correctAnswerDemo)
          return (
            <TextPanel
              topText={correctAnswerDemo.answers.join(", ")}
              bottomText={correctAnswerDemo.comment}
            />
          );
        if (roundDemo)
          return (
            <RoundDemoPanel
              roundDemo={roundDemo}
              onFinish={() => setRoundDemo(null)}
            />
          );
        return (
          <BoardPanel
            currentRoundQuestions={room.currentRoundQuestions!}
            selectQuestion={selectQuestion}
            canSelectQuestion={isHost || user.id === room.currentPlayer}
          />
        );

      case "revealing_question":
        return <QuestionPanel text={room.currentQuestion?.text!} />;

      case "showing_question":
        return (
          <QuestionPanel
            text={room.currentQuestion?.text!}
            timeBar={{
              progress: room.currentQuestion?.timerLastProgress!,
              durationMs:
                room.currentQuestion!.timerEndsAt.getTime() - Date.now(),
            }}
          />
        );
      case "answering":
        return (
          <QuestionPanel
            text={room.currentQuestion?.text!}
            timeBar={{
              progress: 1,
              durationMs:
                room.answeringPlayer!.timerEndsAt.getTime() - Date.now(),
            }}
          />
        );

      case "passing":
        return (
          <TextPanel
            topText="Cat in bag!"
            bottomText="Pass the question to another player"
          />
        );

      case "betting":
        return <TextPanel topText="Auction!" bottomText="Place your bet" />;

      case "selecting_final_round_category":
        return (
          <FinalRoundBoardPanel
            availableCategories={room.finalRoundState?.availableCategories!}
            canRemoveCategory={user.id === room.currentPlayer || isHost}
            removeCategory={removeFinalRoundCategory}
          />
        );

      case "final_round_betting":
        return <TextPanel topText="Final round!" bottomText="Place your bet" />;

      case "showing_final_round_question":
        return (
          <QuestionPanel
            text={room.finalRoundState!.question!.text}
            timeBar={{
              progress: 1,
              durationMs:
                room.finalRoundState!.timerEndsAt!.getTime() - Date.now(),
            }}
          />
        );

      case "validating_final_round_answers":
        return <TextPanel topText={room.finalRoundState?.question?.text!} />;

      case "game_over":
        if (correctAnswerDemo)
          return (
            <TextPanel
              topText={correctAnswerDemo.answers.join(", ")}
              bottomText={correctAnswerDemo.comment}
            />
          );
        return <TextPanel topText="Game over!" />;

      default:
        return <TextPanel topText={`Unexpected room state: ${room.state}`} />;
    }
  };

  const getMainBottomSection = (room: Room) => {
    switch (room.state) {
      case "betting":
      case "final_round_betting":
        return (
          <BettingPanel
            player={player!}
            allowedToBet={player!.score > 0 && !player!.betAmount}
            placeBet={room.state === "betting" ? placeBet : placeFinalRoundBet}
          />
        );
      case "showing_final_round_question":
        return (
          <FinalRoundAnswerPanel
            allowedToAnswer={room.allowedToAnswer.includes(user.id)}
            submitFinalRoundAnswer={submitFinalRoundAnswer}
          />
        );

      case "validating_final_round_answers":
        return;

      default:
        return (
          <div
            className="w-full h-12 rounded primary hover:opacity-85 focus:outline-none focus:opacity-85"
            tabIndex={-1}
            ref={answerButton}
            onClick={submitAnswer}
          ></div>
        );
    }
  };

  return (
    <>
      <div className="flex-[3_0_0%] flex flex-col gap-2 min-w-0 min-h-0">
        <div className="flex-1 w-full rounded surface p-2">
          {getMainTopSection(room)}
        </div>
        {room.players.length > 0 && (
          <div
            className={`w-full flex flex-wrap justify-around gap-3 border rounded p-3`}
          >
            {room.players.map((p, index) => (
              <PlayerTable
                key={index}
                user={user}
                room={room}
                player={p}
                passQuestion={passQuestion}
              />
            ))}
          </div>
        )}
        {player && getMainBottomSection(room)}
      </div>
      {isHost && (
        <ValidateAnswerModal
          isOpen={
            room.state === "answering" ||
            room.state === "validating_final_round_answers"
          }
          question={
            {
              answering: room.currentQuestion as CurrentQuestion,
              validating_final_round_answers: room.finalRoundState
                ?.question as FinalRoundQuestion,
            }[room.state as string]
          }
          playerAnswer={
            room.state === "validating_final_round_answers"
              ? (room.finalRoundState as FinalRoundState)?.playersAnswers![
                  room.currentPlayer!
                ]
              : undefined
          }
          validateAnswer={
            {
              answering: validateAnswer,
              validating_final_round_answers: validateFinalRoundAnswer,
            }[room.state as string]
          }
        />
      )}
    </>
  );
}
