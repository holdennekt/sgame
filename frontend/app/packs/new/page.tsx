import { createPack } from "@/app/actions";
import Navbar from "@/app/components/Navbar";
import PackEditor from "@/app/components/pack/PackEditor";
import { USER_HEADER_NAME, User } from "@/middleware";
import { headers } from "next/headers";

export default function Page() {
  const user: User = JSON.parse(headers().get(USER_HEADER_NAME)!);

  return (
    <>
      <Navbar user={user} />
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded surface p-4">
          <p className="text-xl font-semibold leading-none mb-2">
            Create your pack:
          </p>
          <PackEditor savePack={createPack} />
        </div>
      </main>
    </>
  );
}
