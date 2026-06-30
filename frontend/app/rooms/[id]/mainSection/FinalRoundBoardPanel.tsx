import React from "react";
import { FiX } from "react-icons/fi";

export default function FinalRoundBoardPanel({
  availableCategories,
  canRemoveCategory,
  removeCategory,
  isPaused = false,
}: {
  availableCategories: Record<string, boolean>;
  canRemoveCategory: boolean;
  removeCategory: (category: string) => void;
  isPaused?: boolean;
}) {
  return (
    <div className="w-full h-full flex flex-col items-center justify-center gap-2 p-4 overflow-y-auto">
      {Object.entries(availableCategories).map(([category, available]) => (
        <button
          key={category}
          type="button"
          disabled={!available || !canRemoveCategory || isPaused}
          onClick={() =>
            available &&
            canRemoveCategory &&
            !isPaused &&
            removeCategory(category)
          }
          className={`group w-full max-w-sm flex items-center justify-between gap-2 px-4 py-2.5 rounded-lg border text-sm font-medium transition-all duration-150 ${
            available
              ? isPaused
                ? "border-border text-on-surface opacity-50 cursor-default"
                : canRemoveCategory
                ? "border-border text-on-surface hover:border-danger hover:text-danger hover:bg-surface-raised cursor-pointer"
                : "border-border text-on-surface cursor-default"
              : "border-border/40 text-on-surface-muted line-through opacity-40 cursor-default"
          }`}
        >
          <span className="flex-1 text-center">{category}</span>
          {available && canRemoveCategory && !isPaused && (
            <FiX
              size={13}
              className="opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
            />
          )}
        </button>
      ))}
    </div>
  );
}
