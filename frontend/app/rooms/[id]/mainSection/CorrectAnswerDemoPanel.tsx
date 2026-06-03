import { CorrectAnswerDemo } from "@/types/room";
import { FaMusic } from "react-icons/fa6";

export default function CorrectAnswerDemoPanel({
  correctAnswerDemo,
}: {
  correctAnswerDemo: CorrectAnswerDemo;
}) {
  const { answers, comment } = correctAnswerDemo;

  return (
    <div className="h-full flex flex-col items-center justify-center gap-4 p-4 sm:p-6 overflow-hidden">
      <div className="flex flex-col items-center gap-1">
        <p className="text-[10px] font-semibold uppercase tracking-widest text-on-surface-muted">
          Correct answer
        </p>
        <p className="text-3xl font-bold text-on-surface text-center">
          {answers.join(", ")}
        </p>
      </div>

      {comment?.text && (
        <div className="flex flex-col items-center gap-1">
          <p className="text-[10px] font-semibold uppercase tracking-widest text-on-surface-muted">
            Comment
          </p>
          <p className="text-base text-center text-on-surface-muted">
            {comment.text}
          </p>
        </div>
      )}

      {comment?.attachment && (
        <div className="w-full flex-1 min-h-0 flex justify-center items-center">
          {comment.attachment.type === "image" && (
            <img
              src={comment.attachment.url}
              alt="Comment attachment"
              className="w-full h-full object-contain rounded-lg"
            />
          )}
          {comment.attachment.type === "video" && (
            <video
              src={comment.attachment.url}
              className="w-full h-full object-contain rounded-md"
              controls
              autoPlay
            />
          )}
          {comment.attachment.type === "audio" && (
            <div className="flex flex-col items-center gap-2">
              <div className="w-14 h-14 rounded-full bg-surface-raised border border-border flex items-center justify-center text-primary animate-pulse">
                <FaMusic size={22} />
              </div>
              <audio src={comment.attachment.url} controls autoPlay />
            </div>
          )}
        </div>
      )}
    </div>
  );
}
