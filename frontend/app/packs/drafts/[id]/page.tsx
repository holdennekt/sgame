import { getDraft } from "@/app/server-fetch";
import { HttpError } from "@/types/error";
import ErrorPage from "@/app/error";
import Navbar from "@/components/Navbar";
import { notFound } from "next/navigation";
import PackEditor from "../../PackEditor";
import { convertDraftToFormData } from "@/types/pack_draft";

export default async function DraftPage({
  params,
}: {
  params: { id: string };
}) {
  const result = await getDraft(params.id).catch((e: unknown): HttpError => {
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
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded-md bg-surface text-on-surface p-4 border border-border">
          <PackEditor
            initialPack={convertDraftToFormData(result)}
            backHref="/packs/drafts"
            draftId={params.id}
          />
        </div>
      </main>
    </>
  );
}
