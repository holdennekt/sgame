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
      className={`w-full h-12 flex gap-4 surface p-2 ${
        !allowedToAnswer ? "opacity-50 pointer-events-none select-none" : ""
      }`}
    >
      <input
        className="flex-1 rounded-lg p-1 text-black"
        type="text"
        name="value"
        value={answer}
        onChange={(e) => setAnswer(e.target.value)}
        onKeyDown={(e) => {
          if (e.key !== "Enter") return;
          submitFinalRoundAnswer(answer);
        }}
        disabled={!allowedToAnswer}
      />
      <button
        className="primary rounded-lg font-medium px-4"
        onClick={() => setAnswer(answer)}
        disabled={!allowedToAnswer}
      >
        Answer
      </button>
    </div>
  );
}
