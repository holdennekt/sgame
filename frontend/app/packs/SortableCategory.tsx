import React from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { CategoryFormData } from "@/types/pack";
import { FaTrashCan } from "react-icons/fa6";
import { RiDraggable } from "react-icons/ri";
import { FiCopy } from "react-icons/fi";

const iconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150";
const dangerIconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-danger hover:bg-surface-raised transition-colors duration-150";

export default function SortableCategory({
  id,
  cat,
  readOnly,
  onSelect,
  onDuplicate,
  onDelete,
}: {
  id: string;
  cat: CategoryFormData;
  readOnly: boolean;
  onSelect: () => void;
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
      className={`group relative flex items-center ${
        isDragging ? "opacity-50 z-50" : ""
      }`}
    >
      <button
        type="button"
        className={`w-full text-left text-sm truncate pl-5 pr-1 py-1 rounded-md ${
          cat.selected
            ? "bg-surface-raised text-primary font-medium"
            : "text-on-surface-muted group-hover:bg-surface-raised group-hover:text-on-surface"
        }`}
        onClick={onSelect}
      >
        {cat.name || <em className="opacity-50">Unnamed</em>}
      </button>
      {!readOnly && (
        <div className="absolute right-0 flex items-center gap-0.5 sm:opacity-0 sm:group-hover:opacity-100 sm:group-hover:bg-surface-raised transition-opacity shrink-0 rounded-sm">
          <button
            type="button"
            title="Duplicate category"
            className={iconBtnCls}
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
            <FaTrashCan size={10} />
          </button>
        </div>
      )}
    </div>
  );
}
