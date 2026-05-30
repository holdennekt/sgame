import React from "react";
import { FiLogOut, FiPlay, FiPause } from "react-icons/fi";

export default function ControlButtons({
  isHost,
  isGameStarted,
  isPaused,
  start,
  togglePause,
  leave,
}: {
  isHost: boolean;
  isGameStarted: boolean;
  isPaused: boolean;
  start: () => void;
  togglePause: () => void;
  leave: () => void;
}) {
  const gameState = isGameStarted ? (isPaused ? "paused" : "started") : "idle";
  const stateToContent = {
    paused: { text: "Continue", icon: <FiPlay size={14} />, func: togglePause },
    started: { text: "Pause", icon: <FiPause size={14} />, func: togglePause },
    idle: { text: "Start", icon: <FiPlay size={14} />, func: start },
  };
  const { text, icon, func } = stateToContent[gameState];

  return (
    <div className="flex gap-2 p-2 border-t border-border">
      {isHost && (
        <button
          className="inline-flex items-center justify-center gap-1.5 flex-1 px-3.5 py-1.5 rounded-lg text-sm font-medium cursor-pointer bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
          onClick={func}
        >
          {icon}
          {text}
        </button>
      )}
      <button
        className="inline-flex items-center justify-center gap-1.5 flex-1 px-3.5 py-1.5 rounded-lg text-sm font-medium cursor-pointer bg-surface-raised text-danger border border-danger hover:bg-border transition-colors duration-150"
        onClick={leave}
      >
        <FiLogOut size={14} />
        Leave
      </button>
    </div>
  );
}
