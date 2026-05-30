"use client";

import { KeyboardEventHandler, useEffect, useRef, useState } from "react";
import Message, { ChatMessage } from "./Message";
import SystemMessage from "./SystemMessage";
import { RiSendPlaneFill } from "react-icons/ri";
import { useRequiredUser } from "@/contexts/UserContext";

export default function Chat({
  messages,
  sendMessage,
  className = "rounded-md border border-border",
}: {
  messages: ChatMessage[];
  sendMessage: (text: string) => void;
  className?: string;
}) {
  const user = useRequiredUser();
  const [input, setInput] = useState("");
  const scrollableRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    scrollableRef.current?.scroll({
      top: scrollableRef.current?.scrollHeight,
      behavior: "smooth",
    });
  }, [messages]);

  const handleSend = () => {
    if (!input.trim()) return;
    sendMessage(input.trim());
    setInput("");
  };

  const onInputKeyDown: KeyboardEventHandler = (ev) => {
    ev.stopPropagation();
    if (ev.key !== "Enter") return;
    handleSend();
  };

  return (
    <div
      className={`w-full h-full flex flex-col overflow-hidden bg-surface ${className}`}
    >
      <div
        className="flex-1 flex flex-col overflow-y-auto px-3 py-2 min-h-0"
        ref={scrollableRef}
      >
        {messages.length === 0 && (
          <div className="flex-1 flex items-center justify-center">
            <p className="text-sm text-on-surface-muted">No messages yet</p>
          </div>
        )}
        {messages.map((message, index) => {
          const isOwn = user.id === message.from.id;
          const isPrevUserSame =
            message.from.id === messages[index - 1]?.from.id;
          const isNextUserSame =
            message.from.id === messages[index + 1]?.from.id;
          const isSystem = message.from.id === "";
          return isSystem ? (
            <SystemMessage key={index} text={message.text} />
          ) : (
            <Message
              key={index}
              message={message}
              isOwn={isOwn}
              isPrevUserSame={isPrevUserSame}
              isNextUserSame={isNextUserSame}
            />
          );
        })}
      </div>

      <div className="flex gap-2 p-2 border-t border-border">
        <input
          className="flex-1 min-w-0 h-9 px-2.5 rounded-lg border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring transition-[border-color] duration-150"
          placeholder="Say something..."
          value={input}
          onChange={(ev) => setInput(ev.target.value)}
          onKeyDown={onInputKeyDown}
          onKeyUp={(ev) => ev.stopPropagation()}
        />
        <button
          className="h-9 w-9 inline-flex items-center justify-center rounded-lg bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 shrink-0"
          onClick={handleSend}
        >
          <RiSendPlaneFill size={16} />
        </button>
      </div>
    </div>
  );
}
