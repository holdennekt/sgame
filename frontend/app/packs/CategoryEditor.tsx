"use client";

import { CategoryFormData, PackFormData, QuestionFormData } from "@/types/pack";
import React from "react";
import { IoIosAdd } from "react-icons/io";
import {
  DndContext,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
  arrayMove,
} from "@dnd-kit/sortable";
import SortableQuestion from "./SortableQuestion";

export default function CategoryEditor({
  roundIndex,
  categoryIndex,
  category,
  pack,
  setPack,
  setQuestionModal,
  readOnly = false,
}: {
  roundIndex: number;
  categoryIndex: number;
  category: CategoryFormData;
  pack: PackFormData;
  setPack: React.Dispatch<React.SetStateAction<PackFormData>>;
  setQuestionModal: React.Dispatch<
    React.SetStateAction<{
      isOpen: boolean;
      question: QuestionFormData;
      saveQuestion: (q: Omit<QuestionFormData, "index">) => void;
    }>
  >;
  readOnly?: boolean;
}) {
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5, delay: 100, tolerance: 5 },
    })
  );

  const renameCategory = (name: string) => {
    setPack((pack) => {
      pack.rounds[roundIndex].categories[categoryIndex].name = name;
      return { ...pack };
    });
  };

  const addQuestion = () => {
    const qi =
      pack.rounds[roundIndex].categories[categoryIndex].questions.length;
    setQuestionModal({
      isOpen: true,
      question: {
        index: qi,
        value: 0,
        text: "",
        attachment: { type: "file" },
        type: "regular",
        answers: [],
        comment: { text: "", attachment: { type: "file" } },
      },
      saveQuestion: (q) =>
        setPack((pack) => {
          pack.rounds[roundIndex].categories[categoryIndex].questions.push({
            ...q,
            index: qi,
          });
          return { ...pack };
        }),
    });
  };

  const duplicateQuestion = (qi: number) => {
    setPack((pack) => {
      const questions =
        pack.rounds[roundIndex].categories[categoryIndex].questions;
      const src = questions[qi];
      const copy = {
        ...src,
        answers: [...src.answers],
        attachment: { ...src.attachment },
      };
      questions.splice(qi + 1, 0, copy);
      pack.rounds[roundIndex].categories[categoryIndex].questions =
        questions.map((q, i) => ({ ...q, index: i }));
      return { ...pack };
    });
  };

  const deleteQuestion = (qi: number) => {
    setPack((pack) => {
      pack.rounds[roundIndex].categories[categoryIndex].questions = pack.rounds[
        roundIndex
      ].categories[categoryIndex].questions.filter((_, i) => i !== qi);
      return { ...pack };
    });
  };

  const openQuestionModal = (question: QuestionFormData, qi: number) => {
    setQuestionModal({
      isOpen: true,
      question,
      saveQuestion: (q) =>
        setPack((pack) => {
          pack.rounds[roundIndex].categories[categoryIndex].questions[qi] = {
            ...q,
            index: qi,
          };
          return { ...pack };
        }),
    });
  };

  const onDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    setPack((pack) => {
      const questions =
        pack.rounds[roundIndex].categories[categoryIndex].questions;
      const oldIndex = questions.findIndex((_, i) => String(i) === active.id);
      const newIndex = questions.findIndex((_, i) => String(i) === over.id);
      pack.rounds[roundIndex].categories[categoryIndex].questions = arrayMove(
        questions,
        oldIndex,
        newIndex
      ).map((q, i) => ({ ...q, index: i }));
      return { ...pack };
    });
  };

  const questions = category.questions;

  return (
    <div className="h-full flex flex-col gap-3">
      {/* Category name */}
      <div className="shrink-0">
        <input
          className="w-full h-9 px-2.5 bg-background border border-border text-on-background rounded-lg text-sm font-semibold outline-none focus-ring placeholder:text-on-surface-muted transition-[border-color] duration-150"
          placeholder="Category name"
          value={category.name}
          onChange={(e) => renameCategory(e.target.value)}
          readOnly={readOnly}
        />
      </div>

      {/* Questions */}
      {questions.length === 0 && readOnly ? (
        <div className="flex-1 flex items-center justify-center text-on-surface-muted">
          <p className="text-sm">No questions</p>
        </div>
      ) : (
        <div className="flex flex-col gap-1">
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={onDragEnd}
          >
            <SortableContext
              items={questions.map((_, i) => String(i))}
              strategy={verticalListSortingStrategy}
            >
              {questions.map((question, qi) => (
                <SortableQuestion
                  key={qi}
                  id={String(qi)}
                  question={question}
                  readOnly={readOnly}
                  onOpen={() => openQuestionModal(question, qi)}
                  onDuplicate={() => duplicateQuestion(qi)}
                  onDelete={() => deleteQuestion(qi)}
                />
              ))}
            </SortableContext>
          </DndContext>

          {!readOnly && (
            <button
              type="button"
              className="w-full flex items-center justify-center gap-1.5 px-4 py-3 rounded-md border border-dashed border-border text-on-surface-muted hover:border-primary hover:text-primary transition-colors duration-150 text-sm font-medium"
              onClick={addQuestion}
            >
              <IoIosAdd size={14} />
              Add question
            </button>
          )}
        </div>
      )}
    </div>
  );
}
