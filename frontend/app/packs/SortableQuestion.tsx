import React from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { QuestionFormData } from "@/types/pack";
import { FaTrashCan } from "react-icons/fa6";
import { RiDraggable } from "react-icons/ri";
import { FiCopy } from "react-icons/fi";

const iconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150";
const dangerIconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-danger hover:bg-surface-raised transition-colors duration-150";

const typeLabel: Record<string, string> = {
  regular: "Regular",
  catInBag: "Cat in bag",
  auction: "Auction",
};

export default function SortableQuestion({
  id,
  question,
  readOnly,
  onOpen,
  onDuplicate,
  onDelete,
}: {
  id: string;
  question: QuestionFormData;
  readOnly: boolean;
  onOpen: () => void;
  onDuplicate: () => void;
  onDelete: () => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id, disabled: readOnly });

  const hasAttachment =
    question.attachment.type === "existing" ||
    (question.attachment.type === "file" && question.attachment.file != null) ||
    (question.attachment.type === "url" && question.attachment.url != null);

  return (
    <div
      ref={setNodeRef}
      style={{
        transform: CSS.Transform.toString(transform ? { ...transform, x: 0 } : null),
        transition: transform ? transition : undefined,
      }}
      className={`group relative flex items-center gap-2 px-4 py-3 rounded-md border border-border hover:bg-surface-raised hover:border-primary transition-colors duration-150 select-none ${isDragging ? "opacity-50 z-50" : ""}`}
    >
      {!readOnly && (
        <button
          type="button"
          className={`${iconBtnCls} sm:opacity-0 sm:group-hover:opacity-100 shrink-0 cursor-grab active:cursor-grabbing touch-none`}
          {...attributes}
          {...listeners}
          onClick={(e) => e.stopPropagation()}
        >
          <RiDraggable size={14} />
        </button>
      )}
      <div className="flex items-center gap-4 flex-1 min-w-0 cursor-pointer" onClick={onOpen}>
        <span
          className={`font-black text-primary w-12 shrink-0 leading-none tabular-nums text-center ${question.value >= 1000 ? "text-lg" : "text-2xl"}`}
        >
          {question.value || "—"}
        </span>
        <div className="flex-1 min-w-0 flex flex-col gap-1">
          <p className="text-sm text-on-surface truncate">
            {question.text || <em className="text-on-surface-muted opacity-50">No text</em>}
          </p>
          <div className="flex items-center gap-1.5 text-[10px] text-on-surface-muted">
            <span className="px-1.5 py-0.5 rounded-full bg-surface-raised font-medium border border-border">
              {typeLabel[question.type] ?? question.type}
            </span>
            {[
              question.answers.length > 0 &&
                `${question.answers.length} answer${question.answers.length !== 1 ? "s" : ""}`,
              hasAttachment && "attachment",
              question.comment && "comment",
            ]
              .filter(Boolean)
              .map((item, i) => (
                <React.Fragment key={i}>
                  <span>·</span>
                  <span>{item}</span>
                </React.Fragment>
              ))}
          </div>
        </div>
      </div>
      {!readOnly && (
        <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 shrink-0">
          <button
            type="button"
            className={iconBtnCls}
            title="Duplicate question"
            onClick={(e) => { e.stopPropagation(); onDuplicate(); }}
          >
            <FiCopy size={12} />
          </button>
          <button
            type="button"
            className={dangerIconBtnCls}
            onClick={(e) => { e.stopPropagation(); onDelete(); }}
          >
            <FaTrashCan size={12} />
          </button>
        </div>
      )}
    </div>
  );
}
