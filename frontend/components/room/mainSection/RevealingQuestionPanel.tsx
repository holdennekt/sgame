import { Attachment } from "@/types/pack";
import Image from "next/image";
import { useEffect, useRef, useState } from "react";
import { FaMusic } from "react-icons/fa6";

export default function RevealingQuestionPanel({
  attachment,
  attachmentEndsAt,
  attachmentLastProgress,
  text,
  textEndsAt,
  textLastProgress,
}: {
  attachment: Attachment | null;
  attachmentEndsAt: Date;
  attachmentLastProgress: number;
  text: string;
  textEndsAt: Date;
  textLastProgress: number;
}) {
  const [currentText, setCurrentText] = useState(
    text.slice(0, Math.floor(text.length * textLastProgress)),
  );
  const mediaRef = useRef<HTMLVideoElement | HTMLAudioElement>(null);
  const textIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const startTextReveal = () => {
    if (textIntervalRef.current) {
      clearInterval(textIntervalRef.current);
    }

    const now = new Date();
    const timeUntilEnd = textEndsAt.getTime() - now.getTime();

    if (timeUntilEnd <= 0) return setCurrentText(text);

    const remainingChars = text.length - currentText.length;
    const updateInterval = timeUntilEnd / remainingChars;

    let charIndex = currentText.length;

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
    if (!attachment) return startTextReveal();

    if (mediaRef.current && attachment) {
      const mediaDuration = attachment.duration ?? mediaRef.current.duration;
      mediaRef.current.currentTime = mediaDuration * attachmentLastProgress;

      if (attachmentLastProgress < 1) {
        mediaRef.current.play();

        return () => {
          if (!mediaRef.current) return;
          mediaRef.current.pause();
          mediaRef.current.currentTime = 0;
        };
      }
    }
  }, [attachment, attachmentLastProgress]);

  useEffect(
    () => () => {
      if (textIntervalRef.current) clearInterval(textIntervalRef.current);
    },
    [],
  );

  useEffect(() => {
    if (attachment?.type === "image") {
      const now = new Date();
      const timeUntilAttachmentEnds =
        attachmentEndsAt.getTime() - now.getTime();

      if (timeUntilAttachmentEnds <= 0) return startTextReveal();

      const timer = setTimeout(() => {
        startTextReveal();
      }, timeUntilAttachmentEnds);

      return () => clearTimeout(timer);
    }
  }, [attachment, attachmentEndsAt]);

  return (
    <div className="relative h-full">
      <div className="h-full flex flex-col justify-center items-center gap-2 p-10">
        {attachment && (
          <div className="w-full max-w-4xl mb-4">
            {attachment.type === "video" && (
              <video
                ref={mediaRef as React.RefObject<HTMLVideoElement>}
                className="w-full rounded-md"
                src={attachment.url}
                onEnded={startTextReveal}
                controls={false}
              />
            )}

            {attachment.type === "audio" && (
              <div className="w-full flex justify-center">
                <FaMusic className="w-1/6" />
                <audio
                  ref={mediaRef as React.RefObject<HTMLAudioElement>}
                  className="w-full"
                  src={attachment.url}
                  onEnded={startTextReveal}
                  controls={false}
                />
              </div>
            )}

            {attachment.type === "image" && (
              <img
                src={attachment.url}
                alt="Question attachment"
                className="w-full rounded-lg object-contain max-h-96"
              />
            )}
          </div>
        )}

        <p className="text-center text-3xl font-semibold whitespace-pre-wrap">
          {currentText}
        </p>
      </div>
    </div>
  );
}
