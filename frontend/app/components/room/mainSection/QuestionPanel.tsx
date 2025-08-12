import TimerBar from "./TimerBar";

export default function QuestionPanel({
  text,
  timeBar,
}: {
  text: string;
  timeBar?: {
    progress: number;
    durationMs: number;
  };
}) {
  return (
    <div className="relative h-full">
      {timeBar && (
        <TimerBar
          initProgress={timeBar.progress}
          durationMs={timeBar.durationMs}
        />
      )}
      <div className="h-full flex flex-col justify-center items-center gap-2 p-10">
        <p className="text-center text-3xl font-semibold">{text}</p>
      </div>
    </div>
  );
}
