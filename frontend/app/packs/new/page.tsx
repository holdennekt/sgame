import { createPack } from "@/app/actions";
import Navbar from "@/components/Navbar";
import PackEditor from "../PackEditor";
import { Pack } from "@/types/pack";

export default function Page() {
  const initialPack: Omit<Pack, "id" | "createdBy"> = {
    name: "",
    type: "public",
    rounds: [{ name: "Round 1", categories: [] }],
    finalRound: { categories: [] },
  };

  return (
    <>
      <Navbar />
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded-md bg-surface text-on-surface p-4 border border-border">
          <PackEditor savePack={createPack} initialPack={initialPack} />
        </div>
      </main>
    </>
  );
}
