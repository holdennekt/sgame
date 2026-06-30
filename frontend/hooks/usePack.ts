import { useCallback, useRef, useState } from "react";
import {
  closestCenter,
  DragEndEvent,
  DragOverEvent,
  DragStartEvent,
  PointerSensor,
  UniqueIdentifier,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { arrayMove } from "@dnd-kit/sortable";
import {
  CategoryFormData,
  FinalRoundCategoryFormData,
  PackFormData,
} from "@/types/pack";

const genId = () => Math.random().toString(36).slice(2);

export function usePack(initialPack: PackFormData) {
  const [pack, setPack] = useState(initialPack);

  const [expandedRounds, setExpandedRounds] = useState<boolean[]>(() =>
    initialPack.rounds.map(() => true)
  );
  const [finalRoundExpanded, setFinalRoundExpanded] = useState(true);
  const [selectedRI, setSelectedRI] = useState(() =>
    initialPack.rounds.length > 0 && initialPack.rounds[0].categories.length > 0
      ? 0
      : -1
  );
  const [selectedCI, setSelectedCI] = useState(() =>
    initialPack.rounds.length > 0 && initialPack.rounds[0].categories.length > 0
      ? 0
      : -1
  );

  // Stable IDs for categories, mirrors pack.rounds[ri].categories[ci]
  const [categoryStableIds, setCategoryStableIds] = useState<string[][]>(() =>
    initialPack.rounds.map((round) => round.categories.map(() => genId()))
  );

  const sidebarRef = useRef<HTMLDivElement>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5, delay: 250, tolerance: 5 },
    })
  );

  const selectedCategory =
    selectedRI >= 0 && selectedCI >= 0
      ? pack.rounds[selectedRI]?.categories[selectedCI]
      : undefined;

  const addRound = () => {
    const newRi = pack.rounds.length;
    pack.rounds.push({
      name: `Round ${pack.rounds.length + 1}`,
      categories: [{ name: "Category 1", comment: "", questions: [] }],
    });
    setExpandedRounds((prev) => [...prev, true]);
    setCategoryStableIds((prev) => [...prev, [genId()]]);
    setSelectedRI(newRi);
    setSelectedCI(0);
    setPack({ ...pack });
    requestAnimationFrame(() => {
      const rounds = sidebarRef.current?.querySelectorAll("[data-round]");
      rounds?.[rounds.length - 1]?.scrollIntoView({
        behavior: "smooth",
        block: "nearest",
      });
    });
  };

  const duplicateRound = (ri: number) => {
    const src = pack.rounds[ri];
    const copy = {
      ...src,
      name: `${src.name} (copy)`,
      categories: src.categories.map((cat) => ({
        ...cat,
        questions: cat.questions.map((q) => ({
          ...q,
          answers: [...q.answers],
          attachment: { ...q.attachment },
        })),
      })),
    };
    pack.rounds.splice(ri + 1, 0, copy);
    setExpandedRounds((prev) => {
      const next = [...prev];
      next.splice(ri + 1, 0, prev[ri]);
      return next;
    });
    setCategoryStableIds((prev) => {
      const next = [...prev];
      next.splice(
        ri + 1,
        0,
        prev[ri].map(() => genId())
      );
      return next;
    });
    if (selectedRI > ri) setSelectedRI(selectedRI + 1);
    setPack({ ...pack });
  };

  const deleteRound = (ri: number) => {
    pack.rounds = pack.rounds.filter((_, i) => i !== ri);
    setExpandedRounds((prev) => prev.filter((_, i) => i !== ri));
    setCategoryStableIds((prev) => prev.filter((_, i) => i !== ri));
    if (selectedRI === ri) {
      setSelectedRI(-1);
      setSelectedCI(-1);
    } else if (selectedRI > ri) {
      setSelectedRI(selectedRI - 1);
    }
    setPack({ ...pack });
  };

  const renameRound = (ri: number, name: string) => {
    pack.rounds[ri].name = name;
    setPack({ ...pack });
  };

  const toggleRoundExpand = (ri: number) => {
    setExpandedRounds((prev) => prev.map((e, i) => (i === ri ? !e : e)));
  };

  const selectCategory = (ri: number, ci: number) => {
    setSelectedRI(ri);
    setSelectedCI(ci);
  };

  const addCategory = (ri: number) => {
    const ci = pack.rounds[ri].categories.length;
    pack.rounds[ri].categories.push({
      name: `Category ${ci + 1}`,
      comment: "",
      questions: [],
    });
    setExpandedRounds((prev) => prev.map((e, i) => (i === ri ? true : e)));
    setCategoryStableIds((prev) =>
      prev.map((ids, i) => (i === ri ? [...ids, genId()] : ids))
    );
    setSelectedRI(ri);
    setSelectedCI(ci);
    setPack({ ...pack });
  };

  const duplicateCategory = (ri: number, ci: number) => {
    const src = pack.rounds[ri].categories[ci];
    const copy = {
      ...src,
      name: `${src.name} (copy)`,
      questions: src.questions.map((q) => ({
        ...q,
        answers: [...q.answers],
        attachment: { ...q.attachment },
      })),
    };
    pack.rounds[ri].categories.splice(ci + 1, 0, copy);
    setCategoryStableIds((prev) =>
      prev.map((ids, i) => {
        if (i !== ri) return ids;
        const next = [...ids];
        next.splice(ci + 1, 0, genId());
        return next;
      })
    );
    if (selectedRI === ri && selectedCI > ci) setSelectedCI(selectedCI + 1);
    setPack({ ...pack });
  };

  const deleteCategory = (ri: number, ci: number) => {
    pack.rounds[ri].categories = pack.rounds[ri].categories.filter(
      (_, i) => i !== ci
    );
    setCategoryStableIds((prev) =>
      prev.map((ids, i) => (i === ri ? ids.filter((_, j) => j !== ci) : ids))
    );
    if (selectedRI === ri) {
      if (selectedCI === ci) {
        setSelectedRI(-1);
        setSelectedCI(-1);
      } else if (selectedCI > ci) {
        setSelectedCI(selectedCI - 1);
      }
    }
    setPack({ ...pack });
  };

  const addCategoryFromJson = (ri: number, category: CategoryFormData) => {
    const ci = pack.rounds[ri].categories.length;
    pack.rounds[ri].categories.push(category);
    setExpandedRounds((prev) => prev.map((e, i) => (i === ri ? true : e)));
    setCategoryStableIds((prev) =>
      prev.map((ids, i) => (i === ri ? [...ids, genId()] : ids))
    );
    setSelectedRI(ri);
    setSelectedCI(ci);
    setPack({ ...pack });
  };

  const addFinalRoundCategory = (category: FinalRoundCategoryFormData) => {
    pack.finalRound.categories.push(category);
    setPack({ ...pack });
  };

  const changeFinalRoundCategory = (
    index: number,
    category: FinalRoundCategoryFormData
  ) => {
    pack.finalRound.categories[index] = category;
    setPack({ ...pack });
  };

  const duplicateFinalRoundCategory = (index: number) => {
    const src = pack.finalRound.categories[index];
    const copy: FinalRoundCategoryFormData = {
      name: `${src.name} (copy)`,
      question: {
        ...src.question,
        answers: [...src.question.answers],
        attachment: { ...src.question.attachment },
        comment: {
          ...src.question.comment,
          attachment: { ...src.question.comment.attachment },
        },
      },
    };
    pack.finalRound.categories.splice(index + 1, 0, copy);
    setPack({ ...pack });
  };

  const deleteFinalRoundCategory = (index: number) => {
    pack.finalRound.categories = pack.finalRound.categories.filter(
      (_, i) => i !== index
    );
    setPack({ ...pack });
  };

  const onDragEndFinalRoundCategories = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const cats = pack.finalRound.categories;
    const oldIndex = cats.findIndex((_, i) => String(i) === active.id);
    const newIndex = cats.findIndex((_, i) => String(i) === over.id);
    pack.finalRound.categories = arrayMove(cats, oldIndex, newIndex);
    setPack({ ...pack });
  };

  const remapIndex = (idx: number, oldIndex: number, newIndex: number) => {
    if (idx === oldIndex) return newIndex;
    if (oldIndex < newIndex && idx > oldIndex && idx <= newIndex)
      return idx - 1;
    if (oldIndex > newIndex && idx >= newIndex && idx < oldIndex)
      return idx + 1;
    return idx;
  };

  const [activeId, setActiveId] = useState<UniqueIdentifier | null>(null);

  const onDragStart = ({ active }: DragStartEvent) => setActiveId(active.id);

  const collisionDetection = useCallback(
    (args: Parameters<typeof closestCenter>[0]) => {
      const isRound =
        activeId !== null && String(activeId).startsWith("round-");
      return closestCenter({
        ...args,
        droppableContainers: args.droppableContainers.filter((c) =>
          isRound
            ? String(c.id).startsWith("round-")
            : !String(c.id).startsWith("round-")
        ),
      });
    },
    [activeId]
  );

  // Always-fresh ref so onDragOver callback doesn't need deps
  const dragStateRef = useRef({
    categoryStableIds,
    pack,
    selectedRI,
    selectedCI,
  });
  dragStateRef.current = { categoryStableIds, pack, selectedRI, selectedCI };

  const onDragOver = useCallback(
    (event: DragOverEvent) => {
      const { active, over } = event;
      if (!over) return;
      if (String(active.id).startsWith("round-")) return;

      const { categoryStableIds, pack, selectedRI, selectedCI } =
        dragStateRef.current;

      // Find source
      let srcRi = -1,
        srcCi = -1;
      outer: for (let ri = 0; ri < categoryStableIds.length; ri++) {
        for (let ci = 0; ci < categoryStableIds[ri].length; ci++) {
          if (categoryStableIds[ri][ci] === String(active.id)) {
            srcRi = ri;
            srcCi = ci;
            break outer;
          }
        }
      }
      if (srcRi === -1) return;

      // Find destination
      let dstRi = -1,
        dstCi = -1;
      outer: for (let ri = 0; ri < categoryStableIds.length; ri++) {
        for (let ci = 0; ci < categoryStableIds[ri].length; ci++) {
          if (categoryStableIds[ri][ci] === String(over.id)) {
            dstRi = ri;
            dstCi = ci;
            break outer;
          }
        }
      }
      if (dstRi === -1 || (srcRi === dstRi && srcCi === dstCi)) return;

      const newStableIds = categoryStableIds.map((arr) => [...arr]);

      if (srcRi === dstRi) {
        newStableIds[srcRi] = arrayMove(newStableIds[srcRi], srcCi, dstCi);
        pack.rounds[srcRi].categories = arrayMove(
          pack.rounds[srcRi].categories,
          srcCi,
          dstCi
        );
        if (selectedRI === srcRi)
          setSelectedCI((ci) => remapIndex(ci, srcCi, dstCi));
      } else {
        const [stableId] = newStableIds[srcRi].splice(srcCi, 1);
        newStableIds[dstRi].splice(dstCi, 0, stableId);
        const [cat] = pack.rounds[srcRi].categories.splice(srcCi, 1);
        pack.rounds[dstRi].categories.splice(dstCi, 0, cat);

        if (selectedRI === srcRi && selectedCI === srcCi) {
          setSelectedRI(dstRi);
          setSelectedCI(dstCi);
        } else if (selectedRI === srcRi && selectedCI > srcCi) {
          setSelectedCI((ci) => ci - 1);
        } else if (selectedRI === dstRi && selectedCI >= dstCi) {
          setSelectedCI((ci) => ci + 1);
        }
      }

      setCategoryStableIds(newStableIds);
      setPack({ ...pack });
      dragStateRef.current.categoryStableIds = newStableIds;
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  );

  const onDragEnd = (event: DragEndEvent) => {
    setActiveId(null);
    const { active, over } = event;
    if (!over || active.id === over.id) return;

    // Category moves are committed live in onDragOver; only handle round reorder here
    if (!String(active.id).startsWith("round-")) return;

    const oldIndex = pack.rounds.findIndex(
      (_, i) => `round-${i}` === active.id
    );
    const newIndex = pack.rounds.findIndex((_, i) => `round-${i}` === over.id);
    if (oldIndex === -1 || newIndex === -1) return;
    pack.rounds = arrayMove(pack.rounds, oldIndex, newIndex);
    setCategoryStableIds((prev) => arrayMove(prev, oldIndex, newIndex));
    setExpandedRounds((prev) => arrayMove(prev, oldIndex, newIndex));
    setSelectedRI((ri) => remapIndex(ri, oldIndex, newIndex));
    setPack({ ...pack });
  };

  return {
    pack,
    setPack,
    sidebarRef,
    sensors,
    selectedCategory,
    selectedCategoryIndex: selectedCI,
    selectedRoundIndex: selectedRI,
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
    addCategoryFromJson,
    duplicateCategory,
    deleteCategory,
    addFinalRoundCategory,
    changeFinalRoundCategory,
    duplicateFinalRoundCategory,
    deleteFinalRoundCategory,
    categoryStableIds,
    activeId,
    onDragStart,
    onDragOver,
    onDragEnd,
    collisionDetection,
    onDragEndFinalRoundCategories,
  };
}
