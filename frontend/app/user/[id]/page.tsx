import { getUser } from "@/app/server-fetch";
import Navbar from "@/components/Navbar";
import { USER_HEADER_NAME, User } from "@/middleware";
import { headers } from "next/headers";
import ProfilePage from "./ProfilePage";

export default async function Page({ params }: { params: { id: string } }) {
  const user: User = JSON.parse(
    decodeURIComponent(headers().get(USER_HEADER_NAME)!)
  );
  const isOwn = user.id === params.id;
  const profileUser = isOwn ? user : await getUser(params.id);

  return (
    <>
      <Navbar />
      <main className="flex-1 min-w-0 min-h-0 overflow-y-auto">
        <ProfilePage user={profileUser} isOwn={isOwn} />
      </main>
    </>
  );
}
