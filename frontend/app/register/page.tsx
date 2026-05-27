"use client";

import { toast, ToastContainer } from "react-toastify";
import Navbar from "../../components/Navbar";
import { FormEventHandler } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { register } from "../actions";

const inputClass = "h-9 w-full px-2.5 rounded-lg border border-border bg-background text-on-background text-sm outline-none placeholder:text-on-surface-muted focus-ring transition-[border-color] duration-150";
const btnPrimary = "inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium cursor-pointer bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150";
const btnSecondary = "inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium cursor-pointer bg-surface-raised text-on-surface border border-border hover:bg-border transition-colors duration-150";

export default function Page() {
  const router = useRouter();

  const onSubmit: FormEventHandler<HTMLFormElement> = async e => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    try {
      await register({
        login: formData.get("login") as string,
        password: formData.get("password") as string,
      });
      router.push("/");
    } catch (error) {
      if (error instanceof Error) {
        toast.error(error.message, { containerId: "register" });
      }
    }
  };

  return (
    <>
      <Navbar />
      <main className="flex-1 flex justify-center items-center px-4">
        <div className="bg-surface border border-border rounded-md shadow p-8 w-full max-w-sm">
          <h2 className="text-2xl font-bold mb-1 text-on-background">
            Create account
          </h2>
          <p className="text-sm mb-6 text-on-surface-muted">Join SGame and start playing</p>

          <form onSubmit={onSubmit} className="flex flex-col gap-4">
            <label className="flex flex-col gap-1">
              <span className="text-sm font-medium text-on-surface">
                Login
              </span>
              <input
                className={inputClass}
                type="text"
                placeholder="Choose a login"
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
                placeholder="Choose a password"
                name="password"
                minLength={1}
                maxLength={50}
                required
              />
            </label>

            <div className="flex justify-between items-center mt-2">
              <Link className={btnSecondary} href="/login">
                Sign in instead
              </Link>
              <button type="submit" className={btnPrimary}>
                Create account
              </button>
            </div>
          </form>
        </div>
      </main>

      <ToastContainer containerId="register" position="bottom-left" theme="colored" />
    </>
  );
}
