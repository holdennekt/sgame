"use client";

import { CategoryFormData, PackFormData, QuestionFormData } from "@/types/pack";
import React, { useState } from "react";
import { IoIosAdd } from "react-icons/io";
import { MdOutlineContentCopy } from "react-icons/md";
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
  setPack,
  setQuestionModal,
  onCopyJson,
  readOnly = false,
}: {
  roundIndex: number;
  categoryIndex: number;
  category: CategoryFormData;
  setPack: React.Dispatch<React.SetStateAction<PackFormData>>;
  setQuestionModal: React.Dispatch<
    React.SetStateAction<{
      isOpen: boolean;
      question: QuestionFormData;
      saveQuestion: (q: QuestionFormData) => void;
    }>
  >;
  onCopyJson?: () => void;
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

  const updateComment = (comment: string) => {
    setPack((pack) => {
      pack.rounds[roundIndex].categories[categoryIndex].comment = comment;
      return { ...pack };
    });
  };

  const addQuestion = () => {
    setQuestionModal({
      isOpen: true,
      question: {
        value: 0,
        text: "",
        attachment: { type: "file" },
        type: "regular",
        answers: [],
        comment: { text: "", attachment: { type: "file" } },
      },
      saveQuestion: (q) =>
        setPack((pack) => {
          pack.rounds[roundIndex].categories[categoryIndex].questions.push(q);
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
          pack.rounds[roundIndex].categories[categoryIndex].questions[qi] = q;
          return { ...pack };
        }),
    });
  };

  const [autoAssignModal, setAutoAssignModal] = useState(false);
  const [autoAssignFactor, setAutoAssignStep] = useState(100);

  const autoAssignValues = () => {
    setPack((pack) => {
      pack.rounds[roundIndex].categories[categoryIndex].questions.forEach(
        (q, i) => {
          q.value = (i + 1) * autoAssignFactor;
        }
      );
      return { ...pack };
    });
    setAutoAssignModal(false);
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
      );
      return { ...pack };
    });
  };

  const questions = category.questions;

  return (
    <div className="h-full flex flex-col gap-3">
      {/* Category name + comment */}
      <div className="shrink-0 flex flex-col gap-1.5">
        <div className="flex items-center gap-1.5">
          <input
            className="flex-1 min-w-0 h-9 px-2.5 bg-background border border-border text-on-background rounded-lg text-sm font-semibold outline-none focus-ring placeholder:text-on-surface-muted transition-[border-color] duration-150"
            placeholder="Category name"
            value={category.name}
            onChange={(e) => renameCategory(e.target.value)}
            readOnly={readOnly}
          />
          {!readOnly && (
            <div className="shrink-0 flex items-center gap-1.5">
              {questions.length > 0 && (
                <button
                  type="button"
                  className="h-9 px-2.5 inline-flex items-center gap-1.5 rounded-lg border border-border text-xs text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150"
                  onClick={() => setAutoAssignModal(true)}
                >
                  Auto-assign values
                </button>
              )}
              {onCopyJson && (
                <button
                  type="button"
                  className="h-9 px-2.5 inline-flex items-center gap-1.5 rounded-lg border border-border text-xs text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150"
                  onClick={onCopyJson}
                >
                  <MdOutlineContentCopy size={13} />
                  Copy category
                </button>
              )}
            </div>
          )}
        </div>

        {autoAssignModal && (
          <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
            onClick={() => setAutoAssignModal(false)}
          >
            <div
              className="bg-background border border-border rounded-xl p-5 flex flex-col gap-4 w-72 shadow-xl"
              onClick={(e) => e.stopPropagation()}
            >
              <p className="text-sm font-semibold text-on-background">
                Auto-assign values
              </p>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-on-surface-muted">Factor</label>
                <input
                  type="text"
                  inputMode="numeric"
                  pattern="[0-9]*"
                  className="h-9 px-2.5 bg-background border border-border text-on-background rounded-lg text-sm outline-none focus-ring"
                  value={autoAssignFactor}
                  onChange={(e) =>
                    setAutoAssignStep(
                      Number(e.target.value.replace(/[^0-9]/g, ""))
                    )
                  }
                  autoFocus
                  onKeyDown={(e) => e.key === "Enter" && autoAssignValues()}
                />
                <p className="text-xs text-on-surface-muted">
                  Questions will be assigned {autoAssignFactor},{" "}
                  {2 * autoAssignFactor}, {3 * autoAssignFactor}…
                </p>
              </div>
              <div className="flex gap-2 justify-end">
                <button
                  type="button"
                  className="h-8 px-3 rounded-lg border border-border text-xs text-on-surface-muted hover:bg-surface-raised transition-colors duration-150"
                  onClick={() => setAutoAssignModal(false)}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="h-8 px-3 rounded-lg bg-primary text-on-primary text-xs font-medium hover:opacity-90 transition-opacity duration-150"
                  onClick={autoAssignValues}
                >
                  Apply
                </button>
              </div>
            </div>
          </div>
        )}
        <textarea
          className="w-full px-2.5 py-1.5 bg-background border border-border text-on-background rounded-lg text-sm outline-none focus-ring placeholder:text-on-surface-muted transition-[border-color] duration-150 resize-none"
          placeholder="Category comment (optional)"
          rows={2}
          value={category.comment}
          onChange={(e) => updateComment(e.target.value)}
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
