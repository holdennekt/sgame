import { getPack, updatePack } from "@/app/actions";
import Navbar from "@/components/Navbar";
import PackEditor from "@/components/pack/PackEditor";
import { USER_HEADER_NAME, User } from "@/middleware";
import { CreatePackRequest } from "@/types/pack";
import { headers } from "next/headers";

export default async function Page({
  params,
  searchParams,
}: {
  params: { id: string };
  searchParams: { [key: string]: string | undefined };
}) {
  const user: User = JSON.parse(headers().get(USER_HEADER_NAME)!);
  const pack = await getPack(params.id);

  const handleUpdatePack = async (pack: CreatePackRequest) => {
    "use server";
    return await updatePack(params.id, pack);
  };

  return (
    <>
      <Navbar user={user} />
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded surface p-4">
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
