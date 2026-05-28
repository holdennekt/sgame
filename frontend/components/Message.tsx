import Link from "next/link";
import { User } from "../middleware";

export type ChatMessage = {
  from: User;
  text: string;
};

const dummyChatMessage: ChatMessage = {
  from: { id: "1", name: "user", avatar: null },
  text: "hello",
};

export const isChatMessage = (obj: unknown): obj is ChatMessage => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyChatMessage).every(key => Object.hasOwn(obj, key));
};

export default function Message({
  message,
  isOwn,
  isPrevUserSame,
  isNextUserSame,
}: {
  message: ChatMessage;
  isOwn: boolean;
  isPrevUserSame: boolean;
  isNextUserSame: boolean;
}) {
  const initials = message.from.name
    .split(" ")
    .map(word => word[0].toUpperCase())
    .join("");

  const avatar = message.from.avatar ? (
    <img
      className="h-9 w-9 aspect-square rounded-full object-cover shrink-0"
      src={message.from.avatar}
      alt="avatar"
    />
  ) : (
    <div className="flex justify-center items-center h-9 w-9 rounded-full bg-primary text-on-primary text-xs font-bold shrink-0">
      {initials}
    </div>
  );

  const spacer = <div className="h-9 w-9 shrink-0" />;

  return (
    <div
      className={`flex items-end gap-2 ${isPrevUserSame ? "mt-0.5" : "mt-2"}${
        isOwn ? " flex-row-reverse" : ""
      }`}
    >
      {!isOwn && (isNextUserSame ? spacer : avatar)}
      <div className={`flex flex-col max-w-[72%]${isOwn ? " items-end" : " items-start"}`}>
        <div
          className={`text-sm break-words px-3 py-1.5 ${
            isOwn
              ? "bg-primary text-on-primary rounded-t-2xl rounded-bl-2xl rounded-br-md"
              : "bg-surface-raised text-on-surface rounded-t-2xl rounded-br-2xl rounded-bl-md"
          }`}
        >
          {!isOwn && !isPrevUserSame && (
            message.from.isGuest ? (
              <span className="block text-xs font-semibold text-primary mb-0.5">
                {message.from.name}
              </span>
            ) : (
              <Link
                href={`/user/${message.from.id}`}
                target="_blank"
                className="block text-xs font-semibold text-primary mb-0.5 hover:underline"
              >
                {message.from.name}
              </Link>
            )
          )}
          {message.text}
        </div>
      </div>
    </div>
  );
}
