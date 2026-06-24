import { getRoom } from "@/app/server-fetch";
import { HttpError } from "@/types/error";
import ErrorPage from "@/app/error";
import Navbar from "@/components/Navbar";
import { notFound } from "next/navigation";
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
  const result = await getRoom(
    params.id,
    isSpectator ? password : undefined
  ).catch((e: unknown): HttpError => {
    if (e instanceof HttpError) return e;
    throw e;
  });

  if (result instanceof HttpError) {
    if (result.status === 404) notFound();
    return <ErrorPage error={result} reset={() => {}} />;
  }

  return (
    <>
      <Navbar />
      <RoomPage
        initialRoom={result}
        isSpectator={isSpectator}
        password={password}
      />
    </>
  );
}
