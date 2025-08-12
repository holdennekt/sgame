"use client";

import React, { FormEventHandler, useState } from "react";
import { HiddenQuestion } from "../room/Room";
import { HiddenCategory } from "../PacksList";
import QuestionModal from "./QuestionModal";
import { toast, ToastContainer } from "react-toastify";
import { usePathname, useRouter } from "next/navigation";
import RoundEditor from "./RoundEditor";
import Accordion from "../Accordion";
import Link from "next/link";
import FinalRoundCategoryModal from "./FinalRoundCategoryModal";
import { FaTrashCan } from "react-icons/fa6";

export type PackDTO = {
  name: string;
  type: "public" | "private";
  rounds: Round[];
  finalRound: FinalRound;
};
export type Round = {
  name: string;
  categories: Category[];
};
export type Category = HiddenCategory & {
  questions: Question[];
};
export type Answer = {
  answers: string[];
  comment: string | null;
};
export type QuestionType = "regular" | "catInBag" | "auction";
export type Question = HiddenQuestion & { type: QuestionType } & Answer;

const dummyQuestion: Question = {
  index: 0,
  value: 0,
  text: "",
  attachment: null,
  type: "regular",
  answers: [],
  comment: null,
};

export const isQuestion = (obj: unknown): obj is Question => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyQuestion).every((key) => Object.hasOwn(obj, key));
};

export type FinalRound = {
  categories: FinalRoundCategory[];
};
export type FinalRoundCategory = {
  name: string;
  question: FinalRoundQuestion;
};
export type FinalRoundQuestion = Answer & {
  text: string;
  attachment: {
    mediaType: "image" | "audio" | "video";
    contentUrl: string;
  } | null;
};

export default function PackEditor({
  handlePack,
  initialPack,
  readOnly = false,
}: {
  handlePack: (pack: PackDTO) => Promise<{ id: string }>;
  initialPack?: PackDTO;
  readOnly?: boolean;
}) {
  const router = useRouter();
  const pathname = usePathname();

  const [pack, setPack] = useState<PackDTO>(
    initialPack ?? {
      name: "",
      type: "public",
      rounds: [{ name: "Round 1", categories: [] }],
      finalRound: { categories: [] },
    }
  );
  const [finalRoundCategoryNameInput, setFinalRoundCategoryNameInput] =
    useState("");
  const [questionModal, setQuestionModal] = useState<{
    isOpen: boolean;
    roundIndex: number;
    categoryIndex: number;
    questionIndex: number;
    question: Question;
  }>({
    isOpen: false,
    roundIndex: -1,
    categoryIndex: -1,
    questionIndex: -1,
    question: {
      index: 0,
      value: 0,
      text: "",
      attachment: null,
      type: "regular",
      answers: [],
      comment: null,
    },
  });
  const [finalRoundCategoryModal, setFinalCategoryModal] = useState<{
    isOpen: boolean;
    index: number;
    category: FinalRoundCategory;
  }>({
    isOpen: false,
    index: -1,
    category: {
      name: "",
      question: {
        text: "",
        attachment: null,
        answers: [],
        comment: null,
      },
    },
  });

  const addRound = () => {
    pack.rounds.push({
      name: `Round ${pack.rounds.length + 1}`,
      categories: [],
    });
    setPack({ ...pack });
  };

  const changeQuestion = (
    roundIndex: number,
    categoryIndex: number,
    questionIndex: number,
    question: Question
  ) => {
    pack.rounds[roundIndex].categories[categoryIndex].questions[questionIndex] =
      { ...question, index: questionModal.questionIndex };
    setPack({ ...pack });
  };

  const addFinalRoundCategory = (name: string) => {
    pack.finalRound.categories.push({
      name,
      question: { text: "", attachment: null, answers: [], comment: null },
    });
    setPack({ ...pack });
  };

  const changeFinalRoundCategory = (
    index: number,
    category: FinalRoundCategory
  ) => {
    pack.finalRound.categories[index] = category;
    setPack({ ...pack });
  };

  const deleteFinalRoundCategory = (categoryIndex: number) => {
    setPack((pack) => {
      pack.finalRound.categories = pack.finalRound.categories.filter(
        (c, i) => categoryIndex !== i
      );
      return { ...pack };
    });
  };

  const onSubmit: FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();

    try {
      const obj = await handlePack(pack);
      const url = `/packs/${obj.id}`;
      router.push(url);
      toast.success("Pack successfully saved!", { containerId: "editor" });
    } catch (error) {
      if (error instanceof Error)
        return toast.error(error.message, { containerId: "editor" });
    }
  };

  return (
    <>
      <form className="min-h-0 h-full flex flex-col gap-4" onSubmit={onSubmit}>
        <div className="flex items-end gap-4">
          <label>
            <p className="font-medium">Pack name</p>
            <input
              className="w-48 h-8 rounded-md mt-1 p-1 text-black"
              type="text"
              placeholder="Name"
              value={pack.name}
              onChange={(e) => setPack({ ...pack, name: e.target.value })}
              readOnly={readOnly}
            />
          </label>
          <label>
            <p className="font-medium">Privacy Type</p>
            {readOnly ? (
              <input
                className="w-48 h-8 rounded-md mt-1 p-1 text-black"
                type="text"
                value={pack.type}
                readOnly={readOnly}
              />
            ) : (
              <select
                className="w-48 h-8 mt-1 p-0.5 rounded-md text-black"
                value={pack.type}
                onChange={(e) =>
                  setPack({
                    ...pack,
                    type: e.target.value as "public" | "private",
                  })
                }
              >
                <option value="public">Public</option>
                <option value="private">Private</option>
              </select>
            )}
          </label>
          {!readOnly && (
            <button
              className="w-fit h-fit rounded px-2 py-1 primary"
              type="button"
              onClick={addRound}
            >
              Add round
            </button>
          )}
        </div>
        <div className="flex-1 flex flex-col gap-2 overflow-y-auto">
          {pack.rounds.map((round, roundIndex) => (
            <Accordion title={round.name} key={roundIndex}>
              <RoundEditor
                round={round}
                index={roundIndex}
                pack={pack}
                setPack={setPack}
                setQuestionModal={setQuestionModal}
                readOnly={readOnly}
              />
            </Accordion>
          ))}
          <Accordion title="Final round">
            {!readOnly && (
              <>
                <label>
                  <p className="mt-2 font-medium">Category name</p>
                  <input
                    className="w-48 h-8 rounded-md mt-1 p-1 text-black"
                    type="text"
                    placeholder="Name"
                    value={finalRoundCategoryNameInput}
                    onChange={(e) =>
                      setFinalRoundCategoryNameInput(e.target.value)
                    }
                    readOnly={readOnly}
                  />
                </label>
                <button
                  className="w-fit h-fit ml-4 rounded px-2 py-1 primary"
                  type="button"
                  onClick={() => {
                    addFinalRoundCategory(finalRoundCategoryNameInput);
                    setFinalRoundCategoryNameInput("");
                  }}
                >
                  Add category
                </button>
              </>
            )}
            <ul className="flex flex-col gap-2 mt-4 pl-4 list-inside list-disc">
              {pack.finalRound.categories.map((category, index) => (
                <li key={index}>
                  <button
                    className="w-fit h-fit px-2 py-1 border rounded"
                    type="button"
                    onClick={() =>
                      setFinalCategoryModal({ isOpen: true, index, category })
                    }
                  >
                    {category.name}
                  </button>
                  <button
                    className="h-8 aspect-square px-2 py-1 border rounded text-red-600"
                    type="button"
                    onClick={() => deleteFinalRoundCategory(index)}
                  >
                    <FaTrashCan size="auto" />
                  </button>
                </li>
              ))}
            </ul>
          </Accordion>
        </div>
        {readOnly ? (
          <div className="flex flex-row-reverse">
            <Link
              className="w-fit h-fit rounded px-2 py-1 primary"
              href={`${usePathname()}?edit=true`}
            >
              Edit
            </Link>
          </div>
        ) : (
          <div className="flex flex-row-reverse">
            <div>
              <button
                className="w-fit h-fit px-2 py-1 border rounded"
                type="button"
                onClick={() => {
                  router.push(pathname.split("?")[0]);
                }}
              >
                Discard
              </button>
              <button className="w-fit h-fit ml-4 px-2 py-1 rounded primary">
                Save
              </button>
            </div>
          </div>
        )}
      </form>
      <QuestionModal
        isOpen={questionModal.isOpen}
        close={() => setQuestionModal({ ...questionModal, isOpen: false })}
        question={questionModal.question}
        saveQuestion={changeQuestion.bind(
          null,
          questionModal.roundIndex,
          questionModal.categoryIndex,
          questionModal.questionIndex
        )}
        readOnly={readOnly}
      />
      <FinalRoundCategoryModal
        isOpen={finalRoundCategoryModal.isOpen}
        close={() =>
          setFinalCategoryModal({ ...finalRoundCategoryModal, isOpen: false })
        }
        category={finalRoundCategoryModal.category}
        saveCategory={changeFinalRoundCategory.bind(
          null,
          finalRoundCategoryModal.index
        )}
        readOnly={readOnly}
      />
      <ToastContainer
        containerId="editor"
        position="bottom-left"
        theme="colored"
      />
    </>
  );
}
