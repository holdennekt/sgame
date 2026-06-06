import { getDraft } from "@/app/server-fetch";
import Navbar from "@/components/Navbar";
import PackEditor from "../../PackEditor";
import { convertDraftToFormData } from "@/types/pack_draft";

export default async function DraftPage({
  params,
}: {
  params: { id: string };
}) {
  const draft = await getDraft(params.id);

  return (
    <>
      <Navbar />
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded-md bg-surface text-on-surface p-4 border border-border">
          <PackEditor
            initialPack={convertDraftToFormData(draft)}
            backHref="/packs/drafts"
            draftId={params.id}
          />
        </div>
      </main>
    </>
  );
}
