import { headers } from "next/headers";
import Navbar from "../components/Navbar";
import { User, USER_HEADER_NAME } from "../../middleware";

export default function About() {
  const user: User = JSON.parse(headers().get(USER_HEADER_NAME)!);
  return (
    <>
      <Navbar user={user} />
      <div className="About flex justify-center items-center w-full px-10 py-2">
        <div className="mx-2">
          <h1 className="text-2xl font-medium">Contact me:</h1>
        </div>
        <div className="mx-2 text-xs">
          <ul>
            <li>Email</li>
            <li>Telegram</li>
            <li>Discord</li>
          </ul>
        </div>
      </div>
    </>
  );
}
