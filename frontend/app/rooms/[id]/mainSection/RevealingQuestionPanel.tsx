import { Attachment } from "@/types/pack";
import { useEffect, useRef, useState } from "react";
import { FaMusic, FaVideo } from "react-icons/fa6";

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

export default function RevealingQuestionPanel({
  attachment,
  attachmentEndsAt,
  attachmentLastProgress,
  text,
  textEndsAt,
  textLastProgress,
  paused,
  category,
  value,
}: {
  attachment: Attachment | null;
  attachmentEndsAt: string;
  attachmentLastProgress: number;
  text: string | null;
  textEndsAt: string;
  textLastProgress: number;
  paused?: boolean;
  category: string;
  value: number;
}) {
  const [audioBlocked, setAudioBlocked] = useState(false);
  const [currentText, setCurrentText] = useState(
    text ? text.slice(0, Math.floor(text.length * textLastProgress)) : ""
  );
  const mediaRef = useRef<HTMLVideoElement | HTMLAudioElement>(null);
  const textIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const stableAttachmentRef = useRef(attachment);
  if (attachment?.url !== stableAttachmentRef.current?.url) {
    stableAttachmentRef.current = attachment;
  }
  const stableAttachment = stableAttachmentRef.current;

  const startTextReveal = () => {
    if (!text) return;
    if (textIntervalRef.current) {
      clearInterval(textIntervalRef.current);
    }

    const timeUntilEnd = new Date(textEndsAt).getTime() - Date.now();

    if (timeUntilEnd <= 0) return setCurrentText(text);

    const startIndex = Math.floor(text.length * textLastProgress);
    const remainingChars = text.length - startIndex;

    if (remainingChars <= 0) return setCurrentText(text);

    const updateInterval = timeUntilEnd / remainingChars;

    setCurrentText(text.slice(0, startIndex));
    let charIndex = startIndex;

    textIntervalRef.current = setInterval(() => {
      charIndex++;
      if (charIndex >= text.length) {
        if (textIntervalRef.current) {
          clearInterval(textIntervalRef.current);
        }
        return setCurrentText(text);
      }
      setCurrentText(text.slice(0, charIndex));
    }, updateInterval);
  };

  useEffect(() => {
    if (paused) return;
    if (!stableAttachment) return startTextReveal();

    const timer = setTimeout(
      startTextReveal,
      new Date(attachmentEndsAt).getTime() - Date.now()
    );

    const cleanupFuncs = [() => clearTimeout(timer)];

    if (
      ["video", "audio"].includes(stableAttachment.type) &&
      new Date(attachmentEndsAt).getTime() > Date.now()
    ) {
      const mediaDuration =
        stableAttachment.duration ?? mediaRef.current!.duration;
      mediaRef.current!.currentTime = mediaDuration * attachmentLastProgress;
      const playPromise = mediaRef
        .current!.play()
        .catch(() => setAudioBlocked(true));

      cleanupFuncs.push(() => {
        playPromise.then(() => {
          mediaRef.current?.pause();
          if (mediaRef.current) mediaRef.current.currentTime = 0;
        });
      });
    }

    return () => cleanupFuncs.forEach((fn) => fn());
  }, [stableAttachment, attachmentEndsAt, attachmentLastProgress, paused]);

  useEffect(() => {
    if (!paused) return;
    if (textIntervalRef.current) clearInterval(textIntervalRef.current);
    mediaRef.current?.pause();
  }, [paused]);

  useEffect(
    () => () => {
      if (textIntervalRef.current) clearInterval(textIntervalRef.current);
    },
    []
  );

  const handleUnblock = () => {
    mediaRef.current?.play().catch(() => {});
    setAudioBlocked(false);
  };

  return (
    <div className="relative h-full flex flex-col">
      <div className="flex-1 flex flex-col items-center justify-center p-4 sm:p-6 gap-3 overflow-hidden">
        <p className="shrink-0 text-sm font-medium text-on-surface-muted">
          {category} — {value}
        </p>
        {attachment && (
          <div className="w-full flex justify-center items-center flex-1 min-h-0">
            {attachment.type === "video" && (
              <video
                ref={mediaRef as React.RefObject<HTMLVideoElement>}
                className="w-full h-full object-contain rounded-md"
                src={attachment.url}
                onEnded={startTextReveal}
                controls={false}
              />
            )}

            {attachment.type === "audio" && (
              <div className="flex flex-col items-center gap-2">
                <div className="w-14 h-14 rounded-full bg-surface-raised border border-border flex items-center justify-center text-primary animate-pulse">
                  <FaMusic size={22} />
                </div>
                <audio
                  ref={mediaRef as React.RefObject<HTMLAudioElement>}
                  src={attachment.url}
                  onEnded={startTextReveal}
                />
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
            )} min-h-[1em]`}
          >
            {currentText}
          </p>
        )}
      </div>
      {attachment && audioBlocked && (
        <button
          onClick={handleUnblock}
          className="absolute bottom-3 right-3 flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-surface-raised border border-border text-xs font-medium text-on-surface-muted hover:text-on-surface transition-colors"
        >
          {attachment.type === "video" ? (
            <FaVideo size={11} />
          ) : (
            <FaMusic size={11} />
          )}
          Tap to play {attachment.type}
        </button>
      )}
    </div>
  );
}
