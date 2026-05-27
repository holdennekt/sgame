import React, { useState } from "react";
import Modal from "@/components/Modal";
import { useRouter } from "next/navigation";
import Link from "next/link";

export default function PasswordModal({
  isOpen,
  close,
  roomId,
}: {
  isOpen: boolean;
  close: () => void;
  roomId: string | undefined;
}) {
  const router = useRouter();
  const [password, setPassword] = useState("");

  const roomLink = `/rooms/${roomId}?password=${password}`;

  return (
    <Modal
      isOpen={isOpen}
      onClose={() => {
        setPassword("");
        close();
      }}
      closeByClickingOutside
    >
      <h3 className="text-base/7 font-medium">
        Enter password
      </h3>

      <div className="w-48 flex flex-col gap-2 flex-1">
        <input
          className="w-full h-9 px-2.5 rounded-lg border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring mt-1 transition-[border-color] duration-150"
          type="text"
          placeholder="Password"
          value={password}
          onChange={e => setPassword(e.target.value)}
          onKeyDown={e => {
            if (e.code === "Enter") router.push(roomLink);
          }}
        />
      </div>
      <div className="mt-4 flex flex-row-reverse">
        <Link
          className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium cursor-pointer bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
          href={roomLink}
        >
          Connect
        </Link>
      </div>
    </Modal>
  );
}
