"use client";

import React, { useState } from "react";
import { useDebouncedCallback } from "use-debounce";
import { ToastContainer } from "react-toastify";
import { getPacks } from "../actions";
import NewRoomModal, { PackPreview } from "./lobby/NewRoomModal";
import { User } from "@/middleware";
import Private from "@/public/private.png";
import Public from "@/public/public.png";
import Image from "next/image";
import Accordion from "./Accordion";
import AddButton from "./AddButton";
import { useRouter } from "next/navigation";
import Link from "next/link";

export type PacksResp = {
  items: HiddenPack[];
  total: number;
  page: number;
  pageSize: number;
  hasNext: boolean;
};
export type HiddenPack = {
  id: string;
  createdBy: User;
  name: string;
  type: "public" | "private";
  rounds: HiddenRound[];
  finalRound: HiddenFinalRound;
};
export type HiddenRound = {
  name: string;
  categories: HiddenCategory[];
};
export type HiddenCategory = {
  name: string;
};
type HiddenFinalRound = {
  categories: HiddenCategory[];
};

const MAX_VISIBLE_PAGES = 5;

export default function PacksList({
  user,
  initialPacks,
}: {
  user: User;
  initialPacks: PacksResp;
}) {
  const router = useRouter();
  const [packsFilter, setPacksFilter] = useState("");
  const [pagesCount, setPagesCount] = useState(
    Math.floor(initialPacks.total / initialPacks.pageSize) + 1
  );
  const [currentPage, setCurrentPage] = useState(1);
  const [packs, setPacks] = useState(initialPacks.items);
  const [newRoomModal, setNewRoomModal] = useState<{
    isOpen: boolean;
    pack?: PackPreview;
  }>({ isOpen: false });

  const fetchPacks = useDebouncedCallback(async (packFilter: string) => {
    const packs = await getPacks(packFilter);
    setPacks(packs.items);
    setCurrentPage(1);
    setPagesCount(Math.ceil(packs.total / packs.pageSize));
  }, 500);

  const onPacksFilterChange = (packsFilter: string) => {
    setPacksFilter(packsFilter);
    fetchPacks(packsFilter.trim())?.catch(console.log);
  };

  const selectPage = async (page: number) => {
    if (page === currentPage) return;
    const packs = await getPacks(packsFilter, page);
    setPacks(packs.items);
  };

  const selectPrevPage = async () => {
    if (currentPage === 1) return;
    const packs = await getPacks(packsFilter, currentPage - 1);
    setCurrentPage(currentPage - 1);
    setPacks(packs.items);
  };

  const selectNextPage = async () => {
    if (currentPage === pagesCount) return;
    const packs = await getPacks(packsFilter, currentPage + 1);
    setCurrentPage(currentPage + 1);
    setPacks(packs.items);
  };

  return (
    <>
      <main className="flex-1 min-w-0 min-h-0 p-2">
        <div className="flex flex-col h-full rounded relative surface p-4">
          <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-2">
            <div className="flex items-center">
              <p className="w-fit text-xl font-semibold leading-none">
                Search for packs:
              </p>
              <input
                className="rounded text-black ml-2 p-1"
                type="text"
                placeholder="Name"
                value={packsFilter}
                onChange={(e) => onPacksFilterChange(e.target.value)}
              />
            </div>
            <ul className="flex h-fit">
              <li
                className="border cursor-pointer rounded-l px-2 py-1"
                onClick={selectPrevPage}
              >
                prev
              </li>
              {[...Array(Math.min(pagesCount, MAX_VISIBLE_PAGES)).keys()].map(
                (val) => (
                  <li
                    className={`border cursor-pointer px-2 py-1${
                      val + 1 === currentPage ? " bg-white text-black" : ""
                    }`}
                    onClick={() => selectPage(val + 1)}
                    key={val}
                  >
                    {val + 1}
                  </li>
                )
              )}
              {pagesCount > MAX_VISIBLE_PAGES && (
                <li className="border px-2 py-1">...</li>
              )}
              <li
                className="border cursor-pointer rounded-r px-2 py-1"
                onClick={selectNextPage}
              >
                next
              </li>
            </ul>
          </div>
          <div className="flex-1 flex flex-col gap-2 overflow-y-auto mt-2">
            {packs.length ? (
              packs.map((pack, index) => (
                <div
                  className="w-full flex gap-2 rounded border p-2"
                  key={index}
                >
                  <div className="flex-1">
                    <div className="flex items-center">
                      {user.id === pack.createdBy.id ? (
                        <Link
                          className="w-fit text-lg underline leading-none font-semibold truncate"
                          href={`/packs/${pack.id}`}
                          title={pack.name}
                        >
                          {pack.name}
                        </Link>
                      ) : (
                        <p
                          className="w-fit text-lg leading-none font-semibold truncate"
                          title={pack.name}
                        >
                          {pack.name}
                        </p>
                      )}
                      <Image
                        className="w-5 h-5 ml-1"
                        src={pack.type === "public" ? Public : Private}
                        alt={pack.type}
                      />
                    </div>
                    <p className="mt-1 text-sm font-normal">
                      Author: <span className="italic">{pack.createdBy.name}</span>
                    </p>
                    <div className="mt-1">
                      {pack.rounds.map((round, index) => (
                        <Accordion
                          title={round.name}
                          key={index}
                        >
                          <ul className="px-5 list-inside list-disc">
                            {round.categories.map((category, index) => (
                              <li key={index}>{category.name}</li>
                            ))}
                          </ul>
                        </Accordion>
                      ))}
                      <Accordion title="Final round">
                        <ul className="px-5 list-inside list-disc">
                          {pack.finalRound.categories.map((category, index) => (
                            <li key={index}>{category.name}</li>
                          ))}
                        </ul>
                      </Accordion>
                    </div>
                  </div>
                  <div className="flex flex-col justify-center">
                    <button
                      className="px-2 py-1 rounded primary"
                      onClick={() =>
                        setNewRoomModal({
                          isOpen: true,
                          pack: { id: pack.id, name: pack.name },
                        })
                      }
                    >
                      Play
                    </button>
                  </div>
                </div>
              ))
            ) : (
              <div className="w-full h-full flex justify-center items-center">
                <p className="text-2xl font-bold">No packages</p>
              </div>
            )}
          </div>
          <AddButton onClick={() => router.push("/packs/new")} />
        </div>
      </main>
      <NewRoomModal
        isOpen={newRoomModal.isOpen}
        close={() => setNewRoomModal({ isOpen: false })}
        fixedPack={newRoomModal.pack}
        toastContainerId="packs"
      />
      <ToastContainer
        containerId="packs"
        position="bottom-left"
        theme="colored"
      />
    </>
  );
}
