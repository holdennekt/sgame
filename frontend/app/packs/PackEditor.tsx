"use client";

import { FormEventHandler, useState } from "react";
import QuestionModal from "./QuestionModal";
import FinalRoundCategoryModal from "./FinalRoundCategoryModal";
import CategoryEditor from "./CategoryEditor";
import SortableRound from "./SortableRound";
import SortableCategory from "./SortableCategory";
import { toast, ToastContainer } from "react-toastify";
import { usePathname, useRouter } from "next/navigation";
import Link from "next/link";
import { FaTrashCan } from "react-icons/fa6";
import {
  IoIosAdd,
  IoIosArrowDown as ChevronDown,
  IoIosArrowDown as SelectArrow,
} from "react-icons/io";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import {
  convertPackFormDataToRequest,
  convertPackToFormData,
  CreatePackRequest,
  FinalRoundCategoryFormData,
  Pack,
  QuestionFormData,
} from "@/types/pack";
import { signURL } from "@/app/actions";
import { isError } from "@/middleware";
import { usePack } from "@/hooks/usePack";

const inputCls =
  "h-9 px-2.5 bg-background border border-border text-on-background rounded-lg text-sm outline-none focus-ring placeholder:text-on-surface-muted transition-[border-color] duration-150";
const selectCls =
  "h-9 pl-2.5 pr-8 bg-background border border-border text-on-background rounded-lg text-sm outline-none focus-ring appearance-none";
const iconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150";
const dangerIconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-danger hover:bg-surface-raised transition-colors duration-150";

export default function PackEditor({
  savePack,
  initialPack,
  readOnly = false,
}: {
  savePack: (pack: CreatePackRequest) => Promise<{ id: string } | { error: string }>;
  initialPack: Omit<Pack, "id" | "createdBy">;
  readOnly?: boolean;
}) {
  const router = useRouter();
  const pathname = usePathname();

  const {
    pack,
    setPack,
    sidebarRef,
    sensors,
    selectedCategory,
    selectedCategoryIndex,
    selectedRoundIndex,
    addRound,
    duplicateRound,
    deleteRound,
    renameRound,
    toggleRoundExpand,
    selectCategory,
    addCategory,
    duplicateCategory,
    deleteCategory,
    addFinalRoundCategory,
    changeFinalRoundCategory,
    deleteFinalRoundCategory,
    onDragEndRounds,
    onDragEndCategories,
  } = usePack(initialPack);

  const [questionModal, setQuestionModal] = useState<{
    isOpen: boolean;
    question: QuestionFormData;
    saveQuestion: (q: QuestionFormData) => void;
  }>({
    isOpen: false,
    question: {
      index: 0,
      value: 0,
      text: "",
      attachment: { type: "file" },
      type: "regular",
      answers: [],
      comment: null,
    },
    saveQuestion: () => {},
  });

  const [finalRoundCategoryModal, setFinalRoundCategoryModal] = useState<{
    isOpen: boolean;
    category: FinalRoundCategoryFormData;
    saveCategory: (c: FinalRoundCategoryFormData) => void;
  }>({
    isOpen: false,
    category: {
      name: "",
      question: {
        text: "",
        attachment: { type: "file" },
        answers: [],
        comment: null,
      },
    },
    saveCategory: () => {},
  });

  const onSubmit: FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();
    try {
      const result = await savePack(
        await convertPackFormDataToRequest(pack, signURL),
      );
      if (isError(result)) {
        toast.error(result.error, { containerId: "editor" });
        return;
      }
      router.push(`/packs/${result.id}`);
      toast.success("Pack successfully saved!", { containerId: "editor" });
    } catch (error) {
      if (error instanceof Error)
        toast.error(error.message, { containerId: "editor" });
    }
  };

  return (
    <>
      <form className="min-h-0 h-full flex flex-col gap-3" onSubmit={onSubmit}>
        <div className="flex items-center justify-between gap-2 pb-3 border-b border-border shrink-0">
          <div className="flex-1 flex items-center gap-2 min-w-0">
            <input
              className={`${inputCls} flex-1 min-w-0`}
              type="text"
              placeholder="Pack name"
              value={pack.name}
              onChange={(e) => setPack({ ...pack, name: e.target.value })}
              readOnly={readOnly}
            />
            {readOnly ? (
              <span className="h-9 px-2.5 flex items-center bg-surface-raised border border-border text-on-surface-muted rounded-lg text-sm capitalize">
                {pack.type}
              </span>
            ) : (
              <div className="relative">
                <select
                  className={selectCls}
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
                <div className="pointer-events-none absolute inset-y-0 right-2.5 flex items-center text-on-surface-muted">
                  <SelectArrow size={14} />
                </div>
              </div>
            )}
          </div>
          {readOnly ? (
            <Link
              className="shrink-0 inline-flex items-center justify-center px-3.5 py-1.5 rounded-md text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
              href={`${pathname}?edit=true`}
            >
              Edit
            </Link>
          ) : (
            <div className="shrink-0 flex items-center gap-2">
              <button
                className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-md text-sm font-medium border border-border text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150"
                type="button"
                onClick={() => {
                  if (!confirm("Discard all changes?")) return;
                  setPack(convertPackToFormData(initialPack));
                  router.push(pathname.split("?")[0]);
                }}
              >
                Discard
              </button>
              <button className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-md text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150">
                Save
              </button>
            </div>
          )}
        </div>

        {/* ── Body ─────────────────────────────────────────────────── */}
        <div className="flex-1 flex flex-col md:flex-row gap-3 min-h-0">
          {/* ── Sidebar ────────────────────────────────────────────── */}
          <div
            ref={sidebarRef}
            className="w-full md:w-52 shrink-0 flex flex-col min-h-0 max-h-48 md:max-h-none overflow-y-auto"
          >
            {/* Rounds header */}
            <div className="flex items-center justify-between px-1 mb-1">
              <span className="text-[11px] font-semibold uppercase tracking-widest text-on-surface-muted">
                Rounds
              </span>
              {!readOnly && (
                <button
                  type="button"
                  title="Add round"
                  className={iconBtnCls}
                  onClick={addRound}
                >
                  <IoIosAdd size={15} />
                </button>
              )}
            </div>

            {/* Round list */}
            <div className="flex flex-col gap-0.5">
              <DndContext
                sensors={sensors}
                collisionDetection={closestCenter}
                onDragEnd={onDragEndRounds}
              >
                <SortableContext
                  items={pack.rounds.map((_, i) => String(i))}
                  strategy={verticalListSortingStrategy}
                >
                  {pack.rounds.map((round, ri) => (
                    <SortableRound
                      key={ri}
                      id={String(ri)}
                      round={round}
                      readOnly={readOnly}
                      onToggleExpand={() => toggleRoundExpand(ri)}
                      onRename={(name) => renameRound(ri, name)}
                      onDuplicate={() => duplicateRound(ri)}
                      onDelete={() => deleteRound(ri)}
                      onAddRound={addRound}
                    >
                      {round.expanded && (
                        <div className="flex flex-col gap-0.5 mb-1">
                          {round.categories.length === 0 && (
                            <p className="pl-5 text-xs text-on-surface-muted py-1 italic">
                              No categories
                            </p>
                          )}
                          <DndContext
                            sensors={sensors}
                            collisionDetection={closestCenter}
                            onDragEnd={(e) => onDragEndCategories(ri, e)}
                          >
                            <SortableContext
                              items={round.categories.map((_, i) => String(i))}
                              strategy={verticalListSortingStrategy}
                            >
                              {round.categories.map((cat, ci) => (
                                <SortableCategory
                                  key={ci}
                                  id={String(ci)}
                                  cat={cat}
                                  readOnly={readOnly}
                                  onSelect={() => selectCategory(ri, ci)}
                                  onDuplicate={() => duplicateCategory(ri, ci)}
                                  onDelete={() => deleteCategory(ri, ci)}
                                />
                              ))}
                            </SortableContext>
                          </DndContext>
                          {!readOnly && (
                            <button
                              type="button"
                              className="pl-5 pr-1 py-1 text-xs text-on-surface-muted hover:text-primary flex items-center gap-1 transition-colors duration-150"
                              onClick={() => addCategory(ri)}
                            >
                              <IoIosAdd size={13} />
                              Add category
                            </button>
                          )}
                        </div>
                      )}
                    </SortableRound>
                  ))}
                </SortableContext>
              </DndContext>
            </div>

            {/* Final round section */}
            <div className="mt-3 border-t border-border pt-3">
              <button
                type="button"
                className="w-full flex items-center gap-1 px-1 mb-1"
                onClick={() => {
                  pack.finalRound.expanded = !pack.finalRound.expanded;
                  setPack({ ...pack });
                }}
              >
                <ChevronDown
                  size={12}
                  className={`text-on-surface-muted transition-transform duration-150${pack.finalRound.expanded ? "" : " -rotate-90"}`}
                />
                <span className="text-[11px] font-semibold uppercase tracking-widest text-on-surface-muted">
                  Final Round
                </span>
              </button>

              {pack.finalRound.expanded && (
                <div className="flex flex-col gap-0.5">
                  {pack.finalRound.categories.map((cat, i) => (
                    <div key={i} className="group relative flex items-center">
                      <button
                        type="button"
                        className="w-full text-left text-sm truncate pl-3 pr-1 py-1 rounded-md text-on-surface-muted group-hover:bg-surface-raised group-hover:text-on-surface transition-colors duration-150"
                        onClick={() =>
                          setFinalRoundCategoryModal({
                            isOpen: true,
                            category: cat,
                            saveCategory: changeFinalRoundCategory.bind(null, i),
                          })
                        }
                      >
                        {cat.name}
                      </button>
                      {!readOnly && (
                        <button
                          type="button"
                          className={`${dangerIconBtnCls} absolute right-0 transition-opacity opacity-0 group-hover:opacity-100 group-hover:bg-surface-raised rounded-sm`}
                          onClick={() => deleteFinalRoundCategory(i)}
                        >
                          <FaTrashCan size={9} />
                        </button>
                      )}
                    </div>
                  ))}

                  {!readOnly && (
                    <button
                      type="button"
                      className="pl-3 pr-1 py-1 text-xs text-on-surface-muted hover:text-primary flex items-center gap-1 transition-colors duration-150"
                      onClick={() =>
                        setFinalRoundCategoryModal({
                          isOpen: true,
                          category: {
                            name: "",
                            question: {
                              text: "",
                              attachment: { type: "file" },
                              answers: [],
                              comment: null,
                            },
                          },
                          saveCategory: addFinalRoundCategory,
                        })
                      }
                    >
                      <IoIosAdd size={13} />
                      Add category
                    </button>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* ── Main panel ─────────────────────────────────────────── */}
          <div className="flex-1 min-w-0 min-h-0 border-t border-border pt-3 md:border-t-0 md:pt-0 md:border-l md:pl-4 overflow-y-auto">
            {!selectedCategory ? (
              <div className="h-full flex flex-col items-center justify-center gap-2 text-on-surface-muted select-none">
                <p className="text-sm">Select a category from the sidebar</p>
                {!readOnly && pack.rounds.length === 0 && (
                  <button
                    type="button"
                    className="mt-2 inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-md text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
                    onClick={addRound}
                  >
                    <IoIosAdd size={16} /> Add first round
                  </button>
                )}
              </div>
            ) : (
              <CategoryEditor
                key={`${selectedRoundIndex}-${selectedCategoryIndex}`}
                roundIndex={selectedRoundIndex}
                categoryIndex={selectedCategoryIndex}
                category={selectedCategory}
                pack={pack}
                setPack={setPack}
                setQuestionModal={setQuestionModal}
                readOnly={readOnly}
              />
            )}
          </div>
        </div>
      </form>

      <QuestionModal
        isOpen={questionModal.isOpen}
        close={() => setQuestionModal((prev) => ({ ...prev, isOpen: false }))}
        question={questionModal.question}
        saveQuestion={questionModal.saveQuestion}
        readOnly={readOnly}
      />
      <FinalRoundCategoryModal
        isOpen={finalRoundCategoryModal.isOpen}
        close={() =>
          setFinalRoundCategoryModal((prev) => ({ ...prev, isOpen: false }))
        }
        category={finalRoundCategoryModal.category}
        saveCategory={finalRoundCategoryModal.saveCategory}
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
