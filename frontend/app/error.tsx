"use client";

import { useEffect } from "react";
import Navbar from "../components/Navbar";
import { FiAlertTriangle, FiRefreshCw } from "react-icons/fi";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <>
      <Navbar />
      <main className="flex flex-1 items-center justify-center p-4">
        <div className="flex flex-col items-center gap-4 text-center max-w-sm">
          <div className="w-12 h-12 rounded-full bg-surface-raised border border-border flex items-center justify-center text-danger">
            <FiAlertTriangle size={22} />
          </div>
          <div className="flex flex-col gap-1">
            <h2 className="text-base font-semibold text-on-surface">Something went wrong</h2>
            {error.message && (
              <p className="text-sm text-on-surface-muted">{error.message}</p>
            )}
          </div>
          <button
            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
            onClick={() => reset()}
          >
            <FiRefreshCw size={14} />
            Try again
          </button>
        </div>
      </main>
    </>
  );
}
