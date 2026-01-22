import Link from "next/link";
import Private from "@/public/private.png";
import Public from "@/public/public.png";
import Image from "next/image";
import { User } from "../../../middleware";

export type RoomLobby = {
  id: string;
  name: string;
  packPreview: { id: string; name: string; };
  host: User | null;
  players: User[];
  maxPlayers: number;
  type: "public" | "private";
  status: string;
};

const dummyRoom: RoomLobby = {
  id: "1",
  name: "xyz",
  packPreview: { id: "1", name: "wtf" },
  players: [],
  maxPlayers: 3,
  type: "public",
  status: "Idle",
  host: null,
};

export const isRoomLobby = (obj: unknown): obj is RoomLobby => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyRoom).every(key => Object.hasOwn(obj, key));
};

export const getAvatar = (user: User | null) => {
  if (!user) return <div></div>;

  const imgAvatar = (
    <Image className="w-full aspect-square" src={user.avatar!} alt="avatar" />
  );
  const divAvatar = (
    <div className="w-full aspect-square flex justify-center items-center bg-indigo-500">
      {user.name
        .split(" ")
        .map(word => word[0].toUpperCase())
        .join("")}
    </div>
  );

  return user.avatar ? imgAvatar : divAvatar;
};

export default function RoomLobbyPage({
  room,
  openPasswordModal,
}: {
  room: RoomLobby;
  openPasswordModal: (roomId: string) => void;
}) {
  const playersSlots = new Array<User | null>(room.maxPlayers).fill(null);
  for (const [index, player] of room.players.entries()) {
    playersSlots[index] = player;
  }

  return (
    <div className="surface rounded-lg">
      <div className="flex justify-between">
        <div className="flex items-center">
          <p
            className="w-fit inline-block p-2 text-lg leading-none truncate font-semibold"
            title={room.name}
          >
            {room.name}
          </p>
          <Image
            className="w-5 h-5 inline-block"
            src={room.type === "public" ? Public : Private}
            alt={room.type}
          />
        </div>
        <p className="w-fit p-2 text-lg leading-none">
          {room.players.length}/{room.maxPlayers}
        </p>
      </div>
      <p className="px-2 text-sm font-normal">
        Pack:{" "}
        <Link
          className="pack-link"
          href={`/packs/${room.packPreview.id}`}
          target="_blank"
        >
          {room.packPreview.name}
        </Link>
      </p>
      <div className="flex items-center mt-2 px-2">
        <div className="h-7 w-7 border">
          {getAvatar(room.host)}
        </div>
        <div className="flex-1 min-w-7 min-h-7"></div>
        <div className="flex-initial overflow-x-auto">
          <table className="border">
            <tbody>
              <tr>
                {playersSlots.map((player, index) => (
                  <td key={index} className="h-7 w-7 border p-0">
                    {getAvatar(player)}
                  </td>
                ))}
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div className="flex justify-between items-center p-2">
        <p className="text-sm">
          Status:{" "}
          <span
            className={`italic ${
              room.status === "playing" ? "text-green-600" : "text-yellow-600"
            }`}
          >
            {room.status}
          </span>
        </p>
        {room.type === "public" ? (
          <Link
            className="primary rounded-md p-1 text-sm font-normal"
            href={`/rooms/${room.id}`}
          >
            Join
          </Link>
        ) : (
          <button
            className="primary rounded-md p-1 text-sm font-normal"
            onClick={() => openPasswordModal(room.id)}
          >
            Join
          </button>
        )}
      </div>
    </div>
  );
}
