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
  expanded: boolean;
  categories: CategoryFormData[];
}

export interface CategoryFormData {
  name: string;
  selected: boolean;
  questions: QuestionFormData[];
}

export interface CommentFormData {
  text: string;
  attachment: AttachmentFormData;
}

export interface QuestionFormData {
  index: number;
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
  expanded: boolean;
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
  questions: CreateQuestionRequest[];
}

export interface CreateCommentRequest {
  text: string | null;
  attachment: CreateAttachmentRequest | null;
}

export interface CreateQuestionRequest {
  index: number;
  value: number;
  type: QuestionType;
  text: string;
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
  text: string;
  attachment: CreateAttachmentRequest | null;
  answers: string[];
  comment: CreateCommentRequest | null;
}

export interface CreatePackResponse {
  id: string;
}

export interface UpdatePackRequest extends CreatePackRequest {
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
  text: string;
  answers: string[];
  comment: Comment | null;
}

const dummyQuestion: Question = {
  index: 0,
  value: 0,
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

export interface HiddenRound {
  name: string;
  categories: HiddenCategory[];
}

export interface HiddenCategory {
  name: string;
}

export interface HiddenQuestion {
  index: number;
  value: number;
  attachment: Attachment | null;
}

export interface HiddenFinalRound {
  categories: HiddenFinalRoundCategory[];
}

export interface HiddenFinalRoundCategory {
  name: string;
}

export interface HiddenFinalRoundQuestion {
  text: string;
  attachment: Attachment | null;
}

export function convertPackToFormData(
  pack: Omit<Pack, "id" | "createdBy">,
): PackFormData {
  const rounds = pack.rounds.map(convertRoundToFormData);
  if (rounds[0].categories.length) rounds[0].categories[0].selected = true;
  return {
    name: pack.name,
    type: pack.type,
    rounds,
    finalRound: convertFinalRoundToFormData(pack.finalRound),
  };
}

function convertRoundToFormData(round: Round): RoundFormData {
  return {
    name: round.name,
    expanded: true,
    categories: round.categories.map(convertCategoryToFormData),
  };
}

function convertCategoryToFormData(category: Category): CategoryFormData {
  return {
    name: category.name,
    selected: false,
    questions: category.questions.map(convertQuestionToFormData),
  };
}

function convertQuestionToFormData(question: Question): QuestionFormData {
  return {
    index: question.index,
    value: question.value,
    type: question.type,
    text: question.text,
    attachment: convertAttachmentToFormData(question.attachment),
    answers: question.answers,
    comment: convertCommentToFormData(question.comment),
  };
}

function convertAttachmentToFormData(
  attachment: Attachment | null,
): AttachmentFormData {
  if (!attachment) return { type: "file" };
  return { type: "existing", key: attachment.key, url: attachment.url };
}

function convertCommentToFormData(comment: Comment | null): CommentFormData {
  return {
    text: comment?.text ?? "",
    attachment: convertAttachmentToFormData(
      comment ? comment.attachment : null,
    ),
  };
}

function convertFinalRoundToFormData(
  finalRound: FinalRound,
): FinalRoundFormData {
  return {
    expanded: true,
    categories: finalRound.categories.map(convertFinalRoundCategoryToFormData),
  };
}

function convertFinalRoundCategoryToFormData(
  category: FinalRoundCategory,
): FinalRoundCategoryFormData {
  return {
    name: category.name,
    question: {
      ...category.question,
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
): Promise<CreatePackRequest> {
  const isPublic = formData.type === "public";

  const rounds = await Promise.all(
    formData.rounds.map(async (round) => ({
      name: round.name,
      categories: await Promise.all(
        round.categories.map(async (category) => ({
          name: category.name,
          questions: await Promise.all(
            category.questions.map(async (question) => {
              const commentText = question.comment.text || null;
              const commentAttachment = await convertAttachment(
                question.comment.attachment,
                isPublic,
                signURL,
              );
              return {
                ...question,
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
                ),
              };
            }),
          ),
        })),
      ),
    })),
  );

  const finalRound = {
    categories: await Promise.all(
      formData.finalRound.categories.map(async ({ name, question }) => {
        const commentText = question.comment.text || null;
        const commentAttachment = await convertAttachment(
          question.comment.attachment,
          isPublic,
          signURL,
        );
        return {
          name,
          question: {
            ...question,
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
            ),
          },
        };
      }),
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
): Promise<CreateAttachmentRequest | null> {
  switch (attachment.type) {
    case "file": {
      if (!attachment.file) return null;

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

      return { key: reqFormData.key };
    }

    case "url":
      if (!attachment.url) return null;
      return { url: attachment.url };

    case "existing":
      return { key: attachment.key };
  }
}
