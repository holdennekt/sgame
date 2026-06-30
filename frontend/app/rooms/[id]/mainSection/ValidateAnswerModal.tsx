import React from "react";
import { FinalRoundQuestion, Question } from "@/types/pack";
import { FiCheck, FiX } from "react-icons/fi";

export default function ValidateAnswerPanel({
  question,
  playerAnswer,
  validateAnswer,
  disabled = false,
}: {
  question?: Question | FinalRoundQuestion;
  playerAnswer?: string;
  validateAnswer?: (isCorrect: boolean) => void;
  disabled?: boolean;
}) {
  return (
    <div className="w-full flex items-center gap-3 bg-surface border border-border rounded-md px-3 py-2">
      <div className="flex-1 min-w-0">
        <p className="text-[10px] font-semibold uppercase tracking-widest text-on-surface-muted">
          Correct answers
        </p>
        <div className="flex flex-wrap gap-x-3 gap-y-0.5 mt-0.5">
          {question?.answers.map((answer, index) => (
            <span key={index} className="text-sm text-on-surface">
              {answer}
            </span>
          ))}
        </div>
        {playerAnswer && (
          <p className="text-xs text-on-surface-muted mt-1">
            Player: {playerAnswer}
          </p>
        )}
      </div>
      <div className="flex gap-2 shrink-0">
        <button
          className="inline-flex items-center justify-center gap-1 rounded-lg py-1.5 px-3 text-sm font-medium border border-border text-danger hover:bg-surface-raised transition-colors duration-150 disabled:opacity-50 disabled:cursor-default disabled:hover:bg-transparent"
          disabled={disabled}
          onClick={() => validateAnswer?.(false)}
        >
          <FiX size={13} />
          Incorrect
        </button>
        <button
          className="inline-flex items-center justify-center gap-1 rounded-lg py-1.5 px-3 text-sm font-medium bg-green-600 text-white hover:bg-green-700 transition-colors duration-150 disabled:opacity-50 disabled:cursor-default disabled:hover:bg-green-600"
          disabled={disabled}
          onClick={() => validateAnswer?.(true)}
        >
          <FiCheck size={13} />
          Correct
        </button>
      </div>
    </div>
  );
}
