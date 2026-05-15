import { AttachmentFormData } from "@/types/pack";

export default function AttachmentEditor({
  attachment,
  editAttachment,
  setEditAttachment,
  readOnly,
}: {
  attachment: AttachmentFormData;
  editAttachment: AttachmentFormData;
  setEditAttachment: React.Dispatch<React.SetStateAction<AttachmentFormData>>;
  readOnly: boolean;
}) {
  const getAttachmentSection = () => {
    switch (editAttachment.type) {
    case "existing":
      return (
        <p className="text-sm font-medium">
            Link:{" "}
          <a
            href={editAttachment.url}
            target="_blank"
            rel="noopener noreferrer"
          >
            {editAttachment.key}
          </a>
        </p>
      );
    case "file":
      return (
        <label className="block">
          <p className="text-sm font-medium">Select File</p>
          <input
            className="w-full h-8 mt-1 text-sm text-gray-400 file:mr-4 file:py-1 file:px-4 file:rounded-md file:bg-blue-600 file:text-white file:border-0 hover:file:bg-blue-700 cursor-pointer"
            type="file"
            accept="image/*, audio/*, video/*"
            disabled={readOnly}
            onChange={e => {
              const file = e.target.files?.[0];
              if (file) {
                setEditAttachment({ type: "file", file });
              }
            }}
          />
          {editAttachment.file && (
            <p className="text-xs text-green-500 mt-1">File selected</p>
          )}
        </label>
      );
    case "url":
      return (
        <label className="block">
          <p className="text-sm font-medium">Content URL</p>
          <input
            className="w-full h-8 rounded-md mt-1 p-1 text-black"
            type="url"
            placeholder="https://example.com/video.mp4"
            value={editAttachment.url}
            onChange={e =>
              setEditAttachment({ type: "url", url: e.target.value })
            }
            required
            readOnly={readOnly}
          />
        </label>
      );
    default:
      return <></>;
    }
  };

  return (
    <div>
      <p className="text-sm font-medium">Attachment</p>
      {!readOnly && (
        <div className="flex gap-4 mb-2">
          <label className="flex items-center gap-2 cursor-pointer text-sm">
            <input
              type="radio"
              checked={editAttachment.type === "existing"}
              onChange={
                attachment.type === "existing" ?
                  () =>
                    setEditAttachment({
                      type: "existing",
                      key: attachment.key,
                      url: attachment.url,
                    }) :
                  undefined
              }
              disabled={attachment.type !== "existing" || readOnly}
            />
            Existing
          </label>
          <label className="flex items-center gap-2 cursor-pointer text-sm">
            <input
              type="radio"
              checked={editAttachment.type === "file"}
              onChange={() => setEditAttachment({ type: "file" })}
              disabled={readOnly}
            />
            File
          </label>
          <label className="flex items-center gap-2 cursor-pointer text-sm">
            <input
              type="radio"
              checked={editAttachment.type === "url"}
              onChange={() => setEditAttachment({ type: "url" })}
              disabled={readOnly}
            />
            URL
          </label>
        </div>
      )}
      {getAttachmentSection()}
    </div>
  );
}
