"use server";

import { splitCookiesString, parse } from "set-cookie-parser";
import { ErrorBody, isError } from "@/middleware";
import { cookies } from "next/headers";
import {
  CreatePackRequest,
  HiddenPack,
  Pack,
  PackPreview,
  SignURLRequest,
  SignURLResponse,
} from "@/types/pack";
import { CreateRoomRequest, Room, RoomLobby } from "@/types/room";
import { SearchResponse } from "@/types/search";

const PAGE_QUERY_PARAM = "page";
const PASSWORD_QUERY_PARAM = "password";
const SEARCH_QUERY_PARAM = "search";
const FILENAME_QUERY_PARAM = "filename";
const PUBLIC_QUERY_PARAM = "public";

const passCookies = (resp: Response) => {
  const setCookie = resp.headers.get("set-cookie");
  if (setCookie) {
    parse(splitCookiesString(setCookie)).forEach(cookie => {
      cookies().set(cookie.name, cookie.value, {
        maxAge: cookie.maxAge,
        path: cookie.path,
        domain: cookie.domain,
        secure: cookie.secure,
        httpOnly: cookie.httpOnly,
      });
    });
  }
};

export const login = async (body: { login: string; password: string }) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/login`);
  const resp = await fetch(url, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(body),
  });
  passCookies(resp);

  const obj: { id: string } | ErrorBody = await resp?.json();
  if (isError(obj)) throw new Error(obj.error);
  return obj.id;
};

export const register = async (body: { login: string; password: string }) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/register`);
  const resp = await fetch(url, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(body),
  });
  passCookies(resp);

  const obj: { id: string } | ErrorBody = await resp?.json();
  if (isError(obj)) throw new Error(obj.error);
  return obj.id;
};

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

export const createRoom = async (body: CreateRoomRequest) => {
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
    `http://${process.env.BACKEND_HOST}/api/rooms/${id}/join`,
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
    `http://${process.env.BACKEND_HOST}/api/rooms/${id}/leave`,
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
  const packs: SearchResponse<HiddenPack> | ErrorBody = await resp?.json();
  if (isError(packs)) throw new Error(packs.error);
  return packs;
};

export const getPacksPreviews = async (packFilter: string) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/previews`);
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

export const createPack = async (pack: CreatePackRequest) => {
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

export const updatePack = async (id: string, pack: CreatePackRequest) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/${id}`);
  const resp = await fetch(url, {
    method: "PUT",
    headers: { cookie: cookies().toString() },
    body: JSON.stringify({ ...pack, id }),
  });
  if (!resp.ok) {
    const obj: Record<string, never> | ErrorBody = await resp?.json();
    if (isError(obj)) throw new Error(obj.error);
  }
  return { id };
};

export const deletePack = async (id: string) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/${id}`);
  const resp = await fetch(url, {
    method: "DELETE",
    headers: { cookie: cookies().toString() },
  });
  if (!resp.ok) {
    const obj: Record<string, never> | ErrorBody = await resp?.json();
    if (isError(obj)) throw new Error(obj.error);
  }
  return { id };
};

export const signURL = async (dto: SignURLRequest) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/signURL`);
  url.searchParams.set(FILENAME_QUERY_PARAM, dto.filename);
  url.searchParams.set(PUBLIC_QUERY_PARAM, dto.public.toString());
  const resp = await fetch(url, {
    method: "GET",
    headers: { cookie: cookies().toString() },
  });
  const obj: SignURLResponse | ErrorBody = await resp?.json();
  if (isError(obj)) throw new Error(obj.error);
  return obj;
};
