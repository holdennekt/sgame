import React from "react";
import { BoardQuestion } from "../Room";

export default function BoardPanel({
  currentRoundQuestions,
  selectQuestion,
  canSelectQuestion,
}: {
  currentRoundQuestions: { [key: string]: BoardQuestion[] };
  selectQuestion: (question: { category: string; index: number }) => void;
  canSelectQuestion: boolean;
}) {
  const getTableData = (currentRoundQuestions: { [key: string]: BoardQuestion[] }) => {
    const categoriesCount = Object.keys(currentRoundQuestions).length;
    const questionsInCategoryCount = Object.values(currentRoundQuestions)[0].length;
  
    const tableData: {
      value: number;
      hasBeenPlayed: boolean;
      onClick: () => void;
    }[][] = new Array(questionsInCategoryCount)
      .fill(undefined)
      .map(() => new Array(categoriesCount).fill(undefined));
  
    for (const [categoryIndex, [category, questions]] of Object.entries(
      currentRoundQuestions
    ).entries()) {
      for (const question of questions) {
        tableData[question.index][categoryIndex] = {
          value: question.value,
          hasBeenPlayed: question.hasBeenPlayed,
          onClick: () => selectQuestion({ category, index: question.index }),
        };
      }
    }
    return tableData;
  };

  const tableData = getTableData(currentRoundQuestions);

  return (
    <table className="w-full h-full table-fixed">
      <thead>
        <tr>
          {Object.keys(currentRoundQuestions).map((category, index) => (
            <th className="border break-all" key={index} scope="col">
              {category}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {tableData.map((row, i) => (
          <tr key={i}>
            {row.map(({ value, hasBeenPlayed, onClick }, j) =>
              hasBeenPlayed ? (
                <td className="border" key={j}>&nbsp;</td>
              ) : (
                <td
                  className={`text-center text-lg font-bold border${
                    canSelectQuestion ? " hover:bg-white hover:text-black" : ""
                  }`}
                  key={j}
                  onClick={canSelectQuestion ? onClick : undefined}
                >
                  {value}
                </td>
              )
            )}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
