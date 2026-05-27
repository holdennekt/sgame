"use client";

import { User } from "@/middleware";
import { signURL } from "@/app/actions";
import { useRef, useState } from "react";
import { toast } from "react-toastify";
import { FiUploadCloud, FiX, FiUser } from "react-icons/fi";

export function Avatar({ user, size = 20 }: { user: User; size?: number }) {
  if (user.avatar)
    return (
      <img
        src={user.avatar}
        alt={user.name}
        style={{ width: size, height: size }}
        className="rounded-full object-cover border border-border"
      />
    );
  return (
    <div
      style={{ width: size, height: size, fontSize: size * 0.38 }}
      className="rounded-full bg-primary flex items-center justify-center text-on-primary font-bold border border-border shrink-0"
    >
      {user.name[0]?.toUpperCase() ?? "?"}
    </div>
  );
}

export function AvatarPicker({
  current,
  onChange,
}: {
  current: string | null;
  onChange: (url: string) => void;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [preview, setPreview] = useState<string | null>(current);
  const [uploading, setUploading] = useState(false);

  const handleFile = async (file: File) => {
    setUploading(true);
    try {
      const { url, formData: fields, getUrl } = await signURL({
        filename: file.name,
        public: true,
      });
      const body = new FormData();
      Object.entries(fields).forEach(([k, v]) => body.append(k, v));
      body.append("file", file);
      const resp = await fetch(url, { method: "POST", body });
      if (!resp.ok) throw new Error("Upload failed");
      setPreview(getUrl!);
      onChange(getUrl!);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Upload failed", {
        containerId: "profile",
      });
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="relative group shrink-0">
      <div
        className="w-24 h-24 rounded-full border-2 border-border cursor-pointer overflow-hidden bg-surface-raised flex items-center justify-center"
        onClick={() => !uploading && fileInputRef.current?.click()}
      >
        {preview ? (
          <img src={preview} alt="avatar" className="w-full h-full object-cover" />
        ) : (
          <FiUser size={36} className="text-on-surface-muted" />
        )}
        <div className="absolute inset-0 rounded-full flex items-center justify-center bg-background/60 opacity-0 group-hover:opacity-100 transition-opacity duration-150">
          {uploading ? (
            <div className="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin" />
          ) : (
            <FiUploadCloud size={22} className="text-on-surface" />
          )}
        </div>
      </div>
      {preview && (
        <button
          type="button"
          title="Remove photo"
          onClick={() => { setPreview(null); onChange(""); }}
          className="absolute -top-1 -right-1 w-5 h-5 rounded-full bg-surface-raised border border-border flex items-center justify-center text-on-surface-muted hover:text-danger hover:border-danger transition-colors duration-150"
        >
          <FiX size={10} />
        </button>
      )}
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) handleFile(file);
          e.target.value = "";
        }}
      />
    </div>
  );
}
