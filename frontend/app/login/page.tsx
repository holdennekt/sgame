"use client";

import { toast, ToastContainer } from "react-toastify";
import Navbar from "../../components/Navbar";
import { FormEventHandler, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { login, loginAsGuest } from "../actions";
import { isError } from "@/middleware";

const inputClass = "h-9 w-full px-2.5 rounded-lg border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring transition-[border-color] duration-150";
const btnPrimary = "inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium cursor-pointer bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150";
const btnSecondary = "inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium cursor-pointer bg-surface-raised text-on-surface border border-border hover:bg-border transition-colors duration-150";

export default function Page() {
  const router = useRouter();
  const [guestName, setGuestName] = useState("");

  const onSubmit: FormEventHandler<HTMLFormElement> = async e => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const result = await login({
      login: formData.get("login") as string,
      password: formData.get("password") as string,
    });
    if (isError(result)) {
      toast.error(result.error, { containerId: "login" });
      return;
    }
    router.push("/");
  };

  const onGuestSubmit: FormEventHandler<HTMLFormElement> = async e => {
    e.preventDefault();
    const result = await loginAsGuest(guestName.trim());
    if (isError(result)) {
      toast.error(result.error, { containerId: "login" });
      return;
    }
    router.push("/");
  };

  return (
    <>
      <Navbar />
      <main className="flex-1 flex justify-center items-center px-4">
        <div className="flex flex-col gap-4 w-full max-w-sm">
          <div className="bg-surface border border-border rounded-md shadow p-8">
            <h2 className="text-2xl font-bold mb-1 text-on-background">
              Welcome back
            </h2>
            <p className="text-sm mb-6 text-on-surface-muted">Sign in to your account</p>

            <form onSubmit={onSubmit} className="flex flex-col gap-4">
              <label className="flex flex-col gap-1">
                <span className="text-sm font-medium text-on-surface">
                  Login
                </span>
                <input
                  className={inputClass}
                  type="text"
                  placeholder="Your login"
                  name="login"
                  minLength={1}
                  maxLength={50}
                  required
                />
              </label>

              <label className="flex flex-col gap-1">
                <span className="text-sm font-medium text-on-surface">
                  Password
                </span>
                <input
                  className={inputClass}
                  type="password"
                  placeholder="Your password"
                  name="password"
                  minLength={1}
                  maxLength={50}
                  required
                />
              </label>

              <div className="flex justify-between items-center mt-2">
                <Link className={btnSecondary} href="/register">
                  Create account
                </Link>
                <button type="submit" className={btnPrimary}>
                  Sign in
                </button>
              </div>
            </form>
          </div>

          <div className="bg-surface border border-border rounded-md shadow p-6">
            <p className="text-sm font-medium text-on-surface mb-3">Continue as guest</p>
            <form onSubmit={onGuestSubmit} className="flex gap-2">
              <input
                className={inputClass}
                type="text"
                placeholder="Your name"
                value={guestName}
                onChange={e => setGuestName(e.target.value)}
                minLength={1}
                maxLength={50}
                required
              />
              <button type="submit" className={`${btnSecondary} shrink-0`}>
                Play
              </button>
            </form>
          </div>
        </div>
      </main>

      <ToastContainer containerId="login" position="bottom-left" theme="colored" />
    </>
  );
}
