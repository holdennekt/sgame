import { useEffect, useRef, useState } from "react";

export default function TimerBar({
  initProgress,
  endsAt,
  paused,
}: {
  initProgress: number;
  endsAt: number;
  paused?: boolean;
}) {
  const [progress, setProgress] = useState(initProgress);
  const reqIdRef = useRef<number>();

  useEffect(() => {
    if (reqIdRef.current) cancelAnimationFrame(reqIdRef.current);
    if (paused || initProgress <= 0) return;

    const remaining = endsAt - Date.now();
    if (remaining <= 0) { setProgress(0); return; }

    const totalDuration = remaining / initProgress;
    const start = performance.now() - (totalDuration - remaining);

    const tick = (now: number) => {
      const newProgress = Math.max(1 - (now - start) / totalDuration, 0);
      setProgress(newProgress);
      if (newProgress > 0) {
        reqIdRef.current = requestAnimationFrame(tick);
      }
    };

    reqIdRef.current = requestAnimationFrame(tick);
    return () => {
      if (reqIdRef.current) cancelAnimationFrame(reqIdRef.current);
    };
  }, [endsAt, initProgress, paused]);

  return (
    <div className="absolute top-0 left-0 h-1 bg-primary transition-none rounded-full"
      style={{ width: `${progress * 100}%` }}
    />
  );
}
