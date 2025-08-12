import { joinRoom } from "@/app/actions";
import RoomPage, { Room } from "@/app/components/room/Room";
import { USER_HEADER_NAME, User } from "@/middleware";
import { headers } from "next/headers";

export default async function Page({
  params,
  searchParams,
}: {
  params: { id: string };
  searchParams: { [key: string]: string | undefined };
}) {
  const user: User = JSON.parse(headers().get(USER_HEADER_NAME)!);
  const room = await joinRoom(params.id, searchParams.password);
  // const room: Room = {
  //   id: "916a452c-e4b8-42f6-b066-72bd55d04bb3",
  //   name: "qwe",
  //   packPreview: { id: "1", name: "demo" },
  //   host: {
  //     id: "1",
  //     name: "nikita",
  //     avatar: null,
  //     isConnected: true,
  //   },
  //   players: [
  //     {
  //       id: "2",
  //       name: "melnyk",
  //       avatar: null,
  //       score: -1000,
  //       betAmount: null,
  //       isConnected: true,
  //     },
  //     {
  //       id: "3",
  //       name: "manzos",
  //       avatar: null,
  //       score: 300,
  //       betAmount: null,
  //       isConnected: true,
  //     },
  //     {
  //       id: "4",
  //       name: "vova",
  //       avatar: null,
  //       score: 100,
  //       betAmount: null,
  //       isConnected: true,
  //     },
  //   ],
  //   state: "betting",
  //   currentRoundName: "round 1",
  //   currentRoundQuestions: {
  //     КПІ: [
  //       { index: 0, value: 100, hasBeenPlayed: true },
  //       { index: 1, value: 200, hasBeenPlayed: false },
  //       { index: 2, value: 300, hasBeenPlayed: false },
  //       { index: 3, value: 400, hasBeenPlayed: false },
  //       { index: 4, value: 500, hasBeenPlayed: true },
  //     ],
  //   },
  //   currentPlayer: "2",
  //   currentQuestion: {
  //     index: 4,
  //     value: 500,
  //     text: "Скільки разів куля з поляни покидала свій пост?",
  //     attachment: null,
  //     type: "catInBag",
  //     answers: ["ніхто не знає"],
  //     comment: "всі збились з рахунку",
  //     timerStartsAt: null,
  //     timerEndsAt: null,
  //   },
  //   answeringPlayer: null,
  //   allowedToAnswer: [],
  //   finalRoundState: null,
  //   pausedState: { paused: false, pausedAt: null },
  // };

  return <RoomPage user={user} initialRoom={room} />;
}
