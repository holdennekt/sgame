import { useRef, useState } from "react";
import { PointerSensor, useSensor, useSensors, DragEndEvent } from "@dnd-kit/core";
import { arrayMove } from "@dnd-kit/sortable";
import {
  convertPackToFormData,
  FinalRoundCategoryFormData,
  Pack,
  PackFormData,
} from "@/types/pack";

export function usePack(initialPack: Omit<Pack, "id" | "createdBy">) {
  const [pack, setPack] = useState<PackFormData>(convertPackToFormData(initialPack));
  const sidebarRef = useRef<HTMLDivElement>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5, delay: 250, tolerance: 5 },
    }),
  );

  const [selectedCategory, selectedCategoryIndex, selectedRoundIndex] = (() => {
    for (const r of pack.rounds) {
      for (const c of r.categories) {
        if (c.selected)
          return [c, r.categories.indexOf(c), pack.rounds.indexOf(r)] as const;
      }
    }
    return [undefined, -1, -1] as const;
  })();

  const addRound = () => {
    pack.rounds.push({
      name: `Round ${pack.rounds.length + 1}`,
      expanded: true,
      categories: [{ name: "Category 1", questions: [], selected: true }],
    });
    if (selectedRoundIndex !== -1)
      pack.rounds[selectedRoundIndex].categories[selectedCategoryIndex].selected = false;
    setPack({ ...pack });
    requestAnimationFrame(() => {
      const rounds = sidebarRef.current?.querySelectorAll("[data-round]");
      rounds?.[rounds.length - 1]?.scrollIntoView({ behavior: "smooth", block: "nearest" });
    });
  };

  const deleteRound = (ri: number) => {
    pack.rounds = pack.rounds.filter((_, i) => i !== ri);
    setPack({ ...pack });
  };

  const renameRound = (ri: number, name: string) => {
    pack.rounds[ri].name = name;
    setPack({ ...pack });
  };

  const toggleRoundExpand = (ri: number) => {
    pack.rounds[ri].expanded = !pack.rounds[ri].expanded;
    setPack({ ...pack });
  };

  const selectCategory = (ri: number, ci: number) => {
    pack.rounds.forEach((r) => r.categories.forEach((c) => (c.selected = false)));
    pack.rounds[ri].categories[ci].selected = true;
    setPack({ ...pack });
  };

  const addCategory = (ri: number) => {
    pack.rounds.forEach((r) => r.categories.forEach((c) => (c.selected = false)));
    pack.rounds[ri].categories.push({
      name: `Category ${pack.rounds[ri].categories.length + 1}`,
      selected: true,
      questions: [],
    });
    pack.rounds[ri].expanded = true;
    setPack({ ...pack });
  };

  const deleteCategory = (ri: number, ci: number) => {
    pack.rounds[ri].categories = pack.rounds[ri].categories.filter((_, i) => i !== ci);
    setPack({ ...pack });
  };

  const addFinalRoundCategory = (category: FinalRoundCategoryFormData) => {
    pack.finalRound.categories.push(category);
    setPack({ ...pack });
  };

  const changeFinalRoundCategory = (index: number, category: FinalRoundCategoryFormData) => {
    pack.finalRound.categories[index] = category;
    setPack({ ...pack });
  };

  const deleteFinalRoundCategory = (index: number) => {
    pack.finalRound.categories = pack.finalRound.categories.filter((_, i) => i !== index);
    setPack({ ...pack });
  };

  const onDragEndRounds = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const oldIndex = pack.rounds.findIndex((_, i) => String(i) === active.id);
    const newIndex = pack.rounds.findIndex((_, i) => String(i) === over.id);
    pack.rounds = arrayMove(pack.rounds, oldIndex, newIndex);
    setPack({ ...pack });
  };

  const onDragEndCategories = (ri: number, event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const categories = pack.rounds[ri].categories;
    const oldIndex = categories.findIndex((_, i) => String(i) === active.id);
    const newIndex = categories.findIndex((_, i) => String(i) === over.id);
    pack.rounds[ri].categories = arrayMove(categories, oldIndex, newIndex);
    setPack({ ...pack });
  };

  return {
    pack,
    setPack,
    sidebarRef,
    sensors,
    selectedCategory,
    selectedCategoryIndex,
    selectedRoundIndex,
    addRound,
    deleteRound,
    renameRound,
    toggleRoundExpand,
    selectCategory,
    addCategory,
    deleteCategory,
    addFinalRoundCategory,
    changeFinalRoundCategory,
    deleteFinalRoundCategory,
    onDragEndRounds,
    onDragEndCategories,
  };
}
