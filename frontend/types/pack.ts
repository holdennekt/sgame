import { User } from "./user";

export type PrivacyType = "public" | "private";
export type QuestionType = "regular" | "catInBag" | "auction";

export interface PackFormData {
  name: string;
  type: PrivacyType;
  rounds: RoundFormData[];
  finalRound: FinalRoundFormData;
}

export interface RoundFormData {
  name: string;
  categories: CategoryFormData[];
}

export interface CategoryFormData {
  name: string;
  comment: string;
  questions: QuestionFormData[];
}

export interface CommentFormData {
  text: string;
  attachment: AttachmentFormData;
}

export interface QuestionFormData {
  value: number;
  type: QuestionType;
  text: string;
  attachment: AttachmentFormData;
  answers: string[];
  comment: CommentFormData;
}

export type AttachmentFormData =
  | { type: "existing"; key: string; url: string }
  | { type: "file"; file?: File }
  | { type: "url"; url?: string };

export interface FinalRoundFormData {
  categories: FinalRoundCategoryFormData[];
}

export interface FinalRoundCategoryFormData {
  name: string;
  question: FinalRoundQuestionFormData;
}

export interface FinalRoundQuestionFormData {
  text: string;
  attachment: AttachmentFormData;
  answers: string[];
  comment: CommentFormData;
}

export interface CreatePackRequest {
  name: string;
  type: PrivacyType;
  rounds: CreateRoundRequest[];
  finalRound: CreateFinalRoundRequest;
}

export interface CreateRoundRequest {
  name: string;
  categories: CreateCategoryRequest[];
}

export interface CreateCategoryRequest {
  name: string;
  comment: string | null;
  questions: CreateQuestionRequest[];
}

export interface CreateCommentRequest {
  text: string | null;
  attachment: CreateAttachmentRequest | null;
}

export interface CreateQuestionRequest {
  value: number;
  type: QuestionType;
  text: string | null;
  attachment: CreateAttachmentRequest | null;
  answers: string[];
  comment: CreateCommentRequest | null;
}

export interface CreateAttachmentRequest {
  key?: string;
  url?: string;
}

export interface CreateFinalRoundRequest {
  categories: CreateFinalRoundCategoryRequest[];
}

export interface CreateFinalRoundCategoryRequest {
  name: string;
  question: CreateFinalRoundQuestionRequest;
}

export interface CreateFinalRoundQuestionRequest {
  text: string | null;
  attachment: CreateAttachmentRequest | null;
  answers: string[];
  comment: CreateCommentRequest | null;
}

export interface CreatePackResponse {
  id: string;
}

export interface SignURLRequest {
  filename: string;
  public: boolean;
}

export interface SignURLResponse {
  url: string;
  formData: Record<string, string>;
  getUrl?: string;
}

export interface Pack {
  id: string;
  createdBy: User;
  name: string;
  type: PrivacyType;
  rounds: Round[];
  finalRound: FinalRound;
  createdAt: string;
  updatedAt: string;
}

export interface PackPreview {
  id: string;
  name: string;
}

export interface Round {
  name: string;
  categories: Category[];
}

export interface Category {
  name: string;
  comment: string | null;
  questions: Question[];
}

export type FileType = "image" | "audio" | "video";

export interface Attachment {
  key: string;
  url: string;
  type: FileType;
  mimeType: string;
  size: number;
  duration: number;
}

const dummyAttachment: Attachment = {
  key: "",
  url: "",
  type: "image",
  mimeType: "",
  size: 0,
  duration: 0,
};

export const isAttachment = (obj: unknown): obj is Attachment => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyAttachment).every((key) => Object.hasOwn(obj, key));
};

export interface Comment {
  text: string | null;
  attachment: Attachment | null;
}

const dummyComment: Comment = {
  text: null,
  attachment: null,
};

export const isComment = (obj: unknown): obj is Comment => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyComment).every((key) => Object.hasOwn(obj, key));
};

export interface Question extends HiddenQuestion {
  type: QuestionType;
  text: string | null;
  attachment: Attachment | null;
  answers: string[];
  comment: Comment | null;
}

const dummyQuestion: Question = {
  value: 0,
  category: "",
  text: "",
  attachment: null,
  type: "regular",
  answers: [],
  comment: null,
};

export const isQuestion = (obj: unknown): obj is Question => {
  if (typeof obj !== "object" || obj === null) return false;
  if (
    (obj as Question).attachment &&
    !isAttachment((obj as Question).attachment)
  )
    return false;
  return Object.keys(dummyQuestion).every((key) => Object.hasOwn(obj, key));
};

export interface FinalRound {
  categories: FinalRoundCategory[];
}

export interface FinalRoundCategory extends HiddenFinalRoundCategory {
  question: FinalRoundQuestion;
}

export interface FinalRoundQuestion extends HiddenFinalRoundQuestion {
  answers: string[];
  comment: Comment | null;
}

export interface HiddenPack {
  id: string;
  createdBy: User;
  name: string;
  type: PrivacyType;
  rounds: HiddenRound[];
  finalRound: HiddenFinalRound;
}

export const isHiddenPack = (pack: Pack | HiddenPack): pack is HiddenPack =>
  !("createdAt" in pack);

export interface HiddenRound {
  name: string;
  categories: HiddenCategory[];
}

export interface HiddenCategory {
  name: string;
}

export interface HiddenQuestion {
  category: string;
  value: number;
}

export interface HiddenFinalRound {
  categories: HiddenFinalRoundCategory[];
}

export interface HiddenFinalRoundCategory {
  name: string;
}

export interface HiddenFinalRoundQuestion {
  category: string;
  text: string | null;
  attachment: Attachment | null;
}

export function convertPackToFormData(
  pack: Omit<Pack, "id" | "createdBy">
): PackFormData {
  return {
    name: pack.name,
    type: pack.type,
    rounds: pack.rounds.map(convertRoundToFormData),
    finalRound: convertFinalRoundToFormData(pack.finalRound),
  };
}

function convertRoundToFormData(round: Round): RoundFormData {
  return {
    name: round.name,
    categories: round.categories.map(convertCategoryToFormData),
  };
}

function convertCategoryToFormData(category: Category): CategoryFormData {
  return {
    name: category.name,
    comment: category.comment ?? "",
    questions: category.questions.map(convertQuestionToFormData),
  };
}

function convertQuestionToFormData(question: Question): QuestionFormData {
  return {
    value: question.value,
    type: question.type,
    text: question.text ?? "",
    attachment: convertAttachmentToFormData(question.attachment),
    answers: question.answers,
    comment: convertCommentToFormData(question.comment),
  };
}

function convertAttachmentToFormData(
  attachment: Attachment | null
): AttachmentFormData {
  if (!attachment) return { type: "file" };
  return { type: "existing", key: attachment.key, url: attachment.url };
}

function convertCommentToFormData(comment: Comment | null): CommentFormData {
  return {
    text: comment?.text ?? "",
    attachment: convertAttachmentToFormData(
      comment ? comment.attachment : null
    ),
  };
}

function convertFinalRoundToFormData(
  finalRound: FinalRound
): FinalRoundFormData {
  return {
    categories: finalRound.categories.map(convertFinalRoundCategoryToFormData),
  };
}

function convertFinalRoundCategoryToFormData(
  category: FinalRoundCategory
): FinalRoundCategoryFormData {
  return {
    name: category.name,
    question: {
      ...category.question,
      text: category.question.text ?? "",
      attachment: convertAttachmentToFormData(category.question.attachment),
      comment: convertCommentToFormData(category.question.comment),
    },
  };
}

export async function convertPackFormDataToRequest(
  formData: PackFormData,
  signURL: (params: {
    filename: string;
    public: boolean;
  }) => Promise<
    | { url: string; formData: Record<string, string>; getUrl?: string }
    | { error: string }
  >,
  fileCache?: Map<File, string>
): Promise<CreatePackRequest> {
  const isPublic = false;

  const rounds = await Promise.all(
    formData.rounds.map(async (round) => ({
      name: round.name,
      categories: await Promise.all(
        round.categories.map(async (category) => ({
          name: category.name,
          comment: category.comment || null,
          questions: await Promise.all(
            category.questions.map(async (question) => {
              const commentText = question.comment.text || null;
              const commentAttachment = await convertAttachment(
                question.comment.attachment,
                isPublic,
                signURL,
                fileCache
              );
              return {
                ...question,
                text: question.text || null,
                comment:
                  commentText || commentAttachment
                    ? {
                        text: commentText,
                        attachment: commentAttachment,
                      }
                    : null,
                attachment: await convertAttachment(
                  question.attachment,
                  isPublic,
                  signURL,
                  fileCache
                ),
              };
            })
          ),
        }))
      ),
    }))
  );

  const finalRound = {
    categories: await Promise.all(
      formData.finalRound.categories.map(async ({ name, question }) => {
        const commentText = question.comment.text || null;
        const commentAttachment = await convertAttachment(
          question.comment.attachment,
          isPublic,
          signURL,
          fileCache
        );
        return {
          name,
          question: {
            ...question,
            text: question.text || null,
            comment:
              commentText || commentAttachment
                ? {
                    text: commentText,
                    attachment: commentAttachment,
                  }
                : null,
            attachment: await convertAttachment(
              question.attachment,
              isPublic,
              signURL,
              fileCache
            ),
          },
        };
      })
    ),
  };

  return {
    name: formData.name,
    type: formData.type,
    rounds,
    finalRound,
  };
}

async function convertAttachment(
  attachment: AttachmentFormData,
  isPublic: boolean,
  signURL: (params: {
    filename: string;
    public: boolean;
  }) => Promise<
    | { url: string; formData: Record<string, string>; getUrl?: string }
    | { error: string }
  >,
  fileCache?: Map<File, string>
): Promise<CreateAttachmentRequest | null> {
  switch (attachment.type) {
    case "file": {
      if (!attachment.file) return null;

      if (fileCache?.has(attachment.file)) {
        return { key: fileCache.get(attachment.file)! };
      }

      const signResult = await signURL({
        filename: attachment.file.name,
        public: isPublic,
      });
      if ("error" in signResult) throw new Error(signResult.error);
      const { url, formData: reqFormData } = signResult;

      const formData = new FormData();
      Object.entries(reqFormData).forEach(([key, value]) => {
        formData.append(key, value);
      });
      formData.append("file", attachment.file);

      const response = await fetch(url, {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        throw new Error(`Failed to upload ${attachment.file.name} to S3`);
      }

      fileCache?.set(attachment.file, reqFormData.key);
      return { key: reqFormData.key };
    }

    case "url":
      if (!attachment.url) return null;
      return { url: attachment.url };

    case "existing":
      return { key: attachment.key };
  }
}
