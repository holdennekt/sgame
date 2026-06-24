"use server";

import { cookies } from "next/headers";
import { HiddenPack, Pack } from "@/types/pack";
import { PackDraft } from "@/types/pack_draft";
import { RoomHost, RoomLobby, RoomPlayer } from "@/types/room";
import { User } from "@/middleware";
import { HttpError } from "@/types/error";

const backendUrl = (path: string) =>
  `http://${process.env.BACKEND_HOST}${path}`;

const cookieHeader = () => ({ cookie: cookies().toString() });

const throwHttpError = async (resp: Response) => {
  const body = await resp.json().catch(() => ({}));
  throw new HttpError(resp.status, body.error ?? resp.statusText);
};

export const getRooms = async (): Promise<RoomLobby[]> => {
  const resp = await fetch(backendUrl("/api/rooms"), {
    cache: "no-store",
    headers: cookieHeader(),
  });
  if (!resp.ok) await throwHttpError(resp);
  return resp.json();
};

export const getUser = async (id: string): Promise<User> => {
  const resp = await fetch(backendUrl(`/api/users/${id}`), {
    cache: "no-store",
    headers: cookieHeader(),
  });
  if (!resp.ok) await throwHttpError(resp);
  return resp.json();
};

export const getPack = async (
  id: string
): Promise<Pack | HiddenPack | null> => {
  const resp = await fetch(backendUrl(`/api/packs/${id}`), {
    cache: "no-store",
    headers: cookieHeader(),
  });
  if (resp.status === 403) return null;
  if (!resp.ok) await throwHttpError(resp);
  return resp.json();
};

export const getDraft = async (id: string): Promise<PackDraft> => {
  const resp = await fetch(backendUrl(`/api/packs/drafts/${id}`), {
    cache: "no-store",
    headers: cookieHeader(),
  });
  if (!resp.ok) await throwHttpError(resp);
  return resp.json();
};

// Used for both members (already joined via the lobby) and spectators.
// GET /rooms/:id returns the appropriate projection based on membership.
// For private room spectators, password must be provided.
export const getRoom = async (
  id: string,
  password?: string
): Promise<RoomHost | RoomPlayer> => {
  const url = new URL(backendUrl(`/api/rooms/${id}`));
  if (password) url.searchParams.set("password", password);
  const resp = await fetch(url, {
    cache: "no-store",
    headers: cookieHeader(),
  });
  if (!resp.ok) await throwHttpError(resp);
  return resp.json();
};
