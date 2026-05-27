import { getPack, updatePack } from "@/app/actions";
import Navbar from "@/components/Navbar";
import PackEditor from "../PackEditor";
import { CreatePackRequest } from "@/types/pack";
import { isError } from "@/middleware";

export default async function Page({
  params,
  searchParams,
}: {
  params: { id: string };
  searchParams: { [key: string]: string | undefined };
}) {
  const pack = await getPack(params.id);
  if (isError(pack)) throw new Error(pack.error);

  return (
    <>
      <Navbar />
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded-md bg-surface text-on-surface p-4 border border-border">
          <PackEditor
            savePack={updatePack.bind(null, params.id)}
            initialPack={pack}
            readOnly={!searchParams.edit}
          />
        </div>
      </main>
    </>
  );
}
