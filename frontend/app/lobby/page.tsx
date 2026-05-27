import Navbar from "@/components/Navbar";
import Lobby from "./Lobby";
import { getRooms } from "@/app/actions";
import { isError } from "@/middleware";

export default async function LobbyPage() {
  const rooms = await getRooms();
  if (isError(rooms)) throw new Error(rooms.error);

  return (
    <>
      <Navbar />
      <Lobby initialRooms={rooms} />
    </>
  );
}
