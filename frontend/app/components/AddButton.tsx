import React from "react";

export default function AddButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      className={`new-room flex justify-center items-center w-14 h-14 
              rounded-full absolute bottom-2 right-2 secondary outline-none`}
      onClick={onClick}
    >
      <svg viewBox="0 0 128 128" width="56" height="56">
        <polygon
          points={`60,94 68,94 68,68 94,68 94,60 68,60 68,34 
                60,34 60,60 34,60 34,68 60,68`}
        />
      </svg>
    </button>
  );
}
