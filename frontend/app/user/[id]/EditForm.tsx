"use client";

import { User } from "@/middleware";
import { AvatarPicker } from "./Avatar";
import { FormEventHandler, useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "react-toastify";
import { FiCheck } from "react-icons/fi";
import { updateUser } from "@/app/actions";
import { isError } from "@/middleware";

const inputCls =
  "h-9 w-full px-2.5 rounded-lg border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring transition-[border-color] duration-150";
const labelCls = "flex flex-col gap-1";
const labelTextCls =
  "text-xs font-medium text-on-surface-muted uppercase tracking-wide";

export function EditForm({ user, onDone }: { user: User; onDone: () => void }) {
  const router = useRouter();
  const [saving, setSaving] = useState(false);
  const [avatarUrl, setAvatarUrl] = useState(user.avatar ?? "");
  const [confirmPw, setConfirmPw] = useState("");

  const onSubmit: FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const name = fd.get("name") as string;
    const password = fd.get("password") as string;
    if (password && password !== confirmPw) {
      toast.error("Passwords do not match", { containerId: "profile" });
      return;
    }
    setSaving(true);
    try {
      const body: { name: string; avatar: string; password?: string } = {
        name,
        avatar: avatarUrl,
      };
      if (password) body.password = password;
      const result = await updateUser(user.id, body);
      if (isError(result)) {
        toast.error(result.error, { containerId: "profile" });
        return;
      }
      toast.success("Saved", { containerId: "profile" });
      router.refresh();
      onDone();
    } finally {
      setSaving(false);
    }
  };

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-5">
      <div className="flex flex-col sm:flex-row items-start gap-6">
        <AvatarPicker current={user.avatar} onChange={setAvatarUrl} />

        <div className="flex-1 grid grid-cols-1 sm:grid-cols-2 gap-4 w-full">
          <label className={labelCls}>
            <span className={labelTextCls}>Display name</span>
            <input
              className={inputCls}
              type="text"
              name="name"
              defaultValue={user.name}
              minLength={1}
              maxLength={20}
              required
            />
          </label>
          <div />
          <label className={labelCls}>
            <span className={labelTextCls}>New password</span>
            <input
              className={inputCls}
              type="password"
              name="password"
              placeholder="Leave blank to keep"
              minLength={8}
              maxLength={40}
            />
          </label>
          <label className={labelCls}>
            <span className={labelTextCls}>Confirm password</span>
            <input
              className={inputCls}
              type="password"
              value={confirmPw}
              onChange={(e) => setConfirmPw(e.target.value)}
              placeholder="Repeat new password"
            />
          </label>
        </div>
      </div>

      <div className="flex items-center justify-end gap-2">
        <button
          type="button"
          onClick={onDone}
          className="px-3.5 py-1.5 rounded-lg text-sm font-medium border border-border text-on-surface-muted hover:bg-surface-raised transition-colors duration-150"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={saving}
          className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 disabled:opacity-50"
        >
          <FiCheck size={14} />
          {saving ? "Saving…" : "Save"}
        </button>
      </div>
    </form>
  );
}
