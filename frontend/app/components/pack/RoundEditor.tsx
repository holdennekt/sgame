import React from "react";
import Accordion from "../Accordion";
import { Category, PackDTO, Question, Round } from "./PackEditor";
import { FaTrashCan } from "react-icons/fa6";
import { IoIosArrowUp, IoIosArrowDown } from "react-icons/io";
import CategoryEditor from "./CategoryEditor";

export default function RoundEditor({
  round,
  index: roundIndex,
  pack,
  setPack,
  setQuestionModal,
  readOnly = false,
}: {
  round: Round;
  index: number;
  pack: PackDTO;
  setPack: React.Dispatch<React.SetStateAction<PackDTO>>;
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
  const changeRound = (roundIndex: number, round: Round) => {
    setPack((pack) => {
      pack.rounds[roundIndex] = round;
      return { ...pack };
    });
  };

  const deleteRound = (roundIndex: number) => {
    setPack((pack) => {
      pack.rounds = pack.rounds.filter((r, i) => roundIndex !== i);
      return { ...pack };
    });
  };

  const addCategory = (roundIndex: number) => {
    setPack((pack) => {
      pack.rounds[roundIndex].categories.push({
        name: `Category ${pack.rounds[roundIndex].categories.length + 1}`,
        questions: [],
      });
      return { ...pack };
    });
  };

  return (
    <>
      <div className="flex items-end gap-4 mt-2">
        <label>
          <p className="font-medium">Round name</p>
          <input
            className="w-48 h-8 rounded-md mt-2 p-2 text-black"
            type="text"
            placeholder="Name"
            value={round.name}
            onChange={(e) =>
              changeRound(roundIndex, {
                ...round,
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
              onClick={() => addCategory(roundIndex)}
            >
              Add category
            </button>
            <div className="flex-1 flex flex-row-reverse">
              <div className="flex items-center gap-4">
                <button
                  className="h-8 aspect-square px-2 py-1 border rounded"
                  type="button"
                  onClick={() => {
                    if (roundIndex === 0) return;
                    const roundBefore = pack.rounds[roundIndex - 1];
                    pack.rounds[roundIndex - 1] = round;
                    pack.rounds[roundIndex] = roundBefore;
                    setPack({ ...pack });
                  }}
                >
                  <IoIosArrowUp size="auto" />
                </button>
                <button
                  className="h-8 aspect-square px-2 py-1 border rounded"
                  type="button"
                  onClick={() => {
                    if (roundIndex === pack.rounds.length - 1) return;
                    const roundAfter = pack.rounds[roundIndex + 1];
                    pack.rounds[roundIndex + 1] = round;
                    pack.rounds[roundIndex] = roundAfter;
                    setPack({ ...pack });
                  }}
                >
                  <IoIosArrowDown size="auto" />
                </button>
                <button
                  className="h-8 aspect-square px-2 py-1 border rounded text-red-600"
                  type="button"
                  onClick={() => deleteRound(roundIndex)}
                >
                  <FaTrashCan size="auto" />
                </button>
              </div>
            </div>
          </>
        )}
      </div>
      <div className="mt-4 pl-6 flex flex-col gap-2">
        {round.categories.map((category, categoryIndex) => (
          <Accordion
            title={category.name}
            key={categoryIndex}
          >
            <CategoryEditor
              roundIndex={roundIndex}
              category={category}
              index={categoryIndex}
              pack={pack}
              setPack={setPack}
              setQuestionModal={setQuestionModal}
              readOnly={readOnly}
            />
          </Accordion>
        ))}
      </div>
    </>
  );
}
