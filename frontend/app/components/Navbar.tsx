import Link from "next/link";
import { User } from "../../middleware";

export default function Navbar({
  user,
  openNewTab = false,
}: {
  user?: User;
  openNewTab?: boolean;
}) {
  const links = [
    {
      text: user?.name,
      path: `/user/${user?.id}`,
    },
    {
      text: "Packs",
      path: "/packs",
    },
    {
      text: "About",
      path: "/about",
    },
  ];

  const linksTarget = openNewTab ? "_blank" : "_self";

  return (
    <nav className="nav flex justify-between items-center px-10 py-2">
      <div>
        <h1 className="text-4xl font-medium">
          <Link href="/" target={linksTarget}>
            SGame
          </Link>
        </h1>
      </div>

      <ul className="flex">
        {links.map(({ text, path }, i) =>
          text && (
            <li
              key={i}
              className="nav-links px-2 cursor-pointer font-medium hover:text-white duration-200"
            >
              <Link href={path} target={linksTarget}>
                {text}
              </Link>
            </li>
          )
        )}
      </ul>
    </nav>
  );
}
