import { AttachmentFormData } from "@/types/pack";
import { useRef } from "react";
import { IoIosLink, IoIosClose } from "react-icons/io";
import { FiUploadCloud, FiFile, FiExternalLink, FiPaperclip } from "react-icons/fi";

const labelCls = "block text-xs font-medium text-on-surface-muted";

function AttachmentPreview({ attachment }: { attachment: AttachmentFormData }) {
  if (attachment.type === "existing") {
    return (
      <a
        href={attachment.url}
        target="_blank"
        rel="noopener noreferrer"
        className="flex items-center gap-2.5 px-3 py-2.5 rounded-lg border border-border bg-surface-raised hover:border-primary group transition-colors duration-150"
      >
        <FiFile size={15} className="shrink-0 text-on-surface-muted group-hover:text-primary transition-colors duration-150" />
        <span className="flex-1 min-w-0 text-xs text-on-surface truncate">{attachment.key}</span>
        <FiExternalLink size={12} className="shrink-0 text-on-surface-muted group-hover:text-primary transition-colors duration-150" />
      </a>
    );
  }
  if (attachment.type === "url" && attachment.url) {
    return (
      <a
        href={attachment.url}
        target="_blank"
        rel="noopener noreferrer"
        className="flex items-center gap-2.5 px-3 py-2.5 rounded-lg border border-border bg-surface-raised hover:border-primary group transition-colors duration-150"
      >
        <IoIosLink size={15} className="shrink-0 text-on-surface-muted group-hover:text-primary transition-colors duration-150" />
        <span className="flex-1 min-w-0 text-xs text-on-surface truncate">{attachment.url}</span>
        <FiExternalLink size={12} className="shrink-0 text-on-surface-muted group-hover:text-primary transition-colors duration-150" />
      </a>
    );
  }
  if (attachment.type === "file" && attachment.file) {
    return (
      <div className="flex items-center gap-2.5 px-3 py-2.5 rounded-lg border border-border bg-surface-raised">
        <FiFile size={15} className="shrink-0 text-on-surface-muted" />
        <span className="flex-1 min-w-0 text-xs text-on-surface truncate">{attachment.file.name}</span>
      </div>
    );
  }
  return (
    <div className="flex items-center gap-2.5 px-3 py-2.5 rounded-lg border border-dashed border-border">
      <FiPaperclip size={13} className="shrink-0 text-on-surface-muted opacity-40" />
      <span className="text-xs text-on-surface-muted opacity-40">No attachment</span>
    </div>
  );
}

const tabs = ["existing", "file", "url"] as const;
const tabLabel: Record<typeof tabs[number], string> = { existing: "Existing", file: "File", url: "URL" };

export default function AttachmentEditor({
  attachment,
  saveAttachment,
  readOnly,
}: {
  attachment: AttachmentFormData;
  saveAttachment: (attachment: AttachmentFormData) => void;
  readOnly: boolean;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const dragRef = useRef(false);

  if (readOnly) {
    return (
      <div className="flex flex-col gap-1.5">
        <span className={labelCls}>Attachment</span>
        <AttachmentPreview attachment={attachment} />
      </div>
    );
  }

  const switchTab = (t: typeof tabs[number]) => {
    if (t === "existing") {
      if (attachment.type !== "existing") return;
      saveAttachment({ type: "existing", key: attachment.key, url: attachment.url });
    } else {
      saveAttachment({ type: t });
    }
  };

  return (
    <div className="flex flex-col gap-1.5">
      <span className={labelCls}>Attachment</span>

      {/* Tab switcher */}
      <div className="flex rounded-lg border border-border overflow-hidden">
        {tabs.map(t => {
          const active = attachment.type === t;
          const disabled = t === "existing" && attachment.type !== "existing";
          return (
            <button
              key={t}
              type="button"
              disabled={disabled}
              onClick={() => switchTab(t)}
              className={`flex-1 py-1.5 text-xs font-medium transition-colors duration-150
                ${active
                  ? "bg-primary text-on-primary"
                  : disabled
                    ? "bg-surface text-on-surface-muted opacity-40 cursor-not-allowed"
                    : "bg-surface text-on-surface-muted hover:bg-surface-raised hover:text-on-surface"
                }
              `}
            >
              {tabLabel[t]}
            </button>
          );
        })}
      </div>

      {/* Tab content */}
      {attachment.type === "existing" && (
        <AttachmentPreview attachment={attachment} />
      )}

      {attachment.type === "file" && (
        <div>
          {attachment.file ? (
            <div className="flex items-center gap-2.5 px-3 py-2.5 rounded-lg border border-border bg-surface-raised">
              <FiFile size={15} className="shrink-0 text-primary" />
              <span className="flex-1 min-w-0 text-xs text-on-surface truncate">{attachment.file.name}</span>
              <button
                type="button"
                onClick={() => saveAttachment({ type: "file" })}
                className="shrink-0 text-on-surface-muted hover:text-on-surface transition-colors duration-150"
              >
                <IoIosClose size={16} />
              </button>
            </div>
          ) : (
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              onDragOver={e => { e.preventDefault(); dragRef.current = true; }}
              onDragLeave={() => { dragRef.current = false; }}
              onDrop={e => {
                e.preventDefault();
                const file = e.dataTransfer.files?.[0];
                if (file) saveAttachment({ type: "file", file });
              }}
              className="w-full flex flex-col items-center gap-2 px-3 py-5 rounded-lg border-2 border-dashed border-border hover:border-primary hover:bg-surface-raised text-on-surface-muted hover:text-on-surface transition-colors duration-150"
            >
              <FiUploadCloud size={20} />
              <span className="text-xs text-center leading-relaxed">
                Click or drag to upload<br />
                <span className="text-[10px] opacity-60">image, audio, video</span>
              </span>
            </button>
          )}
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*, audio/*, video/*"
            className="hidden"
            onChange={e => {
              const file = e.target.files?.[0];
              if (file) saveAttachment({ type: "file", file });
            }}
          />
        </div>
      )}

      {attachment.type === "url" && (
        <div className="flex items-center rounded-lg border border-border bg-background overflow-hidden focus-within:ring-2 focus-within:ring-primary/40 focus-within:border-primary transition-[border-color] duration-150">
          <div className="pl-3 pr-1 text-on-surface-muted shrink-0">
            <IoIosLink size={15} />
          </div>
          <input
            className="flex-1 min-w-0 h-9 pr-2.5 bg-transparent text-sm text-on-background outline-none placeholder:text-on-surface-muted"
            type="url"
            placeholder="https://example.com/video.mp4"
            value={attachment.url ?? ""}
            onChange={e => saveAttachment({ type: "url", url: e.target.value })}
          />
        </div>
      )}
    </div>
  );
}
