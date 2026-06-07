"use client";

import { User } from "@/middleware";
import { signURL } from "@/app/api";
import { isError } from "@/middleware";
import { useEffect, useRef, useState } from "react";
import { toast } from "react-toastify";
import { FiUploadCloud, FiX, FiUser, FiCheck } from "react-icons/fi";

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

const CROP_SIZE = 288;

function CropModal({
  src,
  onConfirm,
  onCancel,
}: {
  src: string;
  onConfirm: (blob: Blob) => void;
  onCancel: () => void;
}) {
  const [scale, setScale] = useState(1);
  const [offset, setOffset] = useState({ x: 0, y: 0 });
  const dragRef = useRef<{
    startX: number;
    startY: number;
    ox: number;
    oy: number;
  } | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const imgRef = useRef<HTMLImageElement>(null);

  useEffect(() => {
    const img = imgRef.current;
    if (!img) return;
    const init = () => {
      const cover = Math.max(
        CROP_SIZE / img.naturalWidth,
        CROP_SIZE / img.naturalHeight
      );
      setScale(cover);
    };
    if (img.complete && img.naturalWidth) init();
    else img.addEventListener("load", init, { once: true });
  }, []);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const handler = (e: WheelEvent) => {
      e.preventDefault();
      setScale((s) => Math.min(10, Math.max(0.2, s * (1 - e.deltaY * 0.001))));
    };
    el.addEventListener("wheel", handler, { passive: false });
    return () => el.removeEventListener("wheel", handler);
  }, []);

  const onMouseDown = (e: React.MouseEvent) => {
    dragRef.current = {
      startX: e.clientX,
      startY: e.clientY,
      ox: offset.x,
      oy: offset.y,
    };
  };
  const onMouseMove = (e: React.MouseEvent) => {
    if (!dragRef.current) return;
    setOffset({
      x: dragRef.current.ox + e.clientX - dragRef.current.startX,
      y: dragRef.current.oy + e.clientY - dragRef.current.startY,
    });
  };
  const stopDrag = () => {
    dragRef.current = null;
  };

  const onTouchStart = (e: React.TouchEvent) => {
    const t = e.touches[0];
    dragRef.current = {
      startX: t.clientX,
      startY: t.clientY,
      ox: offset.x,
      oy: offset.y,
    };
  };
  const onTouchMove = (e: React.TouchEvent) => {
    if (!dragRef.current || e.touches.length !== 1) return;
    const t = e.touches[0];
    setOffset({
      x: dragRef.current.ox + t.clientX - dragRef.current.startX,
      y: dragRef.current.oy + t.clientY - dragRef.current.startY,
    });
  };

  const handleConfirm = () => {
    const img = imgRef.current;
    if (!img) return;
    const canvas = document.createElement("canvas");
    canvas.width = CROP_SIZE;
    canvas.height = CROP_SIZE;
    const ctx = canvas.getContext("2d")!;
    const imgW = img.naturalWidth * scale;
    const imgH = img.naturalHeight * scale;
    ctx.drawImage(
      img,
      CROP_SIZE / 2 + offset.x - imgW / 2,
      CROP_SIZE / 2 + offset.y - imgH / 2,
      imgW,
      imgH
    );
    canvas.toBlob(
      (blob) => {
        if (blob) onConfirm(blob);
      },
      "image/jpeg",
      0.92
    );
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="flex flex-col gap-5 p-6 rounded-2xl bg-surface border border-border shadow-2xl w-full max-w-sm mx-4">
        <p className="text-sm font-semibold text-on-surface">Crop photo</p>

        <div className="flex justify-center">
          <div
            ref={containerRef}
            className="relative overflow-hidden cursor-grab active:cursor-grabbing select-none"
            style={{
              width: CROP_SIZE,
              height: CROP_SIZE,
              borderRadius: "50%",
              border: "2px solid var(--color-border)",
            }}
            onMouseDown={onMouseDown}
            onMouseMove={onMouseMove}
            onMouseUp={stopDrag}
            onMouseLeave={stopDrag}
            onTouchStart={onTouchStart}
            onTouchMove={onTouchMove}
            onTouchEnd={stopDrag}
          >
            <img
              ref={imgRef}
              src={src}
              alt=""
              draggable={false}
              style={{
                position: "absolute",
                top: "50%",
                left: "50%",
                transform: `translate(calc(-50% + ${offset.x}px), calc(-50% + ${offset.y}px)) scale(${scale})`,
                transformOrigin: "center",
                maxWidth: "none",
                pointerEvents: "none",
              }}
            />
          </div>
        </div>

        <div className="flex items-center gap-3 px-1">
          <span className="text-xs text-on-surface-muted select-none">−</span>
          <input
            type="range"
            min={0.2}
            max={10}
            step={0.01}
            value={scale}
            onChange={(e) => setScale(Number(e.target.value))}
            className="flex-1 accent-primary"
          />
          <span className="text-xs text-on-surface-muted select-none">+</span>
        </div>

        <p className="text-xs text-on-surface-muted text-center -mt-2">
          Drag to reposition · scroll or slider to zoom
        </p>

        <div className="flex gap-2 justify-end">
          <button
            type="button"
            onClick={onCancel}
            className="px-3.5 py-1.5 rounded-lg text-sm font-medium border border-border text-on-surface-muted hover:bg-surface-raised transition-colors duration-150"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleConfirm}
            className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
          >
            <FiCheck size={14} />
            Apply
          </button>
        </div>
      </div>
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
  const [cropSrc, setCropSrc] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);

  const handleFileSelect = (file: File) => {
    const reader = new FileReader();
    reader.onload = (e) => setCropSrc(e.target?.result as string);
    reader.readAsDataURL(file);
  };

  const handleCropConfirm = async (blob: Blob) => {
    setCropSrc(null);
    setUploading(true);
    try {
      const {
        url,
        formData: fields,
        getUrl,
      } = await signURL({ filename: "avatar.jpg", public: true });
      const body = new FormData();
      Object.entries(fields).forEach(([k, v]) => body.append(k, v));
      body.append("file", blob, "avatar.jpg");
      const resp = await fetch(url, { method: "POST", body });
      if (!resp.ok) throw new Error("Upload failed");
      setPreview(getUrl!);
      onChange(getUrl!);
    } catch (e) {
      toast.error(
        isError(e) ? e.error : e instanceof Error ? e.message : "Upload failed",
        { containerId: "profile" }
      );
    } finally {
      setUploading(false);
    }
  };

  return (
    <>
      {cropSrc && (
        <CropModal
          src={cropSrc}
          onConfirm={handleCropConfirm}
          onCancel={() => setCropSrc(null)}
        />
      )}

      <div className="relative group shrink-0">
        <div
          className="w-24 h-24 rounded-full border-2 border-border cursor-pointer overflow-hidden bg-surface-raised flex items-center justify-center"
          onClick={() => !uploading && fileInputRef.current?.click()}
        >
          {preview ? (
            <img
              src={preview}
              alt="avatar"
              className="w-full h-full object-cover"
            />
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
            onClick={() => {
              setPreview(null);
              onChange("");
            }}
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
            if (file) handleFileSelect(file);
            e.target.value = "";
          }}
        />
      </div>
    </>
  );
}
