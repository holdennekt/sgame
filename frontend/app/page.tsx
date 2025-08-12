import Navbar from "./components/Navbar";
import { headers } from "next/headers";
import Lobby from "./components/lobby/Lobby";
import { USER_HEADER_NAME, User } from "../middleware";
import { getRooms } from "./actions";

export default async function Home() {
  const user: User = JSON.parse(headers().get(USER_HEADER_NAME)!);
  const rooms = await getRooms();

  return (
    <>
      <Navbar user={user} />
      <Lobby user={user} initialRooms={rooms} />
    </>
  );
}
