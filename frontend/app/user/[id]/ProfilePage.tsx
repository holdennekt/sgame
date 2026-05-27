"use client";

import { User } from "@/middleware";
import { PackPreview } from "@/types/pack";
import { useState } from "react";
import { toast, ToastContainer } from "react-toastify";
import { FiEdit2, FiLogOut, FiTrash2 } from "react-icons/fi";
import { useRouter } from "next/navigation";
import { logout, deleteUser } from "@/app/actions";
import { isError } from "@/middleware";
import { EditForm } from "./EditForm";
import PacksTab from "./PacksTab";
import HistoryTab from "./HistoryTab";
import NewRoomModal from "@/components/NewRoomModal";

type Tab = "packs" | "history";

export default function ProfilePage({
  user,
  isOwn,
}: {
  user: User;
  isOwn: boolean;
}) {
  const router = useRouter();
  const [editing, setEditing] = useState(false);
  const [tab, setTab] = useState<Tab>("packs");

  const handleLogout = async () => {
    await logout();
    router.push("/login");
  };

  const handleDeleteAccount = async () => {
    if (!confirm("Delete your account? This cannot be undone.")) return;
    await logout();
    const result = await deleteUser(user.id);
    if (isError(result)) {
      toast.error(result.error, { containerId: "profile" });
      return;
    }
    router.push("/login");
  };
  const [newRoomModal, setNewRoomModal] = useState<{
    isOpen: boolean;
    pack?: PackPreview;
  }>({
    isOpen: false,
  });

  const tabs: Tab[] = isOwn ? ["packs", "history"] : ["packs"];

  return (
    <div className="max-w-3xl mx-auto px-4 py-5 sm:py-8 flex flex-col gap-5 sm:gap-6">

      {/* Profile header */}
      {editing ? (
        <div className="border border-border rounded-lg p-5 bg-surface">
          <EditForm user={user} onDone={() => setEditing(false)} />
        </div>
      ) : (
        <div className="flex items-center gap-5">
          {user.avatar ? (
            <img
              src={user.avatar}
              alt={user.name}
              className="w-20 h-20 rounded-full object-cover border-2 border-border shrink-0"
            />
          ) : (
            <div className="w-20 h-20 rounded-full bg-surface-raised border-2 border-border flex items-center justify-center shrink-0 text-2xl font-bold text-on-surface-muted">
              {user.name[0]?.toUpperCase() ?? "?"}
            </div>
          )}

          <div className="flex-1 min-w-0">
            <h1 className="text-2xl font-bold text-on-background truncate">
              {user.name}
            </h1>
          </div>

          {isOwn && (
            <div className="flex items-center gap-2 shrink-0">
              <button
                onClick={() => setEditing(true)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium border border-border text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150"
              >
                <FiEdit2 size={13} />
                Edit
              </button>
              <button
                onClick={handleLogout}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium border border-border text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150"
                title="Log out"
              >
                <FiLogOut size={13} />
                Log out
              </button>
              <button
                onClick={handleDeleteAccount}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium border border-border text-danger hover:bg-danger/10 transition-colors duration-150"
                title="Delete account"
              >
                <FiTrash2 size={13} />
                Delete account
              </button>
            </div>
          )}
        </div>
      )}

      {/* Tabs */}
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-1 border-b border-border">
          {tabs.map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-4 py-2 text-sm font-medium capitalize transition-colors duration-150 border-b-2 -mb-px ${
                tab === t
                  ? "border-primary text-primary"
                  : "border-transparent text-on-surface-muted hover:text-on-surface"
              }`}
            >
              {t}
            </button>
          ))}
        </div>

        {tab === "packs" && (
          <PacksTab
            userId={user.id}
            isOwn={isOwn}
            onPlay={(pack) => setNewRoomModal({ isOpen: true, pack })}
          />
        )}

        {tab === "history" && <HistoryTab />}
      </div>

      <NewRoomModal
        isOpen={newRoomModal.isOpen}
        close={() => setNewRoomModal({ isOpen: false })}
        fixedPack={newRoomModal.pack}
      />
      <ToastContainer
        containerId="profile"
        position="bottom-left"
        theme="colored"
      />
    </div>
  );
}
