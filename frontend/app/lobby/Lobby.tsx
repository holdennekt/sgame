"use client";

import { useEffect, useMemo, useState } from "react";
import RoomsList from "./RoomsList";
import NewRoomModal from "@/components/NewRoomModal";
import { toast, ToastContainer } from "react-toastify";
import PasswordModal from "./PasswordModal";
import Chat from "@/components/Chat";
import { RoomLobby } from "@/types/room";
import { IoIosSearch, IoIosAdd } from "react-icons/io";
import { useLobby } from "@/hooks/useLobby";

type Tab = "rooms" | "chat";

export default function Lobby({ initialRooms }: { initialRooms: RoomLobby[] }) {
  const { rooms, chatMessages, lastError, sendChatMessage } = useLobby(initialRooms);

  const [searchInput, setSearchInput] = useState("");
  const [isNewRoomModalOpen, setIsNewRoomModalOpen] = useState(false);
  const [mobileTab, setMobileTab] = useState<Tab>("rooms");
  const [unreadCount, setUnreadCount] = useState(0);

  const tabs: Tab[] = ["rooms", "chat"];

  const [passwordModal, setPasswordModal] = useState<{
    roomId: string | undefined;
    isOpen: boolean;
  }>({ roomId: undefined, isOpen: false });

  useEffect(() => {
    if (!lastError) return;
    toast.error(lastError.msg, { containerId: "lobby" });
  }, [lastError]);

  useEffect(() => {
    if (chatMessages.length === 0) return;
    if (mobileTab !== "rooms") return;
    setUnreadCount((c) => c + 1);
  }, [chatMessages.length]);

  const filteredRooms = useMemo(
    () =>
      rooms.filter((room) =>
        room.name.toLowerCase().includes(searchInput.trim().toLowerCase()),
      ),
    [rooms, searchInput],
  );

  return (
    <>
      {/* Mobile tab bar */}
      <div className="sm:hidden flex shrink-0 border-b border-border bg-surface">
        {tabs.map((tab) => (
          <button
            key={tab}
            className={`flex-1 py-2.5 text-sm font-medium capitalize transition-colors duration-150 inline-flex items-center justify-center gap-1.5 border-b-2 ${
              mobileTab === tab
                ? "text-primary border-primary"
                : "text-on-surface-muted border-transparent"
            }`}
            onClick={() => { setMobileTab(tab); if (tab === "chat") setUnreadCount(0); }}
          >
            {tab}
            {tab === "chat" && unreadCount > 0 && mobileTab !== "chat" && (
              <span className="min-w-[15px] h-[15px] px-1 rounded-full bg-danger text-white text-[9px] font-bold leading-[15px] text-center">
                {unreadCount > 99 ? "99+" : unreadCount}
              </span>
            )}
          </button>
        ))}
      </div>

      <main className="flex flex-col sm:flex-row flex-1 min-w-0 min-h-0 sm:gap-2 sm:p-2">
        {/* Rooms panel */}
        <div
          className={`flex-col min-w-0 min-h-0 bg-surface sm:flex sm:flex-[1_0_0%] sm:rounded-md sm:border sm:border-border ${
            mobileTab === "rooms" ? "flex flex-1" : "hidden"
          }`}
        >
          <div className="flex items-center gap-2 p-2 border-b border-border shrink-0">
            <div className="relative flex-1">
              <div className="pointer-events-none absolute inset-y-0 left-2.5 flex items-center text-on-surface-muted">
                <IoIosSearch size={16} />
              </div>
              <input
                className="w-full h-9 pl-8 pr-3 rounded-lg border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring transition-[border-color] duration-150"
                placeholder="Search rooms..."
                value={searchInput}
                onChange={(ev) => setSearchInput(ev.target.value)}
              />
            </div>
            <button
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 shrink-0"
              onClick={() => setIsNewRoomModalOpen(true)}
            >
              <IoIosAdd size={16} />
              New room
            </button>
          </div>
          <RoomsList
            rooms={filteredRooms}
            openPasswordModal={(roomId: string) =>
              setPasswordModal({ roomId, isOpen: true })
            }
          />
        </div>

        {/* Chat panel */}
        <div
          className={`min-w-0 min-h-0 sm:flex sm:flex-[2_0_0%] ${
            mobileTab === "chat" ? "flex flex-1" : "hidden"
          }`}
        >
          <Chat
            messages={chatMessages}
            sendMessage={sendChatMessage}
          />
        </div>
      </main>

      <NewRoomModal
        isOpen={isNewRoomModalOpen}
        close={() => setIsNewRoomModalOpen(false)}
      />
      <PasswordModal
        isOpen={passwordModal.isOpen}
        close={() => setPasswordModal({ ...passwordModal, isOpen: false })}
        roomId={passwordModal.roomId}
      />
      <ToastContainer containerId="lobby" position="bottom-left" theme="colored" />
    </>
  );
}
