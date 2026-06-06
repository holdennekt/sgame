import {
  HiddenPack,
  PackPreview,
  SignURLRequest,
  SignURLResponse,
} from "@/types/pack";
import { PackDraft } from "@/types/pack_draft";
import {
  CreateRoomRequest,
  GameHistoryEntry,
  Room,
  RoomLobby,
} from "@/types/room";
import { SearchResponse } from "@/types/search";
import { User } from "@/middleware";

const PAGE_QUERY_PARAM = "page";
const PASSWORD_QUERY_PARAM = "password";
const SEARCH_QUERY_PARAM = "search";

export const login = async (body: {
  login: string;
  password: string;
}): Promise<{ userId: string }> => {
  const resp = await fetch("/api/login", {
    method: "POST",
    body: JSON.stringify(body),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const loginAsGuest = async (
  name: string
): Promise<{ userId: string }> => {
  const resp = await fetch("/api/guest", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const register = async (body: {
  login: string;
  password: string;
}): Promise<{ userId: string }> => {
  const resp = await fetch("/api/register", {
    method: "POST",
    body: JSON.stringify(body),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const logout = async (): Promise<void> => {
  await fetch("/api/logout", { method: "DELETE" });
};

export const getRooms = async (): Promise<RoomLobby[]> => {
  const resp = await fetch("/api/rooms", { cache: "no-store" });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const createRoom = async (
  body: CreateRoomRequest
): Promise<{ id: string }> => {
  const resp = await fetch("/api/rooms", {
    method: "POST",
    body: JSON.stringify(body),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const joinRoom = async (
  id: string,
  password: string | undefined
): Promise<Room> => {
  const url = new URL(`/api/rooms/${id}/join`, window.location.origin);
  if (password) url.searchParams.set(PASSWORD_QUERY_PARAM, password);
  const resp = await fetch(url, { method: "PATCH", cache: "no-store" });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const leaveRoom = async (id: string): Promise<void> => {
  const resp = await fetch(`/api/rooms/${id}/leave`, { method: "PATCH" });
  if (!resp.ok) throw await resp.json();
};

export const getPacks = async (
  packFilter: string,
  page?: number
): Promise<SearchResponse<HiddenPack>> => {
  const url = new URL("/api/packs", window.location.origin);
  url.searchParams.set(SEARCH_QUERY_PARAM, packFilter);
  if (page) url.searchParams.set(PAGE_QUERY_PARAM, page.toString());
  const resp = await fetch(url, { cache: "no-store" });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const getPacksCreatedBy = async (
  createdBy: string,
  packFilter: string,
  page?: number
): Promise<SearchResponse<HiddenPack>> => {
  const url = new URL(`/api/packs/by/${createdBy}`, window.location.origin);
  url.searchParams.set(SEARCH_QUERY_PARAM, packFilter);
  if (page) url.searchParams.set(PAGE_QUERY_PARAM, page.toString());
  const resp = await fetch(url, { cache: "no-store" });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const getPacksPreviews = async (
  packFilter: string
): Promise<SearchResponse<PackPreview>> => {
  const url = new URL("/api/packs/previews", window.location.origin);
  url.searchParams.set(SEARCH_QUERY_PARAM, packFilter);
  const resp = await fetch(url, { cache: "no-store" });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const deletePack = async (id: string): Promise<void> => {
  const resp = await fetch(`/api/packs/${id}`, { method: "DELETE" });
  if (!resp.ok) throw await resp.json();
};

export const getGameHistory = async (
  page?: number
): Promise<SearchResponse<GameHistoryEntry>> => {
  const url = new URL("/api/rooms/history", window.location.origin);
  if (page) url.searchParams.set(PAGE_QUERY_PARAM, page.toString());
  const resp = await fetch(url, { cache: "no-store" });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const getUser = async (id: string): Promise<User> => {
  const resp = await fetch(`/api/users/${id}`, { cache: "no-store" });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const updateUser = async (
  id: string,
  body: { name: string; avatar: string; password?: string }
): Promise<void> => {
  const resp = await fetch(`/api/users/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
  if (!resp.ok) throw await resp.json();
};

export const deleteUser = async (id: string): Promise<void> => {
  const resp = await fetch(`/api/users/${id}`, { method: "DELETE" });
  if (!resp.ok) throw await resp.json();
};

export const signURL = async (
  dto: SignURLRequest
): Promise<SignURLResponse> => {
  const resp = await fetch("/api/packs/signURL", {
    method: "POST",
    body: JSON.stringify(dto),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const createDraft = async (packId?: string): Promise<{ id: string }> => {
  const resp = await fetch("/api/packs/drafts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ from: packId }),
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const getDrafts = async (
  search: string,
  page?: number
): Promise<SearchResponse<PackDraft>> => {
  const url = new URL("/api/packs/drafts/", window.location.origin);
  url.searchParams.set(SEARCH_QUERY_PARAM, search);
  if (page) url.searchParams.set(PAGE_QUERY_PARAM, page.toString());
  const resp = await fetch(url, { cache: "no-store" });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const updateDraft = async (
  id: string,
  body: unknown
): Promise<PackDraft> => {
  const resp = await fetch(`/api/packs/drafts/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!resp.ok) throw await resp.json();
  return await resp.json();
};

export const deleteDraft = async (id: string): Promise<void> => {
  const resp = await fetch(`/api/packs/drafts/${id}`, { method: "DELETE" });
  if (!resp.ok) throw await resp.json();
};

export const publishDraft = async (id: string): Promise<{ id: string }> => {
  const resp = await fetch(`/api/packs/drafts/${id}/publish`, {
    method: "POST",
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};

export const importDraft = async (
  formData: FormData
): Promise<{ id: string }> => {
  const resp = await fetch("/api/packs/drafts/import", {
    method: "POST",
    body: formData,
  });
  if (!resp.ok) throw await resp.json();
  return resp.json();
};
