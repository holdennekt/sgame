"use client";

import { useState, FormEvent, useEffect } from "react";
import Modal from "./Modal";
import { useDebounce } from "use-debounce";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { createRoom, getPacksPreviews } from "@/app/actions";
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
    fixedPack ?? { id: "", name: "" },
  );
  const [maxPlayers, setMaxPlayers] = useState(4);
  const [privacyType, setPrivacyType] = useState<PrivacyType>("public");
  const [questionThinkingTime, setQuestionThinkingTime] = useState(10);
  const [answerThinkingTime, setAnswerThinkingTime] = useState(10);
  const [questionThinkingTimeFinal, setQuestionThinkingTimeFinal] = useState(60);

  useEffect(() => {
    if (fixedPack) setPack(fixedPack);
  }, [fixedPack]);

  const [debouncedPackSearch] = useDebounce(
    fixedPack ? "" : pack.id ? "" : pack.name.trim(),
    400,
  );

  const { data: packs = [] } = useQuery({
    queryKey: ["packPreviews", debouncedPackSearch],
    queryFn: () => getPacksPreviews(debouncedPackSearch),
    enabled: debouncedPackSearch.length > 0,
  });

  const onPackInputChange = (packFilter: string) => {
    setPack({ id: "", name: packFilter });
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
        falseStartAllowed: data.falseStartAllowed === "on",
      },
    };

    close();
    const id = await createRoom(params);

    const pwd = params.options.password;
    const url = `/rooms/${id}${pwd ? `?password=${pwd}` : ""}`;
    router.push(url);
  };

  return (
    <Modal isOpen={isOpen} onClose={close} closeByClickingOutside>
      <h3 className="text-base font-semibold text-on-surface mb-5">
        Create new room
      </h3>
      <form method="dialog" action="" onSubmit={onSubmit}>
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
                minLength={1}
                maxLength={50}
                required
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
                required
                readOnly={!!fixedPack}
              />
              {packs.length > 0 && (
                <div className="absolute z-10 w-full max-h-32 overflow-y-auto rounded-lg top-full mt-0.5 bg-surface-raised border border-border shadow-md">
                  {packs.map((pack, index) => (
                    <div
                      key={index}
                      className={`px-3 py-2 cursor-pointer text-sm truncate hover:bg-primary hover:text-on-primary transition-colors duration-100${index < packs.length - 1 ? " border-b border-border" : ""}`}
                      onClick={() => {
                        setPack(pack);
                      }}
                    >
                      {pack.name}
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
                  minLength={4}
                  maxLength={16}
                  name="password"
                  required
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
              ] as const
            ).map(({ label, value, min, max, onChange, unit }) => (
              <div key={label} className="flex flex-col gap-1">
                <div className="flex items-center justify-between">
                  <span className={labelCls}>{label}</span>
                  <span className="text-xs font-semibold text-on-surface tabular-nums">
                    {value}{unit}
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

        <div className="mt-5 flex justify-end">
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
