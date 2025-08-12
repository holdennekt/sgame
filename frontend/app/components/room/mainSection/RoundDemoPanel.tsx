import React, { useEffect, useRef, useState } from "react";
import { RoundDemo } from "../Room";

export default function RoundDemoPanel({
  roundDemo,
  onFinish,
  speedCharsPerSec = 5,
  pauseBefore = 2000,
  pauseAfter = 2000,
}: {
  roundDemo: RoundDemo;
  onFinish: () => void;
  speedCharsPerSec?: number;
  pauseBefore?: number;
  pauseAfter?: number;
}) {
  const containerRef = useRef<HTMLDivElement>(null);
  const textRef = useRef<HTMLDivElement>(null);
  const [animation, setAnimation] = useState({
    start: false,
    offset: 0,
    duration: 0,
  });

  const text = roundDemo.categories.join(", ");

  useEffect(() => {
    const container = containerRef.current;
    const textEl = textRef.current;
    if (!container || !textEl) return;

    const containerWidth = container.offsetWidth;
    const textWidth = textEl.scrollWidth;
    const overflowWidth = textWidth - containerWidth;

    if (overflowWidth <= 0) {
      setTimeout(onFinish, pauseBefore + pauseAfter);
      return;
    }

    const charCountInOverflow = Math.floor(
      (overflowWidth / textWidth) * text.length
    );
    const duration = (charCountInOverflow / speedCharsPerSec) * 1000;

    const totalDelay = pauseBefore + duration + pauseAfter;
    setTimeout(onFinish, totalDelay);

    const startTimer = setTimeout(() => {
      setAnimation({ start: true, offset: overflowWidth, duration });
    }, pauseBefore);

    return () => clearTimeout(startTimer);
  }, [roundDemo, speedCharsPerSec, pauseBefore, pauseAfter, onFinish]);

  return (
    <div className="w-full h-full flex flex-col justify-center items-center gap-2 p-10">
      <p className="text-4xl font-semibold">{roundDemo.name}</p>
      <div
        ref={containerRef}
        className="max-w-full overflow-hidden whitespace-nowrap text-3xl font-medium"
      >
        <p
          ref={textRef}
          className="inline-block"
          style={{
            transform: animation.start
              ? `translateX(-${animation.offset}px)`
              : "translateX(0)",
            transition: animation.start
              ? `transform ${animation.duration}ms linear`
              : undefined,
          }}
        >
          {text}
        </p>
      </div>
    </div>
  );
}
