import React from "react";
import { FaTrashCan } from "react-icons/fa6";
import {
  IoIosArrowUp,
  IoIosArrowDown,
  IoIosArrowBack,
  IoIosArrowForward,
} from "react-icons/io";
import { Category, Pack, Question } from "./PackEditor";

export default function CategoryEditor({
  roundIndex,
  category,
  index: categoryIndex,
  pack,
  setPack,
  setQuestionModal,
  readOnly = false,
}: {
  roundIndex: number;
  category: Category;
  index: number;
  pack: Pack;
  setPack: React.Dispatch<React.SetStateAction<Pack>>;
  setQuestionModal: React.Dispatch<
    React.SetStateAction<{
      isOpen: boolean;
      roundIndex: number;
      categoryIndex: number;
      questionIndex: number;
      question: Question;
    }>
  >;
  readOnly?: boolean;
}) {
  const changeCategory = (
    roundIndex: number,
    categoryIndex: number,
    category: Category
  ) => {
    setPack(pack => {
      pack.rounds[roundIndex].categories[categoryIndex] = category;
      return { ...pack };
    });
  };

  const deleteCategory = (roundIndex: number, categoryIndex: number) => {
    setPack(pack => {
      pack.rounds[roundIndex].categories = pack.rounds[
        roundIndex
      ].categories.filter((c, i) => categoryIndex !== i);
      return { ...pack };
    });
  };

  const addQuestion = (roundIndex: number, categoryIndex: number) => {
    setPack(pack => {
      pack.rounds[roundIndex].categories[categoryIndex].questions.push({
        index: 0,
        value: 0,
        text: "",
        attachment: null,
        type: "regular",
        answers: [],
        comment: null,
      });
      return { ...pack };
    });
  };

  const deleteQuestion = (
    roundIndex: number,
    categoryIndex: number,
    questionIndex: number
  ) => {
    setPack(pack => {
      pack.rounds[roundIndex].categories[categoryIndex].questions = pack.rounds[
        roundIndex
      ].categories[categoryIndex].questions.filter(
        (c, i) => questionIndex !== i
      );
      return { ...pack };
    });
  };

  return (
    <>
      <div className="flex items-end gap-4 mt-2">
        <label>
          <p className="font-medium">Category name</p>
          <input
            className="w-48 h-8 rounded-md mt-1 p-1 text-black"
            type="text"
            placeholder="Name"
            value={category.name}
            onChange={e =>
              changeCategory(roundIndex, categoryIndex, {
                ...category,
                name: e.target.value,
              })
            }
            readOnly={readOnly}
          />
        </label>
        {!readOnly && (
          <>
            <button
              className="w-fit h-fit rounded px-2 py-1 primary"
              type="button"
              onClick={() => addQuestion(roundIndex, categoryIndex)}
            >
              Add question
            </button>
            <div className="flex-1 flex flex-row-reverse">
              <div className="flex items-center gap-4">
                <button
                  className="h-8 aspect-square px-2 py-1 border rounded"
                  type="button"
                  onClick={() => {
                    if (categoryIndex === 0) return;
                    const categoryBefore =
                      pack.rounds[roundIndex].categories[categoryIndex - 1];
                    pack.rounds[roundIndex].categories[categoryIndex - 1] =
                      category;
                    pack.rounds[roundIndex].categories[categoryIndex] =
                      categoryBefore;
                    setPack({ ...pack });
                  }}
                >
                  <IoIosArrowUp size="auto" />
                </button>
                <button
                  className="h-8 aspect-square px-2 py-1 border rounded"
                  type="button"
                  onClick={() => {
                    if (
                      categoryIndex ===
                      pack.rounds[roundIndex].categories.length - 1
                    )
                      return;
                    const categoryAfter =
                      pack.rounds[roundIndex].categories[categoryIndex + 1];
                    pack.rounds[roundIndex].categories[categoryIndex + 1] =
                      category;
                    pack.rounds[roundIndex].categories[categoryIndex] =
                      categoryAfter;
                    setPack({ ...pack });
                  }}
                >
                  <IoIosArrowDown size="auto" />
                </button>
                <button
                  className="h-8 aspect-square px-2 py-1 border rounded text-red-600"
                  type="button"
                  onClick={() => deleteCategory(roundIndex, categoryIndex)}
                >
                  <FaTrashCan size="auto" />
                </button>
              </div>
            </div>
          </>
        )}
      </div>
      {category.questions.length > 0 && (
        <div className="flex justify-center mt-4">
          <table className="border">
            <thead>
              <tr>
                <th className="p-2 border"></th>
                {category.questions.map((question, questionIndex) => (
                  <th
                    className="p-2 border text-xl font-bold"
                    key={questionIndex}
                  >
                    {question.value}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              <tr>
                <th className="p-2 border">{category.name}</th>
                {category.questions.map((question, questionIndex) => (
                  <td className="p-2 border" align="center" key={questionIndex}>
                    <button
                      className="p-2 aspect-square rounded primary"
                      type="button"
                      onClick={() =>
                        setQuestionModal({
                          isOpen: true,
                          question,
                          roundIndex,
                          categoryIndex,
                          questionIndex,
                        })
                      }
                    >
                      Q{questionIndex + 1}
                    </button>
                  </td>
                ))}
              </tr>
              {!readOnly && (
                <tr>
                  <td className="p-2 border"></td>
                  {category.questions.map((question, questionIndex) => (
                    <td className="p-2 border" key={questionIndex}>
                      <button
                        className="h-8 aspect-square px-2 py-1 border rounded-l"
                        type="button"
                        onClick={() => {
                          if (questionIndex === 0) return;
                          const questionBefore =
                            pack.rounds[roundIndex].categories[categoryIndex]
                              .questions[questionIndex - 1];
                          pack.rounds[roundIndex].categories[
                            categoryIndex
                          ].questions[questionIndex - 1] = question;
                          pack.rounds[roundIndex].categories[
                            categoryIndex
                          ].questions[questionIndex] = questionBefore;
                          setPack({ ...pack });
                        }}
                      >
                        <IoIosArrowBack size="auto" />
                      </button>
                      <button
                        className="h-8 aspect-square px-2 py-1 border-t border-b text-red-600"
                        type="button"
                        onClick={() =>
                          deleteQuestion(
                            roundIndex,
                            categoryIndex,
                            questionIndex
                          )
                        }
                      >
                        <FaTrashCan size="auto" />
                      </button>
                      <button
                        className="h-8 aspect-square px-2 py-1 border rounded-r"
                        type="button"
                        onClick={() => {
                          if (
                            questionIndex ===
                            pack.rounds[roundIndex].categories[categoryIndex]
                              .questions.length -
                              1
                          )
                            return;
                          const questionAfter =
                            pack.rounds[roundIndex].categories[categoryIndex]
                              .questions[questionIndex + 1];
                          pack.rounds[roundIndex].categories[
                            categoryIndex
                          ].questions[questionIndex + 1] = question;
                          pack.rounds[roundIndex].categories[
                            categoryIndex
                          ].questions[questionIndex] = questionAfter;
                          setPack({ ...pack });
                        }}
                      >
                        <IoIosArrowForward size="auto" />
                      </button>
                    </td>
                  ))}
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}
