"use client";

import { useState, FormEvent, useEffect } from "react";
import Modal from "../Modal";
import { useDebouncedCallback } from "use-debounce";
import { useRouter } from "next/navigation";
import { toast } from "react-toastify";
import { createRoom, getPacksPreviews } from "@/app/actions";

export type PackPreview = { id: string; name: string };
export type CreateRoomParams = {
  name: string;
  packId: string;
  options: {
    maxPlayers: number;
    type: string;
    password?: string;
    questionThinkingTime: number;
    answerThinkingTime: number;
    questionThinkingTimeFinal: number;
    falseStartAllowed: boolean;
  };
};

export default function NewRoomModal({
  isOpen,
  close,
  fixedPack,
  toastContainerId,
}: {
  isOpen: boolean;
  close: () => void;
  fixedPack?: PackPreview;
  toastContainerId: string;
}) {
  const router = useRouter();
  const [pack, setPack] = useState<PackPreview>(
    fixedPack ?? { id: "", name: "" }
  );
  const [packs, setPacks] = useState<PackPreview[]>([]);
  const [maxPlayers, setMaxPlayers] = useState(4);
  const [privacyType, setPrivacyType] = useState("public");
  const [questionThinkingTime, setQuestionThinkingTime] = useState(10);
  const [answerThinkingTime, setAnswerThinkingTime] = useState(10);
  const [questionThinkingTimeFinal, setQuestionThinkingTimeFinal] = useState(60);

  useEffect(() => {
    setPack(fixedPack ?? { id: "", name: "" });
  }, [fixedPack]);

  const fetchPacks = useDebouncedCallback(async (packFilter: string) => {
    if (!packFilter) return;
    try {
      const packs = await getPacksPreviews(packFilter);
      setPacks(packs);
    } catch (error) {
      if (error instanceof Error)
        toast.error(error.message, { containerId: toastContainerId });
    }
  }, 500);

  const onPackInputChange = (packFilter: string) => {
    setPack({ id: "", name: packFilter });
    fetchPacks(packFilter.trim())?.catch(console.log);
  };

  const onSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const data = Object.fromEntries(new FormData(e.currentTarget).entries());
    const params: CreateRoomParams = {
      name: data.name as string,
      packId: pack.id,
      options: {
        maxPlayers,
        type: privacyType,
        password: data.password as string | undefined,
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
      <h3 className="text-base/7 font-medium">Create new room</h3>
      <form method="dialog" action="" onSubmit={onSubmit}>
        <div className="flex flex-col sm:flex-row gap-2 mt-2">
          <div className="w-48 flex flex-col gap-2 flex-1">
            <label>
              <p className="text-sm font-medium">Name</p>
              <input
                className="w-full h-8 rounded-lg mt-1 p-1 text-black"
                type="text"
                placeholder="Name"
                name="name"
                minLength={1}
                maxLength={50}
                required
              />
            </label>
            <label className="relative">
              <p className="text-sm font-medium">Pack</p>
              <input
                className={`w-full h-8 ${
                  packs.length > 0 ? "" : "rounded-lg "
                }mt-1 p-1 text-black`}
                type="text"
                placeholder="Name"
                value={pack.name}
                onChange={e => onPackInputChange(e.target.value)}
                required
                readOnly={!!fixedPack}
              />
              {packs.length > 0 && (
                <div className="absolute w-full max-h-32 overflow-y-auto bg-white">
                  <table className="w-full">
                    <tbody className="w-full">
                      {packs.map((pack, index) => (
                        <tr
                          className="w-full"
                          onClick={() => {
                            setPack(pack);
                            setPacks([]);
                          }}
                          key={index}
                        >
                          <td className="w-full h-8 px-1 border-y border-black">
                            <p className="text-black leading-none truncate">
                              {pack.name}
                            </p>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </label>
            <div>
              <p className="text-sm font-medium">Privacy Type</p>
              <select
                className="w-full h-8 mt-1 p-0.5 rounded-md text-black"
                value={privacyType}
                onChange={e => setPrivacyType(e.target.value)}
              >
                <option value="public">Public</option>
                <option value="private">Private</option>
              </select>
            </div>
            {privacyType === "private" && (
              <div>
                <p className="text-sm font-medium">Password</p>
                <input
                  className="w-full rounded-lg mt-1 p-1 text-black"
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
          <div className="w-48 flex flex-col gap-2 flex-1">
            <label>
              <p className="text-sm font-medium">Max Players</p>
              <p className="text-center text-sm font-semibold">{maxPlayers}</p>
              <input
                className="w-full"
                type="range"
                min="1"
                max="10"
                value={maxPlayers}
                onChange={e => setMaxPlayers(Number(e.target.value))}
              />
            </label>
            <label>
              <p className="text-sm font-medium">Question Thinking Time</p>
              <p className="text-center text-sm font-semibold">
                {questionThinkingTime}
              </p>
              <input
                className="w-full"
                type="range"
                min="1"
                max="30"
                value={questionThinkingTime}
                onChange={e => setQuestionThinkingTime(Number(e.target.value))}
              />
            </label>
            <label>
              <p className="text-sm font-medium">Answer Thinking Time</p>
              <p className="text-center text-sm font-semibold">
                {answerThinkingTime}
              </p>
              <input
                className="w-full"
                type="range"
                min="1"
                max="30"
                value={answerThinkingTime}
                onChange={e => setAnswerThinkingTime(Number(e.target.value))}
              />
            </label>
            <label>
              <p className="text-sm font-medium">Question Thinking Time Final</p>
              <p className="text-center text-sm font-semibold">
                {questionThinkingTimeFinal}
              </p>
              <input
                className="w-full"
                type="range"
                min="1"
                max="60"
                value={questionThinkingTimeFinal}
                onChange={e => setQuestionThinkingTimeFinal(Number(e.target.value))}
              />
            </label>
            <label>
              <p className="text-sm font-medium">False Start Allowed</p>
              <input
                className="w-full h-4"
                type="checkbox"
                name="falseStartAllowed"
                defaultChecked
              />
            </label>
          </div>
        </div>
        <div className="mt-4 flex flex-row-reverse">
          <button
            className="rounded-lg py-1.5 px-3 text-base font-medium primary"
            type="submit"
          >
            Create
          </button>
        </div>
      </form>
    </Modal>
  );
}
