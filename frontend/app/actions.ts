"use server";

import { ErrorBody, isError } from "@/middleware";
import { PacksResp } from "./components/PacksList";
import { cookies } from "next/headers";
import { Pack, PackDTO } from "./components/pack/PackEditor";
import { RoomLobby } from "./components/lobby/Room";
import { CreateRoomParams, PackPreview } from "./components/lobby/NewRoomModal";
import { Room } from "./components/room/Room";

const PAGE_QUERY_PARAM = "page";
const PASSWORD_QUERY_PARAM = "password";
const SEARCH_QUERY_PARAM = "search";

export const getRooms = async () => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/rooms`);
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  const rooms: RoomLobby[] | ErrorBody = await resp?.json();
  if (isError(rooms)) throw new Error(rooms.error);
  return rooms;
};

export const createRoom = async (body: CreateRoomParams) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/rooms`);
  const resp = await fetch(url, {
    method: "POST",
    body: JSON.stringify(body),
    headers: { cookie: cookies().toString() },
  });
  const obj: { id: string } | ErrorBody = await resp?.json();
  if (isError(obj)) throw new Error(obj.error);
  return obj.id;
};

export const joinRoom = async (id: string, password: string | undefined) => {
  const url = new URL(
    `http://${process.env.BACKEND_HOST}/api/rooms/${id}/join`
  );
  if (password) url.searchParams.set(PASSWORD_QUERY_PARAM, password);
  const resp = await fetch(url, {
    method: "PATCH",
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  }).catch(console.log);
  const room: Room | ErrorBody = await resp?.json();
  if (isError(room)) throw new Error(room.error);
  return room;
};

export const leaveRoom = async (id: string) => {
  const url = new URL(
    `http://${process.env.BACKEND_HOST}/api/rooms/${id}/leave`
  );
  const resp = await fetch(url, {
    method: "PATCH",
    headers: { cookie: cookies().toString() },
  });
  if (!resp.ok) {
    const obj = await resp?.json();
    if (isError(obj)) throw new Error(obj.error);
  }
};

export const getPacks = async (packFilter: string, page?: number) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs`);
  url.searchParams.set(SEARCH_QUERY_PARAM, packFilter);
  if (page) url.searchParams.set(PAGE_QUERY_PARAM, page.toString());
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  const packs: PacksResp | ErrorBody = await resp?.json();
  if (isError(packs)) throw new Error(packs.error);
  return packs;
};

export const getPacksPreviews = async (packFilter: string) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packsPreview`);
  url.searchParams.set(SEARCH_QUERY_PARAM, packFilter);
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  const packs: PackPreview[] | ErrorBody = await resp?.json();
  if (isError(packs)) throw new Error(packs.error);
  return packs;
};

export const getPack = async (id: string) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/${id}`);
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  const pack: Pack | ErrorBody = await resp?.json();
  if (isError(pack)) throw new Error(pack.error);
  return pack;
};

export const createPack = async (pack: PackDTO) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs`);
  const resp = await fetch(url, {
    method: "POST",
    headers: { cookie: cookies().toString() },
    body: JSON.stringify(pack),
  });
  const obj: { id: string } | ErrorBody = await resp?.json();
  if (isError(obj)) throw new Error(obj.error);
  return obj;
};

export const updatePack = async (pack: Pack) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/${pack.id}`);
  const resp = await fetch(url, {
    method: "PUT",
    headers: { cookie: cookies().toString() },
    body: JSON.stringify(pack),
  });
  const obj: Record<string, never> | ErrorBody = await resp?.json();
  if (isError(obj)) throw new Error(obj.error);
  return { id: pack.id };
};
