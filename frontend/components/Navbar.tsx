"use client";

import Link from "next/link";
import { useUser } from "@/contexts/UserContext";

export default function Navbar({
  openNewTab = false,
}: {
  openNewTab?: boolean;
}) {
  const user = useUser();
  const target = openNewTab ? "_blank" : "_self";

  return (
    <nav className="bg-surface border-b border-border shadow-sm flex items-center justify-between px-6 py-3 min-h-14">
      <Link
        href="/lobby"
        target={target}
        className="text-xl font-bold tracking-tight text-on-background"
      >
        SGame
      </Link>

      <ul className="flex items-center gap-1">
        {user?.name && (
          <li>
            <Link
              href={`/user/${user.id}`}
              target={target}
              className="text-primary font-semibold text-sm px-2.5 py-1 rounded-md hover:text-on-surface hover:bg-surface-raised transition-[color,background] duration-150"
            >
              {user.name}
            </Link>
          </li>
        )}
        <li>
          <Link
            href="/packs"
            target={target}
            className="text-on-surface-muted text-sm font-medium px-2.5 py-1 rounded-md hover:text-on-surface hover:bg-surface-raised transition-[color,background] duration-150"
          >
            Packs
          </Link>
        </li>
        <li>
          <Link
            href="/about"
            target={target}
            className="text-on-surface-muted text-sm font-medium px-2.5 py-1 rounded-md hover:text-on-surface hover:bg-surface-raised transition-[color,background] duration-150"
          >
            About
          </Link>
        </li>
      </ul>
    </nav>
  );
}
