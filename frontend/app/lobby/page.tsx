import Navbar from "@/components/Navbar";
import Lobby from "./Lobby";
import { getRooms } from "@/app/server-fetch";

export default async function LobbyPage() {
  const rooms = await getRooms();

  return (
    <>
      <Navbar />
      <Lobby initialRooms={rooms} />
    </>
  );
}
