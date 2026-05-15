import { createPack } from "@/app/actions";
import Navbar from "@/components/Navbar";
import PackEditor from "@/components/pack/PackEditor";
import { USER_HEADER_NAME, User } from "@/middleware";
import { Pack } from "@/types/pack";
import { headers } from "next/headers";

export default function Page() {
  const user: User = JSON.parse(headers().get(USER_HEADER_NAME)!);
  const initialPack: Omit<Pack, "id" | "createdBy"> = {
    name: "",
    type: "public",
    rounds: [{ name: "Round 1", categories: [] }],
    finalRound: { categories: [] },
  };

  return (
    <>
      <Navbar user={user} />
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded surface p-4">
          <p className="text-xl font-semibold leading-none mb-2">
            Create your pack:
          </p>
          <PackEditor savePack={createPack} initialPack={initialPack} />
        </div>
      </main>
    </>
  );
}
