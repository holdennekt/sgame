import { Attachment, QuestionType } from "@/types/pack";
import { useEffect, useRef } from "react";
import { FaMusic } from "react-icons/fa6";
import TimerBar from "./TimerBar";

const textSize = (len: number) =>
  len < 80
    ? "text-3xl"
    : len < 180
    ? "text-xl"
    : len < 400
    ? "text-base"
    : len < 700
    ? "text-sm"
    : "text-xs";

export default function QuestionPanel({
  attachment,
  attachmentLastProgress,
  text,
  textLastProgress,
  questionType,
  timeBar,
}: {
  attachment: Attachment | null;
  attachmentLastProgress: number;
  text: string | null;
  textLastProgress: number;
  questionType: QuestionType | "final";
  timeBar: {
    progress: number;
    endsAt: number;
    paused?: boolean;
  };
}) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const loopMedia = questionType !== "regular";

  useEffect(() => {
    if (!attachment || !videoRef.current) return;
    if (loopMedia) {
      videoRef.current.play().catch(() => {});
    } else {
      const duration = attachment.duration ?? videoRef.current.duration;
      videoRef.current.currentTime = duration * attachmentLastProgress;
    }
  }, []);

  const visibleText = text
    ? text.slice(0, Math.floor(text.length * textLastProgress))
    : "";

  return (
    <div className="relative h-full">
      <TimerBar
        initProgress={timeBar.progress}
        endsAt={timeBar.endsAt}
        paused={timeBar.paused}
      />
      <div className="h-full flex flex-col items-center justify-center p-4 sm:p-6 gap-3 overflow-hidden">
        {attachment && (
          <div className="w-full flex justify-center flex-1 min-h-0">
            {attachment.type === "video" && (
              <video
                ref={videoRef}
                className="w-full h-full object-contain rounded-md"
                src={attachment.url}
                loop={loopMedia}
              />
            )}

            {attachment.type === "audio" && (
              <div className="flex flex-col items-center gap-2">
                <div
                  className={`w-14 h-14 rounded-full bg-surface-raised border border-border flex items-center justify-center text-primary ${
                    loopMedia ? "animate-pulse" : ""
                  }`}
                >
                  <FaMusic size={22} />
                </div>
                {loopMedia && (
                  <p className="text-xs text-on-surface-muted">
                    Audio playing...
                  </p>
                )}
                {loopMedia && (
                  <audio
                    ref={videoRef as React.RefObject<HTMLAudioElement>}
                    src={attachment.url}
                    loop
                    autoPlay
                  />
                )}
              </div>
            )}

            {attachment.type === "image" && (
              <img
                src={attachment.url}
                alt="Question attachment"
                className="w-full h-full object-contain rounded-lg"
              />
            )}
          </div>
        )}
        {text && (
          <p
            className={`shrink-0 text-center font-semibold whitespace-pre-wrap ${textSize(
              text.length
            )}`}
          >
            {visibleText}
          </p>
        )}
      </div>
    </div>
  );
}
