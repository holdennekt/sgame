"use client";

import { KeyboardEventHandler, useEffect, useRef, useState } from "react";
import Message, { ChatMessage } from "./Message";
import { User } from "../../middleware";
import SystemMessage from "./SystemMessage";

export default function Chat({
  user,
  messages,
  sendMessage,
}: {
  user: User;
  messages: ChatMessage[];
  sendMessage: (text: string) => void;
}) {
  const [input, setInput] = useState("");
  const scrollableRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    scrollableRef.current?.scroll({
      top: scrollableRef.current?.scrollHeight,
      behavior: "smooth",
    });
  }, [messages]);

  const handleSend = () => {
    if (!input) return;
    sendMessage(input);
    setInput("");
  };

  const onInputKeyDown: KeyboardEventHandler = ev => {
    if (ev.key !== "Enter") return;
    handleSend();
  };

  return (
    <div className="w-full h-full flex flex-col rounded surface">
      <div
        className="flex flex-col flex-1 gap-1.5 overflow-x-auto p-2"
        ref={scrollableRef}
      >
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
      <div className="flex flex-row gap-2 h-12 border rounded p-2">
        <input
          className="flex-1 rounded-lg p-1 text-black"
          placeholder="Say something to others"
          value={input}
          onChange={ev => setInput(ev.target.value)}
          onKeyDown={onInputKeyDown}
        />
        <button
          className="w-20 primary rounded-lg font-medium"
          onClick={handleSend}
        >
          Send
        </button>
      </div>
    </div>
  );
}
