import React, { useState } from "react";
import Modal from "../Modal";
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
          className="w-full h-8 rounded-lg mt-1 p-1 text-black"
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
          className="primary rounded-md p-1 text-sm font-normal"
          href={roomLink}
        >
          Connect
        </Link>
      </div>
    </Modal>
  );
}
