import React from "react";
import Modal from "../../Modal";
import { FinalRoundQuestion, Question } from "../../pack/PackEditor";

export default function ValidateAnswerModal({
  isOpen,
  question,
  playerAnswer,
  validateAnswer,
}: {
  isOpen: boolean;
  question?: Question | FinalRoundQuestion;
  playerAnswer?: string;
  validateAnswer?: (isCorrect: boolean) => void;
}) {
  return (
    <Modal isOpen={isOpen} onClose={() => {}} closeByClickingOutside={false}>
      <div className="min-w-48">
        <h3 className="text-base font-medium">Validate answer</h3>
        {playerAnswer && <p className="mt-2">Player answer: {playerAnswer}</p>}
        <p className="mt-2">Correct answers:</p>
        <ul className="px-2 list-inside list-disc">
          {question?.answers.map((answer, index) => (
            <li className="cursor-pointer" key={index}>
              {answer}
            </li>
          ))}
        </ul>
        {question?.comment && (
          <>
            <p className="mt-2">Comment:</p>
            <p>{question.comment}</p>
          </>
        )}
        <div className="mt-4 flex justify-between">
          <button
            className="rounded-lg py-1.5 px-3 text-base font-medium border text-red-600"
            onClick={validateAnswer?.bind(null, false)}
          >
            Incorrect
          </button>
          <button
            className="rounded-lg py-1.5 px-3 text-base font-medium border text-green-600"
            onClick={validateAnswer?.bind(null, true)}
          >
            Correct
          </button>
        </div>
      </div>
    </Modal>
  );
}
