import { getUser } from "@/app/server-fetch";
import { HttpError } from "@/types/error";
import ErrorPage from "@/app/error";
import Navbar from "@/components/Navbar";
import { USER_HEADER_NAME, User } from "@/middleware";
import { headers } from "next/headers";
import { notFound } from "next/navigation";
import ProfilePage from "./ProfilePage";

export default async function Page({ params }: { params: { id: string } }) {
  const user: User = JSON.parse(
    decodeURIComponent(headers().get(USER_HEADER_NAME)!)
  );
  const isOwn = user.id === params.id;
  const result = isOwn
    ? user
    : await getUser(params.id).catch((e: unknown): HttpError => {
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
      <main className="flex-1 min-w-0 min-h-0 overflow-y-auto">
        <ProfilePage user={result} isOwn={isOwn} />
      </main>
    </>
  );
}
