import { useState } from "react";
import Modal from "@/components/Modal";
import { useRouter } from "next/navigation";
import { joinRoom, checkRoomPassword } from "@/app/api";
import { isError } from "@/middleware";

type Mode = "join" | "spectate";

export default function PasswordModal({
  isOpen,
  close,
  roomId,
  mode = "join",
}: {
  isOpen: boolean;
  close: () => void;
  roomId: string | undefined;
  mode?: Mode;
}) {
  const router = useRouter();
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleConfirm = async () => {
    if (!roomId) return;
    setLoading(true);
    try {
      if (mode === "spectate") {
        await checkRoomPassword(roomId, password);
        router.push(
          `/rooms/${roomId}?spectate=true&password=${encodeURIComponent(
            password
          )}`
        );
      } else {
        await joinRoom(roomId, password);
        router.push(`/rooms/${roomId}`);
      }
    } catch (e) {
      setError(isError(e) ? e.error : "Something went wrong");
      setLoading(false);
    }
  };

  const label = mode === "spectate" ? "Watch" : "Connect";

  return (
    <Modal
      isOpen={isOpen}
      onClose={() => {
        setPassword("");
        setError(null);
        close();
      }}
      closeByClickingOutside
    >
      <h3 className="text-base/7 font-medium">Enter password</h3>

      <div className="w-48 flex flex-col gap-2 flex-1">
        <input
          className="w-full h-9 px-2.5 rounded-lg border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring mt-1 transition-[border-color] duration-150"
          type="text"
          placeholder="Password"
          value={password}
          onChange={(e) => {
            setPassword(e.target.value);
            setError(null);
          }}
          onKeyDown={(e) => {
            if (e.code === "Enter") void handleConfirm();
          }}
        />
        {error && <p className="text-xs text-danger">{error}</p>}
      </div>
      <div className="mt-4 flex flex-row-reverse">
        <button
          className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium cursor-pointer bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 disabled:opacity-50"
          onClick={handleConfirm}
          disabled={loading || !password}
        >
          {loading ? `${label}…` : label}
        </button>
      </div>
    </Modal>
  );
}
