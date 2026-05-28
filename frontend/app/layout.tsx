import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { ReactNode } from "react";
import "./globals.css";
import "react-toastify/dist/ReactToastify.css";
import { headers } from "next/headers";
import { USER_HEADER_NAME, User } from "@/middleware";
import { UserProvider } from "@/contexts/UserContext";
import QueryProvider from "@/components/QueryProvider";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "SGame",
  description: "jeopardy-like quiz game",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: ReactNode;
}>) {
  const userHeader = headers().get(USER_HEADER_NAME);
  const user: User | null = userHeader ? JSON.parse(decodeURIComponent(userHeader)) : null;

  return (
    <html lang="en">
      <body className={`${inter.className} background h-svh flex flex-col`}>
        <QueryProvider>
          <UserProvider user={user}>
            {children}
          </UserProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
