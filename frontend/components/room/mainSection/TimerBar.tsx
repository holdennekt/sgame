import { useEffect, useRef, useState } from "react";

export default function TimerBar({
  initProgress,
  durationMs,
}: {
  initProgress: number;
  durationMs: number;
}) {
  const [progress, setProgress] = useState(initProgress);
  const startTimeRef = useRef<number | null>(null);
  const reqIdRef = useRef<number>();

  useEffect(() => {
    const totalTime = durationMs / initProgress;
    const start = performance.now() - totalTime * (1 - initProgress);
    startTimeRef.current = start;

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
  }, [initProgress, durationMs]);

  return (
    <div
      className="absolute h-3 bg-white"
      style={{ width: `${progress * 100}%` }}
    />
  );
}
