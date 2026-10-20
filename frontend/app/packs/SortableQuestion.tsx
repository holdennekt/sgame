import React from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { AttachmentFormData, QuestionFormData } from "@/types/pack";
import { FaTrashCan } from "react-icons/fa6";
import { RiDraggable } from "react-icons/ri";
import {
  FiCopy,
  FiImage,
  FiMic,
  FiVideo,
  FiPaperclip,
  FiAlignLeft,
  FiAlertCircle,
} from "react-icons/fi";

const iconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150";
const dangerIconBtnCls =
  "h-6 w-6 inline-flex items-center justify-center rounded-md text-danger hover:bg-surface-raised transition-colors duration-150";

const typeLabel: Record<string, string> = {
  regular: "Regular",
  catInBag: "Cat in bag",
  auction: "Auction",
};

const attLabel: Record<string, string> = {
  image: "Image",
  audio: "Audio",
  video: "Video",
  other: "File",
  processing: "Processing…",
  failed: "Failed",
};

const imageExts = new Set([
  "jpg",
  "jpeg",
  "png",
  "gif",
  "webp",
  "svg",
  "bmp",
  "tiff",
]);
const audioExts = new Set(["mp3", "wav", "ogg", "flac", "aac", "m4a"]);
const videoExts = new Set(["mp4", "webm", "mkv", "mov", "avi", "wmv"]);

function extType(path: string): "image" | "audio" | "video" | "other" {
  const ext = path.split(".").pop()?.toLowerCase() ?? "";
  if (imageExts.has(ext)) return "image";
  if (audioExts.has(ext)) return "audio";
  if (videoExts.has(ext)) return "video";
  return "other";
}

type AttachmentTypeLabel =
  | "image"
  | "audio"
  | "video"
  | "other"
  | "processing"
  | "failed";

function attachmentType(att: AttachmentFormData): AttachmentTypeLabel | null {
  if (att.kind === "existing") {
    if (att.status === "pending") return "processing";
    if (att.status === "failed") return "failed";
    return att.type;
  }
  if (att.kind === "file") {
    if (!att.file) return null;
    const mime = att.file.type;
    if (mime.startsWith("image/")) return "image";
    if (mime.startsWith("audio/")) return "audio";
    if (mime.startsWith("video/")) return "video";
    return "other";
  }
  if (att.kind === "url") {
    if (!att.url) return null;
    return extType(att.url);
  }
  return null;
}

function hasAtt(att: AttachmentFormData) {
  return (
    att.kind === "existing" ||
    (att.kind === "file" && att.file != null) ||
    (att.kind === "url" && att.url != null)
  );
}

function AttachmentIcon({
  type,
  size = 12,
}: {
  type: AttachmentTypeLabel;
  size?: number;
}) {
  if (type === "processing")
    return (
      <span
        className="shrink-0 border-2 border-primary border-t-transparent rounded-full animate-spin"
        style={{ width: size, height: size }}
      />
    );
  if (type === "failed")
    return <FiAlertCircle size={size} className="shrink-0 text-danger" />;
  const cls = "shrink-0 text-on-surface-muted";
  if (type === "image") return <FiImage size={size} className={cls} />;
  if (type === "audio") return <FiMic size={size} className={cls} />;
  if (type === "video") return <FiVideo size={size} className={cls} />;
  return <FiPaperclip size={size} className={cls} />;
}

function AttachmentLabel({ type }: { type: AttachmentTypeLabel }) {
  return (
    <span
      className={`inline-flex items-center gap-1 shrink-0 ${
        type === "failed" ? "text-danger" : "text-on-surface-muted"
      }`}
    >
      <AttachmentIcon type={type} size={12} />
      <span>{attLabel[type]}</span>
    </span>
  );
}

export default function SortableQuestion({
  id,
  question,
  readOnly,
  onOpen,
  onDuplicate,
  onDelete,
}: {
  id: string;
  question: QuestionFormData;
  readOnly: boolean;
  onOpen: () => void;
  onDuplicate: () => void;
  onDelete: () => void;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id, disabled: readOnly });

  const attType = hasAtt(question.attachment)
    ? attachmentType(question.attachment)
    : null;
  const hasCommentText = !!question.comment.text;
  const commentAttType = hasAtt(question.comment.attachment)
    ? attachmentType(question.comment.attachment)
    : null;

  const metaItems: React.ReactNode[] = [];
  if (question.answers.length > 0)
    metaItems.push(
      `${question.answers.length} ${
        question.answers.length === 1 ? "answer" : "answers"
      }`
    );
  if (hasCommentText)
    metaItems.push(
      <span className="inline-flex items-center gap-1">
        <span>comment</span>
        <FiAlignLeft size={9} className="text-on-surface-muted" />
      </span>
    );
  if (commentAttType)
    metaItems.push(
      <span className="inline-flex items-center gap-1">
        <span>comment</span>
        <AttachmentIcon type={commentAttType} size={9} />
      </span>
    );

  return (
    <div
      ref={setNodeRef}
      style={{
        transform: CSS.Transform.toString(
          transform ? { ...transform, x: 0 } : null
        ),
        transition: transform ? transition : undefined,
      }}
      className={`group relative flex items-center gap-2 px-4 py-3 rounded-md border border-border hover:bg-surface-raised hover:border-primary transition-colors duration-150 select-none ${
        isDragging ? "opacity-50 z-50" : ""
      }`}
    >
      {!readOnly && (
        <button
          type="button"
          className={`${iconBtnCls} can-hover:opacity-0 group-hover:opacity-100 shrink-0 cursor-grab active:cursor-grabbing touch-none`}
          {...attributes}
          {...listeners}
          onClick={(e) => e.stopPropagation()}
        >
          <RiDraggable size={14} />
        </button>
      )}
      <div
        className="flex items-center gap-4 flex-1 min-w-0 cursor-pointer"
        onClick={onOpen}
      >
        <span
          className={`font-black text-primary w-12 shrink-0 leading-none tabular-nums text-center ${
            question.value >= 1000 ? "text-lg" : "text-2xl"
          }`}
        >
          {question.value}
        </span>
        <div className="flex-1 min-w-0 flex flex-col gap-1">
          <div className="flex items-center gap-2 min-w-0">
            {question.text ? (
              <p className="text-sm text-on-surface truncate min-w-0">
                {question.text}
              </p>
            ) : !attType ? (
              <em className="text-sm text-on-surface-muted opacity-40 flex-1">
                No content
              </em>
            ) : null}
            {attType && <AttachmentLabel type={attType} />}
          </div>

          <div className="flex flex-wrap items-center gap-1.5 text-[10px] text-on-surface-muted">
            <span className="px-1.5 py-0.5 rounded-full bg-surface-raised font-medium border border-border">
              {typeLabel[question.type] ?? question.type}
            </span>
            {metaItems.map((item, i) => (
              <React.Fragment key={i}>
                <span className="opacity-40">·</span>
                {typeof item === "string" ? <span>{item}</span> : item}
              </React.Fragment>
            ))}
          </div>
        </div>
      </div>
      {!readOnly && (
        <div className="flex items-center gap-0.5 can-hover:opacity-0 group-hover:opacity-100 shrink-0">
          <button
            type="button"
            className={iconBtnCls}
            title="Duplicate question"
            onClick={(e) => {
              e.stopPropagation();
              onDuplicate();
            }}
          >
            <FiCopy size={12} />
          </button>
          <button
            type="button"
            className={dangerIconBtnCls}
            onClick={(e) => {
              e.stopPropagation();
              onDelete();
            }}
          >
            <FaTrashCan size={12} />
          </button>
        </div>
      )}
    </div>
  );
}
