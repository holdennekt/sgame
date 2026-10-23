import React from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { RoundFormData } from "@/types/pack";
import { FaTrashCan } from "react-icons/fa6";
import { RiDraggable } from "react-icons/ri";
import { FiCopy } from "react-icons/fi";
import { IoIosArrowDown as ChevronDown } from "react-icons/io";

const iconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150";
const dangerIconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-danger hover:bg-surface-raised transition-colors duration-150";

export default function SortableRound({
  id,
  round,
  expanded,
  readOnly,
  processing,
  children,
  onToggleExpand,
  onRename,
  onDuplicate,
  onDelete,
}: {
  id: string;
  round: RoundFormData;
  expanded: boolean;
  readOnly: boolean;
  processing: boolean;
  children: React.ReactNode;
  onToggleExpand: () => void;
  onRename: (name: string) => void;
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
      className={isDragging ? "opacity-50 relative z-50" : ""}
      data-round
    >
      <div className="group flex items-center gap-0.5 rounded-md pl-0.5 pr-0.5 hover:bg-surface-raised transition-colors duration-150">
        <button
          type="button"
          className="p-0.5 text-on-surface-muted shrink-0"
          aria-label="Toggle round"
          onClick={onToggleExpand}
        >
          <ChevronDown
            size={12}
            className={`transition-transform duration-150${
              expanded ? "" : " -rotate-90"
            }`}
          />
        </button>
        <input
          className="flex-1 min-w-0 bg-transparent text-sm font-medium text-on-surface py-1.5 outline-none placeholder:text-on-surface-muted"
          value={round.name}
          placeholder="Round name"
          maxLength={100}
          onChange={(e) => onRename(e.target.value)}
          readOnly={readOnly}
        />
        {processing && (
          <span
            className={`shrink-0 border-2 border-primary border-t-transparent rounded-full animate-spin${
              readOnly ? "" : " can-hover:group-hover:hidden"
            }`}
            style={{ width: 12, height: 12 }}
          />
        )}
        {!readOnly && (
          <div className="flex items-center gap-0.5 shrink-0 overflow-hidden transition-[max-width] duration-150 can-hover:max-w-0 group-hover:max-w-28">
            <button
              type="button"
              title="Duplicate round"
              aria-label="Duplicate round"
              className={iconBtnCls}
              onClick={onDuplicate}
            >
              <FiCopy size={11} />
            </button>
            <button
              type="button"
              className={`${iconBtnCls} cursor-grab active:cursor-grabbing touch-none`}
              aria-label="Reorder round"
              {...attributes}
              {...listeners}
              onClick={(e) => e.stopPropagation()}
            >
              <RiDraggable size={13} />
            </button>
            <button
              type="button"
              className={dangerIconBtnCls}
              title="Delete round"
              aria-label="Delete round"
              onClick={onDelete}
            >
              <FaTrashCan size={10} />
            </button>
          </div>
        )}
      </div>
      {children}
    </div>
  );
}
