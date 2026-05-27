import { useState } from "react";
import { Player } from "@/types/user";

export default function BettingPanel({
  player,
  allowedToBet,
  placeBet,
}: {
  player: Player;
  allowedToBet: boolean;
  placeBet: (amount: number) => void;
}) {
  const [amount, setAmount] = useState(0);

  return (
    <div
      className={`w-full h-12 flex items-center gap-3 bg-surface rounded-md px-3 transition-opacity duration-150 ${
        !allowedToBet ? "opacity-50 pointer-events-none select-none" : ""
      }`}
    >
      <input
        className="flex-1 accent-[var(--primary)] cursor-pointer"
        type="range"
        min={0}
        max={player.score}
        step={1}
        value={amount}
        onChange={e => setAmount(Number(e.target.value))}
        disabled={!allowedToBet}
      />
      <input
        className="w-16 h-8 px-2 rounded-lg border border-border bg-background text-on-background text-sm text-center outline-none focus-ring transition-[border-color] duration-150 tabular-nums"
        type="text"
        inputMode="numeric"
        pattern="[0-9]*"
        value={amount}
        onChange={e => {
          const onlyNums = e.target.value.replace(/[^0-9]/g, "");
          setAmount(Math.min(Number(onlyNums), player.score));
        }}
        onKeyDown={e => {
          if (e.key !== "Enter") return;
          placeBet(amount);
        }}
        disabled={!allowedToBet}
      />
      <button
        className="h-8 px-4 rounded-lg bg-primary text-on-primary text-sm font-medium hover:bg-primary-hover transition-colors duration-150 shrink-0"
        onClick={() => placeBet(amount)}
        disabled={!allowedToBet}
      >
        Bet
      </button>
    </div>
  );
}
