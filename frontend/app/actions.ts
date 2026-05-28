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
import {
  CreateRoomRequest,
  GameHistoryEntry,
  Room,
  RoomLobby,
} from "@/types/room";
import { SearchResponse } from "@/types/search";
import { User } from "@/middleware";

export type ActionResult<T> = T | { error: string };

const PAGE_QUERY_PARAM = "page";
const PASSWORD_QUERY_PARAM = "password";
const SEARCH_QUERY_PARAM = "search";
const FILENAME_QUERY_PARAM = "filename";
const PUBLIC_QUERY_PARAM = "public";

const passCookies = (resp: Response) => {
  const setCookie = resp.headers.get("set-cookie");
  if (setCookie) {
    parse(splitCookiesString(setCookie)).forEach((cookie) => {
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

export const login = async (body: {
  login: string;
  password: string;
}): Promise<ActionResult<string>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/login`);
  const resp = await fetch(url, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(body),
  });
  passCookies(resp);
  return await resp?.json();
};

export const loginAsGuest = async (name: string): Promise<ActionResult<string>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/guest`);
  const resp = await fetch(url, {
    method: "POST",
    body: JSON.stringify({ name }),
  });
  passCookies(resp);
  return await resp?.json();
};

export const register = async (body: {
  login: string;
  password: string;
}): Promise<ActionResult<string>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/register`);
  const resp = await fetch(url, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(body),
  });
  passCookies(resp);
  return await resp?.json();
};

export const getRooms = async (): Promise<ActionResult<RoomLobby[]>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/rooms`);
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  return await resp?.json();
};

export const createRoom = async (
  body: CreateRoomRequest,
): Promise<ActionResult<{ id: string }>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/rooms`);
  const resp = await fetch(url, {
    method: "POST",
    body: JSON.stringify(body),
    headers: { cookie: cookies().toString() },
  });
  return await resp?.json();
};

export const joinRoom = async (
  id: string,
  password: string | undefined,
): Promise<ActionResult<Room>> => {
  const url = new URL(
    `http://${process.env.BACKEND_HOST}/api/rooms/${id}/join`,
  );
  if (password) url.searchParams.set(PASSWORD_QUERY_PARAM, password);
  const resp = await fetch(url, {
    method: "PATCH",
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  }).catch(console.log);
  return await resp?.json();
};

export const leaveRoom = async (id: string): Promise<ActionResult<void>> => {
  const url = new URL(
    `http://${process.env.BACKEND_HOST}/api/rooms/${id}/leave`,
  );
  const resp = await fetch(url, {
    method: "PATCH",
    headers: { cookie: cookies().toString() },
  });
  if (!resp.ok) return await resp?.json();
};

export const getPacks = async (
  packFilter: string,
  page?: number,
): Promise<ActionResult<SearchResponse<HiddenPack>>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs`);
  url.searchParams.set(SEARCH_QUERY_PARAM, packFilter);
  if (page) url.searchParams.set(PAGE_QUERY_PARAM, page.toString());
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  return await resp?.json();
};

export const getPacksCreatedBy = async (
  createdBy: string,
  packFilter: string,
  page?: number,
): Promise<ActionResult<SearchResponse<HiddenPack>>> => {
  const url = new URL(
    `http://${process.env.BACKEND_HOST}/api/packs/by/${createdBy}`,
  );
  url.searchParams.set(SEARCH_QUERY_PARAM, packFilter);
  if (page) url.searchParams.set(PAGE_QUERY_PARAM, page.toString());
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  return await resp?.json();
};

export const getPacksPreviews = async (
  packFilter: string,
): Promise<ActionResult<SearchResponse<PackPreview>>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/previews`);
  url.searchParams.set(SEARCH_QUERY_PARAM, packFilter);
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  return await resp?.json();
};

export const getPack = async (id: string): Promise<ActionResult<Pack>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/${id}`);
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  return await resp?.json();
};

export const createPack = async (
  pack: CreatePackRequest,
): Promise<ActionResult<{ id: string }>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs`);
  const resp = await fetch(url, {
    method: "POST",
    headers: { cookie: cookies().toString() },
    body: JSON.stringify(pack),
  });
  return await resp?.json();
};

export const updatePack = async (
  id: string,
  pack: CreatePackRequest,
): Promise<ActionResult<{ id: string }>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/${id}`);
  const resp = await fetch(url, {
    method: "PUT",
    headers: { cookie: cookies().toString() },
    body: JSON.stringify({ ...pack, id }),
  });
  if (!resp.ok) return await resp?.json();
  return { id };
};

export const deletePack = async (
  id: string,
): Promise<ActionResult<{ id: string }>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/${id}`);
  const resp = await fetch(url, {
    method: "DELETE",
    headers: { cookie: cookies().toString() },
  });
  if (!resp.ok) return await resp?.json();
  return { id };
};

export const getGameHistory = async (
  page?: number,
): Promise<ActionResult<SearchResponse<GameHistoryEntry>>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/rooms/history`);
  if (page) url.searchParams.set(PAGE_QUERY_PARAM, page.toString());
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  return await resp?.json();
};

export const getUser = async (id: string): Promise<ActionResult<User>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/users/${id}`);
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  return await resp?.json();
};

export const updateUser = async (
  id: string,
  body: { name: string; avatar: string; password?: string },
): Promise<ActionResult<User>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/users/${id}`);
  const resp = await fetch(url, {
    method: "PUT",
    headers: { cookie: cookies().toString() },
    body: JSON.stringify(body),
  });
  return await resp?.json();
};

export const logout = async (): Promise<void> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/logout`);
  await fetch(url, {
    method: "DELETE",
    headers: { cookie: cookies().toString() },
  });
  cookies().delete("sessionId");
};

export const deleteUser = async (id: string): Promise<ActionResult<void>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/users/${id}`);
  const resp = await fetch(url, {
    method: "DELETE",
    headers: { cookie: cookies().toString() },
  });
  if (!resp.ok) return await resp?.json();
};

export const signURL = async (
  dto: SignURLRequest,
): Promise<ActionResult<SignURLResponse>> => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/api/packs/signURL`);
  const resp = await fetch(url, {
    method: "POST",
    headers: { cookie: cookies().toString() },
    body: JSON.stringify(dto),
  });
  return await resp?.json();
};
