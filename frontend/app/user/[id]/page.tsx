import { getUser } from "@/app/actions";
import Navbar from "@/components/Navbar";
import { USER_HEADER_NAME, User, isError } from "@/middleware";
import { headers } from "next/headers";
import ProfilePage from "./ProfilePage";

export default async function Page({ params }: { params: { id: string } }) {
  const user: User = JSON.parse(headers().get(USER_HEADER_NAME)!);
  const isOwn = user.id === params.id;
  const profileUser = await getUser(params.id);
  if (isError(profileUser)) throw new Error(profileUser.error);

  return (
    <>
      <Navbar />
      <main className="flex-1 min-w-0 min-h-0 overflow-y-auto">
        <ProfilePage user={profileUser} isOwn={isOwn} />
      </main>
    </>
  );
}
