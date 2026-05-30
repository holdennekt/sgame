import React from "react";
import Modal from "@/components/Modal";
import { FinalRoundQuestion, Question } from "@/types/pack";
import { FiCheck, FiX } from "react-icons/fi";

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
      <div className="w-72 flex flex-col gap-4">
        <h3 className="text-sm font-semibold uppercase tracking-widest text-on-surface-muted">
          Validate answer
        </h3>

        {playerAnswer && (
          <div className="rounded-lg border border-border bg-surface-raised px-3 py-2">
            <p className="text-[10px] font-semibold uppercase tracking-widest text-on-surface-muted mb-1">
              Player's answer
            </p>
            <p className="text-sm text-on-surface">{playerAnswer}</p>
          </div>
        )}

        <div>
          <p className="text-[10px] font-semibold uppercase tracking-widest text-on-surface-muted mb-2">
            Correct answers
          </p>
          <ul className="flex flex-col gap-1">
            {question?.answers.map((answer, index) => (
              <li
                key={index}
                className="flex items-center gap-2 rounded-md px-3 py-1.5 bg-surface-raised text-sm text-on-surface"
              >
                <FiCheck size={12} className="text-green-500 shrink-0" />
                {answer}
              </li>
            ))}
          </ul>
        </div>

        <div className="flex gap-2 pt-1">
          <button
            className="flex-1 inline-flex items-center justify-center gap-1.5 rounded-lg py-2 px-3 text-sm font-medium border border-border text-danger hover:bg-surface-raised transition-colors duration-150"
            onClick={validateAnswer?.bind(null, false)}
          >
            <FiX size={14} />
            Incorrect
          </button>
          <button
            className="flex-1 inline-flex items-center justify-center gap-1.5 rounded-lg py-2 px-3 text-sm font-medium bg-green-600 text-white hover:bg-green-700 transition-colors duration-150"
            onClick={validateAnswer?.bind(null, true)}
          >
            <FiCheck size={14} />
            Correct
          </button>
        </div>
      </div>
    </Modal>
  );
}
