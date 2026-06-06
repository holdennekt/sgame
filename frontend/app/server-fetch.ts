"use server";

import { cookies } from "next/headers";
import { Pack } from "@/types/pack";
import { PackDraft } from "@/types/pack_draft";
import { Room, RoomLobby } from "@/types/room";
import { User } from "@/middleware";

const backendUrl = (path: string) =>
  `http://${process.env.BACKEND_HOST}${path}`;

const cookieHeader = () => ({ cookie: cookies().toString() });

export const getRooms = async (): Promise<RoomLobby[]> => {
  const resp = await fetch(backendUrl("/api/rooms"), {
    cache: "no-store",
    headers: cookieHeader(),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const getUser = async (id: string): Promise<User> => {
  const resp = await fetch(backendUrl(`/api/users/${id}`), {
    cache: "no-store",
    headers: cookieHeader(),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const getPack = async (id: string): Promise<Pack> => {
  const resp = await fetch(backendUrl(`/api/packs/${id}`), {
    cache: "no-store",
    headers: cookieHeader(),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const getDraft = async (id: string): Promise<PackDraft> => {
  const resp = await fetch(backendUrl(`/api/packs/drafts/${id}`), {
    cache: "no-store",
    headers: cookieHeader(),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const joinRoom = async (
  id: string,
  password: string | undefined
): Promise<Room> => {
  const url = new URL(backendUrl(`/api/rooms/${id}/join`));
  if (password) url.searchParams.set("password", password);
  const resp = await fetch(url, {
    method: "PATCH",
    cache: "no-store",
    headers: cookieHeader(),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};
