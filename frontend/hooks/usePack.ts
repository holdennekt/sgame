import { useRef, useState } from "react";
import {
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from "@dnd-kit/core";
import { arrayMove } from "@dnd-kit/sortable";
import { FinalRoundCategoryFormData, PackFormData } from "@/types/pack";

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
    if (selectedRI > ri) setSelectedRI(selectedRI + 1);
    setPack({ ...pack });
  };

  const deleteRound = (ri: number) => {
    pack.rounds = pack.rounds.filter((_, i) => i !== ri);
    setExpandedRounds((prev) => prev.filter((_, i) => i !== ri));
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
    if (selectedRI === ri && selectedCI > ci) setSelectedCI(selectedCI + 1);
    setPack({ ...pack });
  };

  const deleteCategory = (ri: number, ci: number) => {
    pack.rounds[ri].categories = pack.rounds[ri].categories.filter(
      (_, i) => i !== ci
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

  const onDragEndRounds = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const oldIndex = pack.rounds.findIndex((_, i) => String(i) === active.id);
    const newIndex = pack.rounds.findIndex((_, i) => String(i) === over.id);
    pack.rounds = arrayMove(pack.rounds, oldIndex, newIndex);
    setExpandedRounds((prev) => arrayMove(prev, oldIndex, newIndex));
    setSelectedRI((ri) => remapIndex(ri, oldIndex, newIndex));
    setPack({ ...pack });
  };

  const onDragEndCategories = (ri: number, event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const categories = pack.rounds[ri].categories;
    const oldIndex = categories.findIndex((_, i) => String(i) === active.id);
    const newIndex = categories.findIndex((_, i) => String(i) === over.id);
    pack.rounds[ri].categories = arrayMove(categories, oldIndex, newIndex);
    if (selectedRI === ri)
      setSelectedCI((ci) => remapIndex(ci, oldIndex, newIndex));
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
    duplicateCategory,
    deleteCategory,
    addFinalRoundCategory,
    changeFinalRoundCategory,
    duplicateFinalRoundCategory,
    deleteFinalRoundCategory,
    onDragEndRounds,
    onDragEndCategories,
    onDragEndFinalRoundCategories,
  };
}
