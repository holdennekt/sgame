import React from "react";

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
    paused: {
      text: "Continue",
      func: togglePause,
    },
    started: {
      text: "Pause",
      func: togglePause,
    },
    idle: {
      text: "Start",
      func: start,
    },
  };
  const { text, func } = stateToContent[gameState];
  return (
    <table className="w-full table-fixed">
      <tbody>
        <tr>
          <td className="text-center text-lg font-semibold border-r border-t p-2">
            <button
              className={`w-full h-full text-center text-lg font-semibold${
                isHost ? "" : " hidden"
              }`}
              onClick={func}
            >
              {text}
            </button>
          </td>
          <td className="border-l border-t p-2">
            <button
              className="w-full h-full text-center text-lg font-semibold"
              onClick={leave}
            >
              Leave
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  );
}
