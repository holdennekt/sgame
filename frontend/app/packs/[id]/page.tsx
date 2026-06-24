import { getPack } from "@/app/server-fetch";
import ErrorPage from "@/app/error";
import Navbar from "@/components/Navbar";
import { notFound } from "next/navigation";
import PackEditor from "../PackEditor";
import HiddenPackView from "./HiddenPackView";
import { convertPackToFormData, isHiddenPack } from "@/types/pack";
import Link from "next/link";
import { FiArrowLeft, FiLock } from "react-icons/fi";
import { HttpError } from "@/types/error";

export const dynamic = "force-dynamic";

export default async function Page({ params }: { params: { id: string } }) {
  const result = await getPack(params.id).catch((e: unknown): HttpError => {
    if (e instanceof HttpError) return e;
    throw e;
  });

  if (result instanceof HttpError) {
    if (result.status === 404) notFound();
    return <ErrorPage error={result} reset={() => {}} />;
  }

  const pack = result;

  return (
    <>
      <Navbar />
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded-md bg-surface text-on-surface p-4 border border-border">
          {pack === null ? (
            <div className="flex flex-col h-full">
              <div className="flex items-center gap-3 mb-5">
                <Link
                  href="/packs"
                  className="flex items-center justify-center w-8 h-8 rounded-lg border border-border text-on-surface-muted hover:bg-surface-raised transition-colors duration-150 shrink-0"
                >
                  <FiArrowLeft size={16} />
                </Link>
              </div>
              <div className="flex flex-1 flex-col items-center justify-center gap-3 text-center">
                <div className="w-12 h-12 rounded-full bg-surface-raised border border-border flex items-center justify-center text-on-surface-muted">
                  <FiLock size={20} />
                </div>
                <div className="flex flex-col gap-1">
                  <p className="text-base font-semibold text-on-surface">
                    This pack is private
                  </p>
                  <p className="text-sm text-on-surface-muted">
                    You don&apos;t have permission to view this pack.
                  </p>
                </div>
              </div>
            </div>
          ) : isHiddenPack(pack) ? (
            <HiddenPackView pack={pack} backHref="/packs" />
          ) : (
            <PackEditor
              initialPack={convertPackToFormData(pack)}
              readOnly
              backHref="/packs"
              packId={params.id}
            />
          )}
        </div>
      </main>
    </>
  );
}
