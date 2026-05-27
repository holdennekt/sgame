import { useState } from "react";

export default function FinalRoundAnswerPanel({
  allowedToAnswer,
  submitFinalRoundAnswer,
}: {
  allowedToAnswer: boolean;
  submitFinalRoundAnswer: (answer: string) => void;
}) {
  const [answer, setAnswer] = useState("");

  return (
    <div
      className={`w-full h-12 flex gap-4 bg-surface text-on-surface p-2 ${
        !allowedToAnswer ? "opacity-50 pointer-events-none select-none" : ""
      }`}
    >
      <input
        className="flex-1 rounded-lg px-2.5 bg-background text-on-background border border-border outline-none text-sm placeholder:text-on-surface-muted focus-ring transition-[border-color] duration-150"
        type="text"
        name="value"
        value={answer}
        onChange={e => setAnswer(e.target.value)}
        onKeyDown={e => {
          if (e.key !== "Enter") return;
          submitFinalRoundAnswer(answer);
        }}
        disabled={!allowedToAnswer}
      />
      <button
        className="h-9 px-4 rounded-lg bg-primary text-on-primary text-sm font-medium hover:bg-primary-hover transition-colors duration-150 shrink-0"
        onClick={() => submitFinalRoundAnswer(answer)}
        disabled={!allowedToAnswer}
      >
        Answer
      </button>
    </div>
  );
}
