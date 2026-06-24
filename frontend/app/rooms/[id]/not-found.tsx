import Navbar from "@/components/Navbar";
import Link from "next/link";

export default function RoomNotFound() {
  return (
    <>
      <Navbar />
      <main className="flex flex-1 items-center justify-center p-4">
        <div className="flex flex-col items-center gap-4 text-center max-w-sm">
          <div className="flex flex-col gap-1">
            <h2 className="text-base font-semibold text-on-surface">
              Room not found
            </h2>
            <p className="text-sm text-on-surface-muted">
              This room does not exist or has already ended.
            </p>
          </div>
          <Link
            href="/lobby"
            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
          >
            Back to lobby
          </Link>
        </div>
      </main>
    </>
  );
}
