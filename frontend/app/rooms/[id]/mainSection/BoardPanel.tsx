import { CategoryQuestions } from "@/types/room";
import React from "react";

type BoardQuestionView = {
  value: number;
  hasBeenPlayed: boolean;
  onClick: () => void;
};

export default function BoardPanel({
  currentRoundQuestions,
  selectQuestion,
  canSelectQuestion,
}: {
  currentRoundQuestions: CategoryQuestions[];
  selectQuestion: (question: { category: string; index: number }) => void;
  canSelectQuestion: boolean;
}) {
  const categoriesCount = currentRoundQuestions.length;
  const questionsInCategoryCount = currentRoundQuestions[0]?.questions.length ?? 0;
  if (categoriesCount === 0 || questionsInCategoryCount === 0) return null;

  const tableData: BoardQuestionView[][] = new Array<BoardQuestionView[]>(questionsInCategoryCount)
    .fill([])
    .map(() =>
      new Array<BoardQuestionView>(categoriesCount).fill({
        value: 0,
        hasBeenPlayed: true,
        onClick: () => {},
      })
    );

  for (const [categoryIndex, { category, questions }] of currentRoundQuestions.entries()) {
    for (const question of questions) {
      tableData[question.index][categoryIndex] = {
        value: question.value,
        hasBeenPlayed: question.hasBeenPlayed,
        onClick: () => selectQuestion({ category, index: question.index }),
      };
    }
  }

  const cols = { gridTemplateColumns: `repeat(${categoriesCount}, minmax(0, 1fr))` };

  return (
    <div className="w-full h-full flex flex-col gap-1">
      <div className="grid gap-1 shrink-0" style={cols}>
        {currentRoundQuestions.map(({ category }, index) => (
          <div
            key={index}
            className="flex items-center justify-center p-2 min-h-12 rounded-lg text-xs font-bold uppercase tracking-wide bg-primary text-on-primary"
          >
            <span className="text-center break-words w-full min-w-0">{category}</span>
          </div>
        ))}
      </div>

      <div className="flex-1 flex flex-col gap-1 min-h-0">
        {tableData.map((row, i) => (
          <div key={i} className="flex-1 grid gap-1" style={cols}>
            {row.map(({ value, hasBeenPlayed, onClick }, j) =>
              hasBeenPlayed ? (
                <div key={j} className="rounded-lg border border-border opacity-40" />
              ) : (
                <button
                  key={j}
                  className={`w-full h-full rounded-lg flex items-center justify-center text-base font-semibold bg-primary text-on-primary transition-opacity duration-150 ${
                    canSelectQuestion ? "cursor-pointer hover:bg-primary-hover" : "cursor-default"
                  }`}
                  onClick={canSelectQuestion ? onClick : undefined}
                >
                  {value}
                </button>
              )
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
