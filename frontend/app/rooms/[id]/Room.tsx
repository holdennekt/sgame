"use client";

import { useEffect, useRef, useState } from "react";
import { toast, ToastContainer } from "react-toastify";
import Chat from "@/components/Chat";
import { getAvatar } from "@/components/UserAvatar";
import { FiEye } from "react-icons/fi";
import { TbRobot } from "react-icons/tb";
import Link from "next/link";
import ControlButtons from "./ControlButtons";
import MainPanel from "./mainSection/MainPanel";
import { RoomModerator, RoomPlayer } from "@/types/room";
import { useRoom } from "@/hooks/useRoom";
import { RoomProvider } from "@/contexts/RoomContext";
import { useRequiredUser } from "@/contexts/UserContext";

const SKIPPABLE_STATES = new Set([
  "revealing_question",
  "showing_question",
  "answering",
  "passing",
  "betting",
]);

export default function RoomPage({
  initialRoom,
  isSpectator = false,
  password,
}: {
  initialRoom: RoomModerator | RoomPlayer;
  isSpectator?: boolean;
  password?: string;
}) {
  const user = useRequiredUser();
  const { lastError, chat, game } = useRoom(
    initialRoom,
    user.id,
    isSpectator,
    password
  );
  const {
    room,
    startGame,
    togglePause,
    leave,
    joinAsPlayer,
    startAnswer,
    submitTypedAnswer,
    banPlayer,
    skipQuestion,
    skipRound,
    changeScore,
    ...gameContext
  } = game;

  const [mobileTab, setMobileTab] = useState<"game" | "chat">("game");
  const [unreadCount, setUnreadCount] = useState(0);
  const isModerator = room.moderator?.id === user.id;
  const isAiHost = room.options.aiHost;
  const canSkip = isModerator && isAiHost && SKIPPABLE_STATES.has(room.state);
  const canSkipRound =
    isModerator && isAiHost && room.state === "selecting_question";
  const mainContainer = useRef<HTMLDivElement>(null);
  const answerButton = useRef<HTMLDivElement>(null);

  useEffect(() => {
    mainContainer.current?.focus();
  }, []);

  useEffect(() => {
    if (!lastError) return;
    toast.error(lastError.msg, { containerId: "room" });
    mainContainer.current?.focus();
  }, [lastError]);

  useEffect(() => {
    if (chat.messages.length === 0) return;
    if (mobileTab !== "game") return;
    if (chat.messages[chat.messages.length - 1].from.id === "") return;
    setUnreadCount((c) => c + 1);
  }, [chat.messages.length]);

  const handleStartAnswer = () => {
    if (
      (isModerator && !room.options.aiHost) ||
      isSpectator ||
      room.pausedState.paused ||
      (room.state !== "revealing_question" && room.state !== "showing_question")
    )
      return;
    answerButton.current?.blur();
    mainContainer.current?.focus();
    startAnswer();
  };

  return (
    <RoomProvider
      value={{
        room,
        ...gameContext,
        answerButton,
        startAnswer: handleStartAnswer,
        submitTypedAnswer,
        banPlayer,
        skipQuestion,
        skipRound,
        changeScore,
      }}
    >
      <main
        className="flex flex-col md:flex-row gap-2 flex-1 min-w-0 min-h-0 p-2 focus:outline-none"
        tabIndex={-1}
        ref={mainContainer}
        onKeyDown={(e) => {
          if (e.code !== "Space") return;
          answerButton.current?.focus();
        }}
        onKeyUp={(e) => {
          if (e.code !== "Space") return;
          handleStartAnswer();
        }}
      >
        {/* Sidebar: room info + tabs (mobile) + chat */}
        <div
          className={`flex flex-col gap-2 min-w-0 md:flex-1 md:min-h-0 md:order-last${
            mobileTab === "chat" ? " flex-1 min-h-0" : ""
          }`}
        >
          <div className="overflow-hidden shrink-0 bg-surface border border-border rounded-md">
            <div className="flex items-center gap-3 p-3">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 min-w-0">
                  <p className="text-base font-semibold truncate text-on-surface">
                    {room.name}
                  </p>
                  {room.spectatorCount > 0 && (
                    <span className="shrink-0 inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-surface-raised text-on-surface-muted">
                      <FiEye size={11} />
                      {room.spectatorCount}
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-1.5 text-xs text-on-surface-muted min-w-0">
                  <span className="shrink-0">Pack:</span>
                  <Link
                    className="text-primary underline underline-offset-2 hover:text-primary-hover truncate"
                    href={`/packs/${room.packPreview.id}`}
                    target="_blank"
                  >
                    {room.packPreview.name}
                  </Link>
                  {isSpectator && (
                    <>
                      <span className="shrink-0">·</span>
                      <span className="shrink-0 text-secondary font-medium">
                        Spectating
                      </span>
                    </>
                  )}
                </div>
              </div>
              <div
                className={`flex flex-col items-center gap-1 shrink-0${
                  room.moderator?.isConnected ? "" : " opacity-50"
                }`}
              >
                <div className="w-9 h-9 rounded-lg overflow-hidden border border-border">
                  {getAvatar(room.moderator)}
                </div>
                <div className="flex items-center gap-1">
                  <p
                    className="text-xs text-on-surface-muted truncate max-w-16 text-center"
                    title={room.moderator?.name}
                  >
                    {room.moderator?.name}
                  </p>
                  {room.options.aiHost && (
                    <span className="inline-flex items-center gap-0.5 px-1 py-0.5 rounded text-[9px] font-bold bg-primary/15 text-primary shrink-0">
                      <TbRobot size={10} />
                      AI
                    </span>
                  )}
                </div>
              </div>
            </div>
            <ControlButtons
              isModerator={isModerator}
              isSpectator={isSpectator}
              isGameStarted={room.state !== "waiting_for_start"}
              isPaused={room.pausedState.paused}
              canSkip={canSkip}
              canSkipRound={canSkipRound}
              start={startGame}
              togglePause={togglePause}
              leave={leave}
              joinAsPlayer={joinAsPlayer}
              skipQuestion={skipQuestion}
              skipRound={skipRound}
            />
          </div>

          {/* Tabs — mobile only */}
          <div className="md:hidden flex rounded-md overflow-hidden border border-border shrink-0">
            <button
              className={`flex-1 py-2 text-sm font-medium transition-colors duration-150${
                mobileTab === "game"
                  ? " bg-primary text-on-primary"
                  : " bg-surface text-on-surface-muted hover:bg-surface-raised"
              }`}
              onClick={() => setMobileTab("game")}
            >
              Game
            </button>
            <button
              className={`flex-1 py-2 text-sm font-medium transition-colors duration-150 inline-flex items-center justify-center gap-1.5${
                mobileTab === "chat"
                  ? " bg-primary text-on-primary"
                  : " bg-surface text-on-surface-muted hover:bg-surface-raised"
              }`}
              onClick={() => {
                setMobileTab("chat");
                setUnreadCount(0);
              }}
            >
              Chat
              {unreadCount > 0 && mobileTab !== "chat" && (
                <span className="min-w-[15px] h-[15px] px-1 rounded-full bg-danger text-white text-[9px] font-bold leading-[15px] text-center">
                  {unreadCount > 99 ? "99+" : unreadCount}
                </span>
              )}
            </button>
          </div>

          <div
            className={`min-w-0 flex-1 min-h-0${
              mobileTab === "game" ? " hidden md:flex" : " flex"
            }`}
          >
            <Chat messages={chat.messages} sendMessage={chat.send} />
          </div>
        </div>

        {/* Main game panel */}
        <div
          className={`min-w-0 min-h-0 gap-2 md:flex md:flex-[3_0_0%] md:flex-col md:order-first${
            mobileTab === "game" ? " flex flex-1 flex-col" : " hidden"
          }`}
        >
          <MainPanel />
        </div>
      </main>
      <ToastContainer
        containerId="room"
        position="bottom-left"
        theme="colored"
      />
    </RoomProvider>
  );
}
