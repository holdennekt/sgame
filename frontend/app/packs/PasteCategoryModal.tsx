import { CategoryFormData } from "@/types/pack";
import PasteJsonModal from "./PasteJsonModal";

function validate(obj: unknown): CategoryFormData | null {
  if (typeof obj !== "object" || obj === null) return null;
  const c = obj as Record<string, unknown>;
  if (
    typeof c.name !== "string" ||
    typeof c.comment !== "string" ||
    !Array.isArray(c.questions)
  )
    return null;
  return obj as CategoryFormData;
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
  return (
    <PasteJsonModal
      isOpen={isOpen}
      close={close}
      title="Paste category JSON"
      validate={validate}
      onInsert={onInsert}
    />
  );
}
