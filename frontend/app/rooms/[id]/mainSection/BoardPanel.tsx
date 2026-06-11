import { CategoryQuestions } from "@/types/room";
import React from "react";

export default function BoardPanel({
  currentRoundQuestions,
  selectQuestion,
  canSelectQuestion,
}: {
  currentRoundQuestions: CategoryQuestions[];
  selectQuestion: (question: { category: string; index: number }) => void;
  canSelectQuestion: boolean;
}) {
  const questionsInCategoryCount =
    currentRoundQuestions[0]?.questions.length ?? 0;
  if (currentRoundQuestions.length === 0 || questionsInCategoryCount === 0)
    return null;

  const n = questionsInCategoryCount + 1;

  return (
    <div
      className="w-full h-full grid gap-0.5 sm:gap-1 [grid-template-columns:2fr_repeat(var(--q),minmax(0,1fr))] auto-rows-fr sm:grid-cols-none sm:[grid-template-rows:repeat(var(--n),minmax(0,1fr))] sm:[grid-auto-flow:column] sm:auto-cols-fr"
      style={
        {
          "--n": String(n),
          "--q": String(questionsInCategoryCount),
        } as React.CSSProperties
      }
    >
      {currentRoundQuestions.map(({ category, questions }) => (
        <React.Fragment key={category}>
          <div className="flex items-center justify-center p-1 sm:p-2 rounded-lg font-bold uppercase tracking-wide bg-primary text-on-primary">
            <span className="text-[10px] sm:text-xs text-center break-words w-full min-w-0">
              {category}
            </span>
          </div>
          {Array.from({ length: questionsInCategoryCount }, (_, i) => {
            const q = questions[i];
            return !q || q.hasBeenPlayed ? (
              <div
                key={i}
                className="rounded-lg border border-border opacity-40"
              />
            ) : (
              <button
                key={i}
                className={`w-full h-full rounded-lg flex items-center justify-center text-xs sm:text-base font-semibold bg-primary text-on-primary transition-opacity duration-150 ${
                  canSelectQuestion
                    ? "cursor-pointer hover:bg-primary-hover"
                    : "cursor-default"
                }`}
                onClick={
                  canSelectQuestion
                    ? () => selectQuestion({ category, index: i })
                    : undefined
                }
              >
                {q.value}
              </button>
            );
          })}
        </React.Fragment>
      ))}
    </div>
  );
}
