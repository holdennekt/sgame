import { getPack, updatePack } from "@/app/actions";
import Navbar from "@/components/Navbar";
import PackEditor from "../PackEditor";
import { CreatePackRequest } from "@/types/pack";

export default async function Page({
  params,
  searchParams,
}: {
  params: { id: string };
  searchParams: { [key: string]: string | undefined };
}) {
  const pack = await getPack(params.id);

  const handleUpdatePack = async (pack: CreatePackRequest) => {
    "use server";
    return await updatePack(params.id, pack);
  };

  return (
    <>
      <Navbar />
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded-md bg-surface text-on-surface p-4 border border-border">
          <PackEditor
            savePack={handleUpdatePack}
            initialPack={pack}
            readOnly={!searchParams.edit}
          />
        </div>
      </main>
    </>
  );
}
