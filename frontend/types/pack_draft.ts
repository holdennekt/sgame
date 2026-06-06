import {
  Attachment,
  Comment,
  PrivacyType,
  QuestionType,
  PackFormData,
  convertPackToFormData,
} from "./pack";
import { User } from "./user";

export interface PackDraft {
  linkedPackId: string | null;
  id: string;
  createdBy: User;
  name: string;
  type: PrivacyType;
  rounds: DraftRound[];
  finalRound: DraftFinalRound;
  createdAt: string;
  updatedAt: string;
}

export interface DraftRound {
  name: string;
  categories: DraftCategory[];
}

export interface DraftCategory {
  name: string;
  comment: string | null;
  questions: DraftQuestion[];
}

export interface DraftQuestion {
  round: string;
  category: string;
  index: number;
  value: number;
  type: QuestionType;
  text: string | null;
  attachment: Attachment | null;
  answers: string[];
  comment: Comment | null;
}

export interface DraftFinalRound {
  categories: DraftFinalRoundCategory[];
}

export interface DraftFinalRoundCategory {
  name: string;
  question: DraftFinalRoundQuestion;
}

export interface DraftFinalRoundQuestion {
  category: string;
  text: string | null;
  attachment: Attachment | null;
  answers: string[];
  comment: Comment | null;
}

export function countIncompleteQuestions(draft: PackDraft): number {
  let count = 0;
  for (const round of draft.rounds) {
    for (const cat of round.categories) {
      count += cat.questions.filter((q) => q.answers.length === 0).length;
    }
  }
  for (const cat of draft.finalRound.categories) {
    if (cat.question.answers.length === 0) count++;
  }
  return count;
}

// Convert a PackDraft to the PackFormData shape used by PackEditor.
// The draft structure mirrors Pack closely — the main difference is that
// answers can be empty, which the editor handles fine.
export function convertDraftToFormData(draft: PackDraft): PackFormData {
  return convertPackToFormData({
    name: draft.name,
    type: draft.type,
    rounds: draft.rounds.map((r) => ({
      name: r.name,
      categories: r.categories.map((c) => ({
        name: c.name,
        comment: c.comment ?? "",
        questions: c.questions.map((q) => ({
          index: q.index,
          value: q.value,
          category: q.category,
          type: q.type,
          text: q.text,
          attachment: q.attachment,
          answers: q.answers,
          comment: q.comment,
        })),
      })),
    })),
    finalRound: {
      categories: draft.finalRound.categories.map((c) => ({
        name: c.name,
        question: {
          category: c.name,
          text: c.question.text,
          attachment: c.question.attachment,
          answers: c.question.answers,
          comment: c.question.comment,
        },
      })),
    },
    createdAt: draft.createdAt,
    updatedAt: draft.updatedAt,
  });
}
