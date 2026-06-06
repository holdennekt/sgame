"use client";

import { useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import QuestionModal from "./QuestionModal";
import FinalRoundCategoryModal from "./FinalRoundCategoryModal";
import CategoryEditor from "./CategoryEditor";
import SortableRound from "./SortableRound";
import SortableCategory from "./SortableCategory";
import { toast, ToastContainer } from "react-toastify";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { FaTrashCan } from "react-icons/fa6";
import { FiArrowLeft, FiCopy } from "react-icons/fi";
import { RiDraggable } from "react-icons/ri";
import {
  IoIosAdd,
  IoIosArrowDown as ChevronDown,
  IoIosArrowDown as SelectArrow,
} from "react-icons/io";
import { DndContext, closestCenter } from "@dnd-kit/core";
import {
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import {
  FinalRoundCategoryFormData,
  PackFormData,
  QuestionFormData,
  convertPackFormDataToRequest,
} from "@/types/pack";
import { signURL, createDraft, updateDraft, publishDraft } from "@/app/api";
import { isError, isValidationErrors, ValidationError } from "@/middleware";
import { usePack } from "@/hooks/usePack";
import { convertDraftToFormData } from "@/types/pack_draft";

function parsePath(path: string): string[] {
  return path.split(/[.\[\]]+/).filter(Boolean);
}

function formatErrorPath(path: string): string {
  return path
    .replace(/^name$/, "Pack name")
    .replace(/^type$/, "Pack type")
    .replace(
      /^finalRound\.categories\[(\d+)\]/,
      (_, i) => `Final Round, Category ${+i + 1}`
    )
    .replace(/^rounds\[(\d+)\]/, (_, i) => `Round ${+i + 1}`)
    .replace(/\.categories\[(\d+)\]/, (_, i) => `, Category ${+i + 1}`)
    .replace(/\.questions\[(\d+)\]/, (_, i) => `, Question ${+i + 1}`)
    .replace(
      /\.(name|text|type|comment|answers|value|attachment|categories|questions)$/,
      (_, f) => ` › ${f.charAt(0).toUpperCase() + f.slice(1)}`
    );
}

const selectCls =
  "h-9 pl-2.5 pr-8 bg-background border border-border text-on-background rounded-lg text-sm outline-none focus-ring appearance-none";
const iconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150";
const dangerIconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-danger hover:bg-surface-raised transition-colors duration-150";

type SaveStatus =
  | { type: "saving" }
  | { type: "saved"; at: Date }
  | { type: "error" }
  | null;

function SortableFinalCategoryItem({
  id,
  cat,
  readOnly,
  onOpen,
  onDuplicate,
  onDelete,
}: {
  id: string;
  cat: FinalRoundCategoryFormData;
  readOnly: boolean;
  onOpen: () => void;
  onDuplicate: () => void;
  onDelete: () => void;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id, disabled: readOnly });

  return (
    <div
      ref={setNodeRef}
      style={{
        transform: CSS.Transform.toString(
          transform ? { ...transform, x: 0 } : null
        ),
        transition: transform ? transition : undefined,
      }}
      className={`group flex items-center gap-1.5 rounded-md cursor-pointer text-sm pl-3 py-1 text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150${
        isDragging ? " opacity-50 z-50" : ""
      }`}
      onClick={onOpen}
    >
      <span className="truncate flex-1 min-w-0">{cat.name}</span>
      {!readOnly && (
        <div className="flex items-center gap-0.5 shrink-0 overflow-hidden transition-[max-width] duration-150 can-hover:max-w-0 group-hover:max-w-20">
          <button
            type="button"
            className={iconBtnCls}
            title="Duplicate"
            onClick={onDuplicate}
          >
            <FiCopy size={10} />
          </button>
          <button
            type="button"
            className={`${iconBtnCls} cursor-grab active:cursor-grabbing touch-none`}
            {...attributes}
            {...listeners}
            onClick={(e) => e.stopPropagation()}
          >
            <RiDraggable size={11} />
          </button>
          <button type="button" className={dangerIconBtnCls} onClick={onDelete}>
            <FaTrashCan size={9} />
          </button>
        </div>
      )}
    </div>
  );
}

export default function PackEditor({
  initialPack,
  readOnly = false,
  backHref,
  draftId,
  packId,
}: {
  initialPack: PackFormData;
  readOnly?: boolean;
  backHref?: string;
  draftId?: string;
  packId?: string;
}) {
  const router = useRouter();
  const queryClient = useQueryClient();

  const {
    pack,
    setPack,
    sidebarRef,
    sensors,
    selectedCategory,
    selectedCategoryIndex,
    selectedRoundIndex,
    expandedRounds,
    finalRoundExpanded,
    setFinalRoundExpanded,
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
    duplicateFinalRoundCategory,
    deleteFinalRoundCategory,
    onDragEndRounds,
    onDragEndCategories,
    onDragEndFinalRoundCategories,
  } = usePack(initialPack);

  const [questionModal, setQuestionModal] = useState<{
    isOpen: boolean;
    question: QuestionFormData;
    saveQuestion: (q: Omit<QuestionFormData, "index">) => void;
    validationError?: string;
  }>({
    isOpen: false,
    question: {
      index: 0,
      value: 0,
      text: "",
      attachment: { type: "file" },
      type: "regular",
      answers: [],
      comment: { text: "", attachment: { type: "file" } },
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
        comment: { text: "", attachment: { type: "file" } },
      },
    },
    saveCategory: () => {},
  });

  const [saveStatus, setSaveStatus] = useState<SaveStatus>(null);
  const [isPublishing, setIsPublishing] = useState(false);
  const [saveErrors, setSaveErrors] = useState<ValidationError[]>([]);
  const [publishErrors, setPublishErrors] = useState<ValidationError[]>([]);
  const validationErrors = saveErrors.concat(publishErrors);
  const mounted = useRef(false);
  const nameInputRef = useRef<HTMLInputElement>(null);
  const autosaveTimer = useRef<ReturnType<typeof setTimeout>>();
  const serverSync = useRef(false);

  useEffect(() => {
    if (!mounted.current) {
      mounted.current = true;
      if (draftId) setSaveStatus({ type: "saved", at: new Date() });
      return;
    }
    if (!draftId) return;
    if (serverSync.current) {
      serverSync.current = false;
      return;
    }

    setSaveStatus({ type: "saving" });
    autosaveTimer.current = setTimeout(() => saveDraft(pack), 1000);
    return () => clearTimeout(autosaveTimer.current);
  }, [pack, draftId]);

  const handleEdit = async () => {
    try {
      const { id } = await createDraft(packId);
      queryClient.invalidateQueries({ queryKey: ["drafts"] });
      router.push(`/packs/drafts/${id}`);
    } catch (e) {
      toast.error(isError(e) ? e.error : "Failed to create draft", {
        containerId: "editor",
      });
    }
  };

  const saveDraft = async (packToSave: PackFormData) => {
    if (!draftId) return;
    setSaveStatus({ type: "saving" });
    try {
      const req = await convertPackFormDataToRequest(packToSave, signURL);
      const newDraft = await updateDraft(draftId, req);
      serverSync.current = true;
      setPack(convertDraftToFormData(newDraft));
      setSaveErrors([]);
      setSaveStatus({ type: "saved", at: new Date() });
    } catch (e) {
      if (isValidationErrors(e)) setSaveErrors(e.errors);
      setSaveStatus({ type: "error" });
    }
  };

  const handlePublish = async () => {
    if (!draftId) return;
    setIsPublishing(true);
    try {
      const req = await convertPackFormDataToRequest(pack, signURL);
      await updateDraft(draftId, req);
      serverSync.current = true;
      setSaveErrors([]);
    } catch (e) {
      if (isValidationErrors(e)) {
        setPublishErrors(e.errors);
        setSaveStatus({ type: "error" });
      } else {
        toast.error(isError(e) ? e.error : "Save failed", {
          containerId: "editor",
        });
      }
      setIsPublishing(false);
      return;
    }
    try {
      const { id } = await publishDraft(draftId);
      setPublishErrors([]);
      queryClient.invalidateQueries({ queryKey: ["packs"] });
      queryClient.invalidateQueries({ queryKey: ["drafts"] });
      router.refresh();
      router.push(`/packs/${id}`);
    } catch (e) {
      if (isValidationErrors(e)) {
        setPublishErrors(e.errors);
      } else {
        toast.error(isError(e) ? e.error : "Publish failed", {
          containerId: "editor",
        });
      }
      setIsPublishing(false);
    }
  };

  const navigateToError = ({ path, message }: ValidationError) => {
    const segments = parsePath(path);
    const [root, ...rest] = segments;

    if (root === "name" || root === "type") {
      nameInputRef.current?.focus();
      return;
    }

    if (root === "rounds") {
      const ri = parseInt(rest[0]);
      if (!expandedRounds[ri]) toggleRoundExpand(ri);

      if (rest[1] === "name") return;

      if (rest[1] === "categories") {
        const ci = parseInt(rest[2]);
        if (isNaN(ci)) return;
        selectCategory(ri, ci);
        if (rest[3] === "questions") {
          const qi = parseInt(rest[4]);
          const question = pack.rounds[ri]?.categories[ci]?.questions[qi];
          if (question) {
            const field = rest[5];
            const fieldLabel =
              typeof field === "string"
                ? field.charAt(0).toUpperCase() + field.slice(1)
                : null;
            setQuestionModal({
              isOpen: true,
              question,
              validationError: fieldLabel
                ? `${fieldLabel}: ${message}`
                : message,
              saveQuestion: (q) => {
                const newPack = { ...pack };
                newPack.rounds[ri].categories[ci].questions[qi] = {
                  ...q,
                  index: qi,
                };
                setPack(newPack);
                clearTimeout(autosaveTimer.current);
                saveDraft(newPack);
              },
            });
          }
        }
      }
      return;
    }

    if (root === "finalRound" && rest[0] === "categories") {
      setFinalRoundExpanded(true);
      const i = parseInt(rest[1]);
      const cat = pack.finalRound.categories[i];
      if (cat)
        setFinalRoundCategoryModal({
          isOpen: true,
          category: cat,
          saveCategory: changeFinalRoundCategory.bind(null, i),
        });
    }
  };

  const actionButton = readOnly ? (
    packId && (
      <button
        type="button"
        className="w-full sm:w-auto inline-flex items-center justify-center px-3.5 py-1.5 rounded-md text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
        onClick={handleEdit}
      >
        Edit
      </button>
    )
  ) : (
    <button
      type="button"
      className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-3.5 py-1.5 rounded-md text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 disabled:opacity-60 disabled:cursor-not-allowed"
      onClick={handlePublish}
      disabled={isPublishing}
    >
      {isPublishing && (
        <svg
          className="animate-spin h-3.5 w-3.5"
          viewBox="0 0 24 24"
          fill="none"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"
          />
        </svg>
      )}
      {isPublishing ? "Publishing…" : "Publish"}
    </button>
  );

  return (
    <>
      <div className="min-h-0 h-full flex flex-col gap-3">
        <div className="flex flex-col gap-1.5 shrink-0">
          {/* Back navigation + save status */}
          {backHref && (
            <div className="flex items-center justify-between">
              <Link
                href={backHref}
                className="flex items-center gap-1 text-sm text-on-surface-muted hover:text-on-surface transition-colors duration-150"
              >
                <FiArrowLeft size={14} />
                {readOnly ? "Packs" : "Drafts"}
              </Link>
              {!readOnly && saveStatus && (
                <span className="text-xs text-on-surface-muted">
                  {saveStatus.type === "saving" && "Saving…"}
                  {saveStatus.type === "saved" &&
                    `Saved at ${saveStatus.at.toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                    })}`}
                  {saveStatus.type === "error" && (
                    <span className="text-danger">Error saving</span>
                  )}
                </span>
              )}
            </div>
          )}

          {/* Title + Actions */}
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
            <input
              ref={nameInputRef}
              className={`flex-1 min-w-0 text-lg font-semibold bg-transparent outline-none placeholder:text-on-surface-muted text-on-surface py-0.5 transition-[border-color] duration-150 border-b ${
                readOnly
                  ? "border-transparent cursor-default"
                  : "border-border hover:border-primary/60 focus:border-primary cursor-text"
              }`}
              type="text"
              placeholder="Pack name"
              value={pack.name}
              onChange={(e) => setPack({ ...pack, name: e.target.value })}
              readOnly={readOnly}
            />

            <div className="grid grid-cols-2 gap-2 w-full sm:flex sm:w-auto sm:items-center">
              {readOnly ? (
                <span className="h-8 px-2.5 flex items-center justify-center bg-surface-raised border border-border text-on-surface-muted rounded-lg text-sm capitalize">
                  {pack.type}
                </span>
              ) : (
                <div className="relative">
                  <select
                    className={`${selectCls} w-full`}
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
              {actionButton}
            </div>
          </div>
        </div>

        {validationErrors.length > 0 && (
          <button
            type="button"
            className="shrink-0 w-full text-left rounded-md border border-danger/30 bg-danger/10 px-3 py-2 flex items-center gap-3 hover:bg-danger/15 transition-colors duration-150"
            onClick={() => {
              navigateToError(validationErrors[0]);
              if (!saveErrors.length) setPublishErrors((prev) => prev.slice(1));
              else setSaveErrors((prev) => prev.slice(1));
            }}
          >
            <p className="text-xs text-danger flex-1">
              <span className="font-medium">
                {formatErrorPath(validationErrors[0].path)}
              </span>
              {": "}
              {validationErrors[0].message}
            </p>
            {validationErrors.length > 1 && (
              <span className="text-xs text-danger/60 shrink-0">
                {validationErrors.length - 1} more
              </span>
            )}
          </button>
        )}

        <div className="h-px bg-border shrink-0" />

        {/* ── Body ─────────────────────────────────────────────────── */}
        <div className="flex-1 flex flex-col md:flex-row gap-3 min-h-0">
          {/* ── Sidebar ────────────────────────────────────────────── */}
          <div
            ref={sidebarRef}
            className="w-full md:w-52 shrink-0 flex flex-col gap-3 min-h-0 max-h-48 md:max-h-none overflow-y-auto"
          >
            <div className="flex flex-col gap-1">
              {/* Rounds header */}
              <div className="flex items-center justify-between px-1">
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
                        expanded={expandedRounds[ri]}
                        readOnly={readOnly}
                        onToggleExpand={() => toggleRoundExpand(ri)}
                        onRename={(name) => renameRound(ri, name)}
                        onDuplicate={() => duplicateRound(ri)}
                        onDelete={() => deleteRound(ri)}
                        onAddRound={addRound}
                      >
                        {expandedRounds[ri] && (
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
                                items={round.categories.map((_, i) =>
                                  String(i)
                                )}
                                strategy={verticalListSortingStrategy}
                              >
                                {round.categories.map((cat, ci) => (
                                  <SortableCategory
                                    key={ci}
                                    id={String(ci)}
                                    category={cat}
                                    selected={
                                      selectedRoundIndex === ri &&
                                      selectedCategoryIndex === ci
                                    }
                                    readOnly={readOnly}
                                    onSelect={() => selectCategory(ri, ci)}
                                    onDuplicate={() =>
                                      duplicateCategory(ri, ci)
                                    }
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
            </div>

            <div className="h-px bg-border shrink-0" />

            {/* Final round section */}
            <div>
              <button
                type="button"
                className="w-full flex items-center gap-1 px-1 mb-1"
                onClick={() => setFinalRoundExpanded(!finalRoundExpanded)}
              >
                <ChevronDown
                  size={12}
                  className={`text-on-surface-muted transition-transform duration-150${
                    finalRoundExpanded ? "" : " -rotate-90"
                  }`}
                />
                <span className="text-[11px] font-semibold uppercase tracking-widest text-on-surface-muted">
                  Final Round
                </span>
              </button>

              {finalRoundExpanded && (
                <div className="flex flex-col gap-0.5">
                  <DndContext
                    sensors={sensors}
                    collisionDetection={closestCenter}
                    onDragEnd={onDragEndFinalRoundCategories}
                  >
                    <SortableContext
                      items={pack.finalRound.categories.map((_, i) =>
                        String(i)
                      )}
                      strategy={verticalListSortingStrategy}
                    >
                      {pack.finalRound.categories.map((cat, i) => (
                        <SortableFinalCategoryItem
                          key={i}
                          id={String(i)}
                          cat={cat}
                          readOnly={readOnly}
                          onOpen={() =>
                            setFinalRoundCategoryModal({
                              isOpen: true,
                              category: cat,
                              saveCategory: changeFinalRoundCategory.bind(
                                null,
                                i
                              ),
                            })
                          }
                          onDuplicate={() => duplicateFinalRoundCategory(i)}
                          onDelete={() => deleteFinalRoundCategory(i)}
                        />
                      ))}
                    </SortableContext>
                  </DndContext>

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
                              comment: {
                                text: "",
                                attachment: { type: "file" },
                              },
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

          <div className="h-px md:h-auto md:w-px self-stretch bg-border" />

          {/* ── Main panel ─────────────────────────────────────────── */}
          <div className="flex-1 min-w-0 min-h-0 overflow-y-auto">
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
      </div>

      <QuestionModal
        isOpen={questionModal.isOpen}
        close={() => setQuestionModal((prev) => ({ ...prev, isOpen: false }))}
        question={questionModal.question}
        saveQuestion={questionModal.saveQuestion}
        validationError={questionModal.validationError}
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
