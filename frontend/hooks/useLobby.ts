import { useState } from "react";
import { ChatMessage, isChatMessage } from "@/components/Message";
import { isError } from "@/middleware";
import { RoomLobby, isRoomLobby } from "@/types/room";
import { useWebSocket } from "./useWebSocket";

type RoomLobbyDeletedDTO = { id: string };

const isRoomLobbyDeleted = (obj: unknown): obj is RoomLobbyDeletedDTO => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.hasOwn(obj, "id");
};

export function useLobby(initialRooms: RoomLobby[]) {
  const [rooms, setRooms] = useState(initialRooms);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [lastError, setLastError] = useState<{
    msg: string;
    count: number;
  } | null>(null);

  const { wsConn, handlers } = useWebSocket("/api/ws/lobby", "lobby");

  const setError = (msg: string) =>
    setLastError((prev) => ({ msg, count: (prev?.count ?? 0) + 1 }));

  handlers.set("room_updated", (payload) => {
    if (!isRoomLobby(payload)) return;
    setRooms((rooms) =>
      rooms.some((r) => r.id === payload.id)
        ? rooms.map((r) => (r.id === payload.id ? payload : r))
        : [...rooms, payload]
    );
  });
  handlers.set("room_deleted", (payload) => {
    if (!isRoomLobbyDeleted(payload)) return;
    setRooms((rooms) => rooms.filter((r) => r.id !== payload.id));
  });
  handlers.set("chat", (payload) => {
    if (!isChatMessage(payload)) return;
    setChatMessages((msgs) => [...msgs, payload]);
  });
  handlers.set("error", (payload) => {
    if (!isError(payload)) return;
    setError(payload.error);
  });

  const sendChatMessage = (text: string) =>
    wsConn.current?.send(JSON.stringify({ event: "chat", payload: { text } }));

  return { rooms, chatMessages, lastError, sendChatMessage };
}
