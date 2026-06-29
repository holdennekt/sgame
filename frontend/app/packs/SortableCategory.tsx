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
  category,
  selected,
  readOnly,
  onSelect,
  onDuplicate,
  onDelete,
}: {
  id: string;
  category: CategoryFormData;
  selected: boolean;
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
      className={`group flex items-center gap-1.5 rounded-md cursor-pointer text-sm pl-5 py-1 ${
        isDragging ? "opacity-0" : ""
      } ${
        selected
          ? "bg-surface-raised text-primary font-medium"
          : "text-on-surface-muted hover:bg-surface-raised hover:text-on-surface"
      }`}
      onClick={onSelect}
    >
      <span className="truncate flex-1 min-w-0">
        {category.name || <em className="opacity-50">Unnamed</em>}
      </span>
      <span className="shrink-0 text-[10px] font-medium px-1 rounded leading-4">
        {category.questions.length}
      </span>
      {!readOnly && (
        <div className="flex items-center gap-0.5 shrink-0 overflow-hidden transition-[max-width] duration-150 can-hover:max-w-0 group-hover:max-w-20">
          <button
            type="button"
            title="Duplicate category"
            className={iconBtnCls}
            onClick={(e) => {
              e.stopPropagation();
              onDuplicate();
            }}
          >
            <FiCopy size={10} />
          </button>
          <button
            type="button"
            className={`${iconBtnCls} cursor-grab active:cursor-grabbing touch-none`}
            {...attributes}
            {...listeners}
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
