import { Attachment } from "@/types/pack";
import TimerBar from "./TimerBar";
import { FaMusic } from "react-icons/fa6";
import Image from "next/image";

export default function QuestionPanel({
  attachment,
  text,
  timeBar,
}: {
  attachment: Attachment | null;
  text: string;
  timeBar: {
    progress: number;
    durationMs: number;
  };
}) {
  return (
    <div className="relative h-full">
      <TimerBar
        initProgress={timeBar.progress}
        durationMs={timeBar.durationMs}
      />
      <div className="h-full flex flex-col justify-center items-center gap-2 p-10">
        {attachment && (
          <div className="w-full max-w-4xl mb-4">
            {attachment.type === "video" && (
              <video
                className="w-full rounded-md"
                src={attachment.url}
                controls={false}
                autoPlay
                muted
                loop
              />
            )}

            {attachment.type === "audio" && (
              <audio className="w-full" src={attachment.url} controls />
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
        <p className="text-center text-3xl font-semibold">{text}</p>
      </div>
    </div>
  );
}
