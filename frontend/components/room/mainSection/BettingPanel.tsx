import { useState } from "react";
import { Player } from "./PlayerTable";

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
      className={`w-full h-12 flex gap-4 surface p-2 ${
        !allowedToBet ? "opacity-50 pointer-events-none select-none" : ""
      }`}
    >
      <input
        className="flex-1 rounded-lg p-1 text-black"
        type="range"
        min={0}
        max={player.score}
        step={1}
        value={amount}
        onChange={e => setAmount(Number(e.target.value))}
        disabled={!allowedToBet}
      />
      <input
        className="rounded-lg p-1 text-black"
        type="text"
        inputMode="numeric"
        pattern="[0-9]*"
        name="value"
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
        className="primary rounded-lg font-medium px-4"
        onClick={() => placeBet(amount)}
        disabled={!allowedToBet}
      >
        Bet
      </button>
    </div>
  );
}
