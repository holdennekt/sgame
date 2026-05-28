import TimerBar from "./TimerBar";

export default function TextPanel({
  topText,
  bottomText,
  timeBar,
}: {
  topText: string;
  bottomText?: string | null;
  timeBar?: {
    progress: number;
    endsAt: number;
    paused?: boolean;
  };
}) {
  return (
    <div className="relative w-full h-full flex flex-col justify-center items-center gap-2 p-10">
      {timeBar && (
        <TimerBar
          initProgress={timeBar.progress}
          endsAt={timeBar.endsAt}
          paused={timeBar.paused}
        />
      )}
      <p className="text-center text-3xl font-semibold">{topText}</p>
      {bottomText && (
        <p className="text-center text-2xl font-normal">{bottomText}</p>
      )}
    </div>
  );
}
