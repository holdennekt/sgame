"use client";

import { ReactNode, useState } from "react";
import { IoIosArrowDown, IoIosArrowForward } from "react-icons/io";

export default function Accordion({
  title,
  children,
}: {
  title: ReactNode;
  children?: ReactNode;
}) {
  const [isActive, setIsActive] = useState(false);

  return (
    <div className="w-full">
      <div
        className={`flex p-2 items-center cursor-pointer hover:border hover:rounded${
          isActive ? " border rounded" : ""
        }`}
        onClick={() => setIsActive(!isActive)}
      >
        <div>{isActive ? <IoIosArrowDown size="24" /> : <IoIosArrowForward size="24" />}</div>
        <p className="ml-2">{title}</p>
      </div>
      {isActive && children}
    </div>
  );
}
