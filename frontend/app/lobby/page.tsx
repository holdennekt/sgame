import Navbar from "@/components/Navbar";
import Lobby from "./Lobby";
import { getRooms } from "@/app/server-fetch";
import { HttpError } from "@/types/error";
import ErrorPage from "@/app/error";

export default async function LobbyPage() {
  const result = await getRooms().catch((e: unknown): HttpError => {
    if (e instanceof HttpError) return e;
    throw e;
  });

  if (result instanceof HttpError)
    return <ErrorPage error={result} reset={() => {}} />;

  return (
    <>
      <Navbar />
      <Lobby initialRooms={result} />
    </>
  );
}
