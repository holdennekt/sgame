import { getPack } from "@/app/server-fetch";
import Navbar from "@/components/Navbar";
import PackEditor from "../PackEditor";
import { convertPackToFormData } from "@/types/pack";

export const dynamic = "force-dynamic";

export default async function Page({ params }: { params: { id: string } }) {
  const pack = await getPack(params.id);

  return (
    <>
      <Navbar />
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded-md bg-surface text-on-surface p-4 border border-border">
          <PackEditor
            initialPack={convertPackToFormData(pack)}
            readOnly
            backHref="/packs"
            packId={params.id}
          />
        </div>
      </main>
    </>
  );
}
