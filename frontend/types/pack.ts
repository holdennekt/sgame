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
  | {
      kind: "existing";
      key: string;
      url: string;
      type: FileType;
      status?: AttachmentStatus;
      error?: string;
    }
  | { kind: "file"; file?: File }
  | { kind: "url"; url?: string };

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
  public: boolean;
}

export interface SignURLResponse {
  url: string;
  formData: Record<string, string>;
  // must be registered on the draft before the presigned upload is performed
  key: string;
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

// missing status means the attachment predates the raw/compressed split
export type AttachmentStatus = "pending" | "ready" | "failed";

export interface Attachment {
  key: string;
  url: string;
  type: FileType;
  mimeType: string;
  size: number;
  duration: number;
  status?: AttachmentStatus;
  error?: string;
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
  if (!attachment) return { kind: "file" };
  return {
    kind: "existing",
    key: attachment.key,
    url: attachment.url,
    type: attachment.type,
    status: attachment.status,
    error: attachment.error,
  };
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

export function mapPackAttachments(
  pack: PackFormData,
  fn: (a: AttachmentFormData) => AttachmentFormData
): PackFormData {
  const mapQuestion = (q: QuestionFormData): QuestionFormData => ({
    ...q,
    attachment: fn(q.attachment),
    comment: { ...q.comment, attachment: fn(q.comment.attachment) },
  });
  return {
    ...pack,
    rounds: pack.rounds.map((r) => ({
      ...r,
      categories: r.categories.map((c) => ({
        ...c,
        questions: c.questions.map(mapQuestion),
      })),
    })),
    finalRound: {
      categories: pack.finalRound.categories.map((c) => ({
        ...c,
        question: {
          ...c.question,
          attachment: fn(c.question.attachment),
          comment: {
            ...c.question.comment,
            attachment: fn(c.question.comment.attachment),
          },
        },
      })),
    },
  };
}

function* questionAttachments(q: {
  attachment: AttachmentFormData;
  comment: { attachment: AttachmentFormData };
}): Generator<AttachmentFormData> {
  yield q.attachment;
  yield q.comment.attachment;
}

export function* packAttachments(
  pack: PackFormData
): Generator<AttachmentFormData> {
  for (const round of pack.rounds) {
    for (const category of round.categories) {
      for (const question of category.questions) {
        yield* questionAttachments(question);
      }
    }
  }
  for (const category of pack.finalRound.categories) {
    yield* questionAttachments(category.question);
  }
}

function isPending(a: AttachmentFormData): boolean {
  return a.kind === "existing" && a.status === "pending";
}

export function packHasPendingAttachments(pack: PackFormData): boolean {
  for (const a of packAttachments(pack)) {
    if (isPending(a)) return true;
  }
  return false;
}

export function categoryHasPendingAttachments(
  category: CategoryFormData
): boolean {
  for (const q of category.questions) {
    for (const a of questionAttachments(q)) {
      if (isPending(a)) return true;
    }
  }
  return false;
}

export function roundHasPendingAttachments(round: RoundFormData): boolean {
  return round.categories.some(categoryHasPendingAttachments);
}

export function finalRoundCategoryHasPendingAttachments(
  category: FinalRoundCategoryFormData
): boolean {
  for (const a of questionAttachments(category.question)) {
    if (isPending(a)) return true;
  }
  return false;
}

type SignURLFn = (
  params: SignURLRequest
) => Promise<SignURLResponse | { error: string }>;

// deferred until the draft save that registers its key has succeeded
export interface PendingUpload {
  key: string;
  upload: () => Promise<void>;
}

export interface ConvertedPackRequest {
  request: CreatePackRequest;
  uploads: PendingUpload[];
}

export async function convertPackFormDataToRequest(
  formData: PackFormData,
  signURL: SignURLFn,
  fileCache?: Map<File, string>,
  urlCache?: Map<string, string>
): Promise<ConvertedPackRequest> {
  const isPublic = false;
  const uploads: PendingUpload[] = [];

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
                uploads,
                fileCache,
                urlCache
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
                  uploads,
                  fileCache,
                  urlCache
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
          uploads,
          fileCache,
          urlCache
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
              uploads,
              fileCache,
              urlCache
            ),
          },
        };
      })
    ),
  };

  return {
    request: {
      name: formData.name,
      type: formData.type,
      rounds,
      finalRound,
    },
    uploads,
  };
}

async function convertAttachment(
  attachment: AttachmentFormData,
  isPublic: boolean,
  signURL: SignURLFn,
  uploads: PendingUpload[],
  fileCache?: Map<File, string>,
  urlCache?: Map<string, string>
): Promise<CreateAttachmentRequest | null> {
  switch (attachment.kind) {
    case "file": {
      const file = attachment.file;
      if (!file) return null;

      if (fileCache?.has(file)) {
        return { key: fileCache.get(file)! };
      }

      const signResult = await signURL({ public: isPublic });
      if ("error" in signResult) throw new Error(signResult.error);
      const { url, formData: reqFormData, key } = signResult;

      uploads.push({
        key,
        upload: async () => {
          const formData = new FormData();
          Object.entries(reqFormData).forEach(([k, v]) => {
            formData.append(k, v);
          });
          formData.append("file", file);

          const response = await fetch(url, {
            method: "POST",
            body: formData,
          });

          if (!response.ok) {
            fileCache?.delete(file);
            throw new Error(`Failed to upload ${file.name}`);
          }
        },
      });

      fileCache?.set(file, key);
      return { key };
    }

    case "url":
      if (!attachment.url) return null;
      if (urlCache?.has(attachment.url)) {
        return { key: urlCache.get(attachment.url)! };
      }
      return { url: attachment.url };

    case "existing":
      return { key: attachment.key };
  }
}
