import Navbar from "@/components/Navbar";
import { USER_HEADER_NAME, User } from "@/middleware";
import { headers } from "next/headers";
import GameLog from "./GameLog";

export default function Page({ params }: { params: { id: string } }) {
  const user: User = JSON.parse(
    decodeURIComponent(headers().get(USER_HEADER_NAME)!)
  );

  return (
    <>
      <Navbar />
      <main className="flex-1 min-w-0 min-h-0 overflow-y-auto">
        <GameLog id={params.id} backHref={`/user/${user.id}`} />
      </main>
    </>
  );
}
