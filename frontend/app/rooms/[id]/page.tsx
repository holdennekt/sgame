import { getRoom } from "@/app/server-fetch";
import Navbar from "@/components/Navbar";
import RoomPage from "./Room";

export default async function Page({
  params,
  searchParams,
}: {
  params: { id: string };
  searchParams: { [key: string]: string | undefined };
}) {
  const isSpectator = searchParams.spectate === "true";
  const password = searchParams.password;
  const room = await getRoom(params.id, isSpectator ? password : undefined);

  return (
    <>
      <Navbar />
      <RoomPage
        initialRoom={room}
        isSpectator={isSpectator}
        password={password}
      />
    </>
  );
}
