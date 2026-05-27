import { useEffect, useRef, useState } from "react";

export default function TimerBar({
  initProgress,
  durationMs,
  paused,
}: {
  initProgress: number;
  durationMs: number;
  paused?: boolean;
}) {
  const [progress, setProgress] = useState(initProgress);
  const reqIdRef = useRef<number>();

  useEffect(() => {
    if (reqIdRef.current) cancelAnimationFrame(reqIdRef.current);
    if (paused || initProgress <= 0) return;

    const totalTime = durationMs / initProgress;
    const start = performance.now() - totalTime * (1 - initProgress);

    const tick = (now: number) => {
      const elapsed = now - start;
      const newProgress = Math.max(1 - elapsed / totalTime, 0);
      setProgress(newProgress);
      if (newProgress > 0) {
        reqIdRef.current = requestAnimationFrame(tick);
      }
    };

    reqIdRef.current = requestAnimationFrame(tick);

    return () => {
      if (reqIdRef.current) cancelAnimationFrame(reqIdRef.current);
    };
  }, [initProgress, durationMs, paused]);

  return (
    <div className="absolute top-0 left-0 h-1 bg-primary transition-none rounded-full"
      style={{ width: `${progress * 100}%` }}
    />
  );
}
