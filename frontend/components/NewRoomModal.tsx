"use client";

import { useState, FormEvent, useEffect, useRef } from "react";
import Modal from "./Modal";
import { useDebounce } from "use-debounce";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { createRoom, getPacksPreviews, joinRoom } from "@/app/api";
import { isError } from "@/middleware";
import { PackPreview, PrivacyType } from "@/types/pack";
import { CreateRoomRequest } from "@/types/room";
import { IoIosArrowDown } from "react-icons/io";

const inputCls =
  "h-9 w-full px-2.5 rounded-lg border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring transition-[border-color] duration-150";
const selectCls =
  "h-9 w-full pl-2.5 pr-8 rounded-lg border border-border bg-background text-on-background text-sm outline-none focus-ring appearance-none transition-[border-color] duration-150";
const labelCls = "block text-xs font-medium text-on-surface-muted";

export default function NewRoomModal({
  isOpen,
  close,
  fixedPack,
}: {
  isOpen: boolean;
  close: () => void;
  fixedPack?: PackPreview;
}) {
  const router = useRouter();
  const [pack, setPack] = useState<PackPreview>(
    fixedPack ?? { id: "", name: "" }
  );
  const [maxPlayers, setMaxPlayers] = useState(4);
  const [privacyType, setPrivacyType] = useState<PrivacyType>("public");
  const [questionThinkingTime, setQuestionThinkingTime] = useState(10);
  const [answerThinkingTime, setAnswerThinkingTime] = useState(10);
  const [questionThinkingTimeFinal, setQuestionThinkingTimeFinal] =
    useState(60);
  const [readingSymbolsPerSecond, setReadingSymbolsPerSecond] = useState(30);
  const [highlightedIndex, setHighlightedIndex] = useState(-1);
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (fixedPack) setPack(fixedPack);
  }, [fixedPack]);

  const [debouncedPackSearch] = useDebounce(
    fixedPack ? "" : pack.id ? "" : pack.name.trim(),
    250
  );

  const { data: packs = [] } = useQuery<PackPreview[]>({
    queryKey: ["packPreviews", debouncedPackSearch],
    queryFn: async () => {
      const result = await getPacksPreviews(debouncedPackSearch);
      return result.items;
    },
    enabled: debouncedPackSearch.length > 0,
  });

  useEffect(() => {
    setHighlightedIndex(-1);
    setDropdownOpen(packs.length > 0);
  }, [packs]);

  const onPackInputChange = (packFilter: string) => {
    setPack({ id: "", name: packFilter });
    setHighlightedIndex(-1);
  };

  const selectPack = (p: PackPreview) => {
    setPack(p);
    setHighlightedIndex(-1);
    setDropdownOpen(false);
  };

  const onPackKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Escape") {
      e.preventDefault();
      setDropdownOpen(false);
      setHighlightedIndex(-1);
      return;
    }
    if (!packs.length || !dropdownOpen) return;
    if (e.key === "ArrowDown") {
      e.preventDefault();
      const next =
        highlightedIndex < packs.length - 1
          ? highlightedIndex + 1
          : highlightedIndex;
      setHighlightedIndex(next);
      scrollDropdownItem(next);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      const next = highlightedIndex > 0 ? highlightedIndex - 1 : 0;
      setHighlightedIndex(next);
      scrollDropdownItem(next);
    } else if (e.key === "Enter" && highlightedIndex >= 0) {
      e.preventDefault();
      selectPack(packs[highlightedIndex]);
    }
  };

  const scrollDropdownItem = (index: number) => {
    const item = dropdownRef.current?.children[index] as
      | HTMLElement
      | undefined;
    item?.scrollIntoView({ block: "nearest" });
  };

  const onSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const data = Object.fromEntries(new FormData(e.currentTarget).entries());
    const params: CreateRoomRequest = {
      name: data.name as string,
      packId: pack.id,
      options: {
        maxPlayers,
        type: privacyType,
        password: (data.password ?? null) as string | null,
        questionThinkingTime,
        answerThinkingTime,
        questionThinkingTimeFinal,
        readingSymbolsPerSecond,
        falseStartAllowed: data.falseStartAllowed === "on",
      },
    };

    const name = data.name as string;
    if (!name.trim()) {
      setError("Room name is required");
      return;
    }
    if (name.length > 50) {
      setError("Room name must be 50 characters or less");
      return;
    }
    if (!pack.id) {
      setError("Please select a pack from the list");
      return;
    }
    const password = data.password as string | undefined;
    if (privacyType === "private") {
      if (!password) {
        setError("Password is required for private rooms");
        return;
      }
      if (password.length < 4 || password.length > 16) {
        setError("Password must be 4–16 characters");
        return;
      }
    }

    try {
      const { id } = await createRoom(params);
      await joinRoom(id, params.options.password ?? undefined);
      setError(null);
      close();
      const pwd = params.options.password;
      const url = `/rooms/${id}${pwd ? `?password=${pwd}` : ""}`;
      router.push(url);
    } catch (e) {
      setError(isError(e) ? e.error : "Failed to create room");
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={close} closeByClickingOutside>
      <h3 className="text-base font-semibold text-on-surface mb-5">
        Create new room
      </h3>
      <form method="dialog" action="" onSubmit={onSubmit} noValidate>
        <div className="flex flex-col sm:flex-row gap-4">
          {/* Left column */}
          <div className="flex flex-col gap-3 w-52">
            <div className="flex flex-col gap-1.5">
              <span className={labelCls}>Room name</span>
              <input
                className={inputCls}
                type="text"
                placeholder="Name"
                name="name"
              />
            </div>

            <div className="flex flex-col gap-1.5 relative">
              <span className={labelCls}>Pack</span>
              <input
                className={inputCls}
                type="text"
                placeholder="Search packs..."
                value={pack.name}
                onChange={(e) => onPackInputChange(e.target.value)}
                onKeyDown={onPackKeyDown}
                readOnly={!!fixedPack}
              />
              {packs.length > 0 && dropdownOpen && (
                <div
                  ref={dropdownRef}
                  className="absolute z-10 w-full max-h-32 overflow-y-auto rounded-lg top-full mt-0.5 bg-surface-raised border border-border shadow-md"
                >
                  {packs.map((p, index) => (
                    <div
                      key={index}
                      className={`px-3 py-2 cursor-pointer text-sm truncate transition-colors duration-100${
                        index < packs.length - 1
                          ? " border-b border-border"
                          : ""
                      }${
                        index === highlightedIndex
                          ? " bg-primary text-on-primary"
                          : " hover:bg-primary hover:text-on-primary"
                      }`}
                      onClick={() => selectPack(p)}
                    >
                      {p.name}
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="flex flex-col gap-1.5">
              <span className={labelCls}>Privacy</span>
              <div className="relative">
                <select
                  className={selectCls}
                  value={privacyType}
                  onChange={(e) =>
                    setPrivacyType(e.target.value as PrivacyType)
                  }
                >
                  <option value="public">Public</option>
                  <option value="private">Private</option>
                </select>
                <div className="pointer-events-none absolute inset-y-0 right-2.5 flex items-center text-on-surface-muted">
                  <IoIosArrowDown size={14} />
                </div>
              </div>
            </div>

            {privacyType === "private" && (
              <div className="flex flex-col gap-1.5">
                <span className={labelCls}>Password</span>
                <input
                  className={inputCls}
                  type="text"
                  placeholder="Password"
                  name="password"
                />
              </div>
            )}
          </div>

          <div className="hidden sm:block w-px bg-border" />

          {/* Right column */}
          <div className="flex flex-col gap-3 w-52">
            {(
              [
                {
                  label: "Max Players",
                  value: maxPlayers,
                  min: 1,
                  max: 10,
                  onChange: setMaxPlayers,
                  unit: "",
                },
                {
                  label: "Question Thinking Time",
                  value: questionThinkingTime,
                  min: 1,
                  max: 30,
                  onChange: setQuestionThinkingTime,
                  unit: "s",
                },
                {
                  label: "Answer Thinking Time",
                  value: answerThinkingTime,
                  min: 1,
                  max: 30,
                  onChange: setAnswerThinkingTime,
                  unit: "s",
                },
                {
                  label: "Final Round Thinking Time",
                  value: questionThinkingTimeFinal,
                  min: 1,
                  max: 60,
                  onChange: setQuestionThinkingTimeFinal,
                  unit: "s",
                },
                {
                  label: "Reading Speed",
                  value: readingSymbolsPerSecond,
                  min: 10,
                  max: 100,
                  onChange: setReadingSymbolsPerSecond,
                  unit: " ch/s",
                },
              ] as const
            ).map(({ label, value, min, max, onChange, unit }) => (
              <div key={label} className="flex flex-col gap-1">
                <div className="flex items-center justify-between">
                  <span className={labelCls}>{label}</span>
                  <span className="text-xs font-semibold text-on-surface tabular-nums">
                    {value}
                    {unit}
                  </span>
                </div>
                <input
                  type="range"
                  min={min}
                  max={max}
                  value={value}
                  onChange={(e) => onChange(Number(e.target.value))}
                  className="w-full accent-[var(--primary)]"
                />
              </div>
            ))}

            <div className="flex items-center justify-between">
              <span className={labelCls}>False Start Allowed</span>
              <input
                type="checkbox"
                name="falseStartAllowed"
                defaultChecked
                className="w-4 h-4 cursor-pointer accent-[var(--primary)]"
              />
            </div>
          </div>
        </div>

        <div className="mt-5 flex items-center justify-between gap-3">
          {error && <p className="text-xs text-danger">{error}</p>}
          <button
            className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150"
            type="submit"
          >
            Create
          </button>
        </div>
      </form>
    </Modal>
  );
}
