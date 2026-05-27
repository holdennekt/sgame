import { User } from "@/middleware";

export const getAvatar = (user: User | null) => {
  if (!user) return <div className="w-full h-full bg-surface-raised" />;

  if (user.avatar) {
    return (
      <img
        className="w-full aspect-square object-cover"
        src={user.avatar}
        alt="avatar"
      />
    );
  }

  return (
    <div className="w-full aspect-square flex justify-center items-center text-xs font-bold bg-primary text-on-primary">
      {user.name
        .split(" ")
        .map((word) => word[0].toUpperCase())
        .join("")}
    </div>
  );
};
