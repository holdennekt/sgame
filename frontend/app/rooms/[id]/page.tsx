import { joinRoom } from "@/app/actions";
import Navbar from "@/components/Navbar";
import RoomPage from "./Room";
import { isError } from "@/middleware";

export default async function Page({
  params,
  searchParams,
}: {
  params: { id: string };
  searchParams: { [key: string]: string | undefined };
}) {
  const room = await joinRoom(params.id, searchParams.password);
  if (isError(room)) throw new Error(room.error);

  return (
    <>
      <Navbar />
      <RoomPage initialRoom={room} />
    </>
  );
}
