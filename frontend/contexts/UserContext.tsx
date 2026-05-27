"use client";

import { createContext, useContext } from "react";
import React from "react";
import { User } from "@/middleware";

const UserContext = createContext<User | null>(null);

export function UserProvider({
  children,
  user,
}: {
  children: React.ReactNode;
  user: User | null;
}) {
  return <UserContext.Provider value={user}>{children}</UserContext.Provider>;
}

export function useUser(): User | null {
  return useContext(UserContext);
}

export function useRequiredUser(): User {
  const user = useContext(UserContext);
  if (!user) throw new Error("useRequiredUser must be used in an authenticated context");
  return user;
}
