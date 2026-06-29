import { useEffect, useState } from "react";
import Modal from "@/components/Modal";
import { CategoryFormData } from "@/types/pack";

function isCategoryFormData(obj: unknown): obj is CategoryFormData {
  if (typeof obj !== "object" || obj === null) return false;
  const c = obj as Record<string, unknown>;
  return (
    typeof c.name === "string" &&
    typeof c.comment === "string" &&
    Array.isArray(c.questions)
  );
}

export default function PasteCategoryModal({
  isOpen,
  close,
  onInsert,
}: {
  isOpen: boolean;
  close: () => void;
  onInsert: (category: CategoryFormData) => void;
}) {
  const [json, setJson] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isOpen) return;
    setJson("");
    setError(null);
    navigator.clipboard
      .readText()
      .then((text) => {
        try {
          setJson(JSON.stringify(JSON.parse(text), null, 2));
        } catch {
          // not JSON, leave textarea empty
        }
      })
      .catch(() => {});
  }, [isOpen]);

  const onInsertClick = () => {
    try {
      const parsed: unknown = JSON.parse(json);
      if (!isCategoryFormData(parsed)) {
        setError("Not a valid category JSON");
        return;
      }
      onInsert(parsed);
      close();
    } catch {
      setError("Invalid JSON");
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={close} closeByClickingOutside={true}>
      <h3 className="text-base font-semibold text-on-surface mb-4">
        Paste category JSON
      </h3>
      <textarea
        className="w-[480px] max-w-[80vw] h-64 px-2.5 py-2 bg-background border border-border text-on-background rounded-lg text-xs font-mono outline-none focus-ring placeholder:text-on-surface-muted resize-none transition-[border-color] duration-150"
        placeholder="Paste category JSON here…"
        value={json}
        onChange={(e) => {
          setJson(e.target.value);
          setError(null);
        }}
        spellCheck={false}
      />
      <div className="mt-4 flex items-center justify-between gap-4">
        <button
          type="button"
          className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium border border-border text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150"
          onClick={close}
        >
          Cancel
        </button>
        <div className="flex items-center gap-2">
          {error && <p className="text-xs text-danger">{error}</p>}
          <button
            type="button"
            className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 shrink-0"
            onClick={onInsertClick}
          >
            Insert
          </button>
        </div>
      </div>
    </Modal>
  );
}
