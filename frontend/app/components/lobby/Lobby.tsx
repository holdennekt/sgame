"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { ChatMessage, isChatMessage } from "../Message";
import { isError, User } from "../../../middleware";
import RoomsList from "./RoomsList";
import NewRoomModal from "./NewRoomModal";
import { toast, ToastContainer } from "react-toastify";
import PasswordModal from "./PasswordModal";
import { isRoomLobby, RoomLobby } from "./Room";
import Chat from "../Chat";
import AddButton from "../AddButton";

export type WsMessage = { event: string; payload: unknown };
export type WsMessageHandler = (payload: unknown) => void;

type RoomLobbyDeletedDTO = {
  id: string;
};

const dummyRoomLobbyDeleted: RoomLobbyDeletedDTO = {
  id: "1",
};

export const isRoomLobbyDeleted = (
  obj: unknown
): obj is RoomLobbyDeletedDTO => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyRoomLobbyDeleted).every((key) =>
    Object.hasOwn(obj, key)
  );
};

export default function Lobby({
  user,
  initialRooms,
}: {
  user: User;
  initialRooms: RoomLobby[];
}) {
  const [rooms, setRooms] = useState(initialRooms);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [searchInput, setSearchInput] = useState("");
  const [isNewRoomModalOpen, setIsNewRoomModalOpen] = useState(false);

  const [passwordModal, setPasswordModal] = useState<{
    roomId: string | undefined;
    isOpen: boolean;
  }>({
    roomId: undefined,
    isOpen: false,
  });

  const wsConn = useRef<WebSocket | null>(null);

  useEffect(() => {
    const handlers = new Map<string, WsMessageHandler>();

    handlers.set("room_updated", (payload) => {
      if (!isRoomLobby(payload)) return;
      setRooms((rooms) =>
        rooms.some((room) => room.id === payload.id)
          ? rooms.map((room) => (room.id === payload.id ? payload : room))
          : [...rooms, payload]
      );
    });
    handlers.set("room_deleted", (payload) => {
      if (!isRoomLobbyDeleted(payload)) return;
      setRooms((rooms) => rooms.filter((room) => room.id !== payload.id));
    });
    handlers.set("chat", (payload) => {
      if (!isChatMessage(payload)) return;
      setChatMessages((chatMessages) => [...chatMessages, payload]);
    });
    handlers.set("error", (payload) => {
      if (!isError(payload)) return;
      toast.error(payload.error, { containerId: "lobby" });
    });

    wsConn.current = new WebSocket(
      `ws://${process.env.NEXT_PUBLIC_BACKEND_HOST}/ws/lobby`
    );

    wsConn.current.addEventListener("message", (ev: MessageEvent<string>) => {
      const message: WsMessage = JSON.parse(ev.data);
      const handler = handlers.get(message.event);
      if (handler) handler(message.payload);
      console.log("incoming message", message);
    });

    wsConn.current.addEventListener("close", () => {
      toast.error("Disconnected from server", { containerId: "lobby" });
    });

    return () => {
      wsConn.current?.close();
    };
  }, []);

  const filteredRooms = useMemo(
    () =>
      rooms.filter((room) =>
        room.name.toLowerCase().includes(searchInput.trim().toLowerCase())
      ),
    [rooms, searchInput]
  );

  const sendChatMessage = (text: string) => {
    wsConn.current?.send(JSON.stringify({ event: "chat", payload: { text } }));
  };

  return (
    <>
      <main
        className={`flex flex-col sm:flex-row flex-1 gap-3 min-w-0 min-h-0 
        p-3`}
      >
        <div
          className={`flex flex-col min-w-0 min-h-0 relative 
          max-h-[50%] sm:flex-[1_0_0%] sm:max-h-none`}
        >
          <div
            className={`flex items-center min-h-12 border rounded p-2 
            surface`}
          >
            <input
              className="search-room w-full rounded-lg p-1 text-black"
              placeholder="Search existing rooms"
              value={searchInput}
              onChange={(ev) => setSearchInput(ev.target.value)}
            />
          </div>
          <RoomsList
            rooms={filteredRooms}
            openPasswordModal={(roomId: string) =>
              setPasswordModal({ roomId, isOpen: true })
            }
          />
          <AddButton onClick={() => setIsNewRoomModalOpen(true)} />
        </div>
        <div className="flex min-w-0 min-h-0 flex-[1_0_0%] sm:flex-[2_0_0%]">
          <Chat
            user={user}
            messages={chatMessages}
            sendMessage={sendChatMessage}
          />
        </div>
      </main>
      <NewRoomModal
        isOpen={isNewRoomModalOpen}
        close={() => setIsNewRoomModalOpen(false)}
        toastContainerId="lobby"
      />
      <PasswordModal
        isOpen={passwordModal.isOpen}
        close={() => setPasswordModal({ ...passwordModal, isOpen: false })}
        roomId={passwordModal.roomId}
      />
      <ToastContainer
        containerId="lobby"
        position="bottom-left"
        theme="colored"
      />
    </>
  );
}
