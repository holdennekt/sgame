"use client";

import { toast, ToastContainer } from "react-toastify";
import Navbar from "../../components/Navbar";
import { FormEventHandler } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { login } from "../actions";

export default function Page() {
  const router = useRouter();
  const onSubmit: FormEventHandler<HTMLFormElement> = async e => {
    e.preventDefault();

    const formData = new FormData(e.currentTarget);
    try {
      await login({
        login: formData.get("login") as string,
        password: formData.get("password") as string,
      });
      router.push("/");
    } catch (error) {
      if (error instanceof Error) {
        toast.error(error.message, { containerId: "login" });
      }
    }
  };

  return (
    <>
      <Navbar />
      <main className="flex-1 flex justify-center items-center">
        <div className="rounded-xl p-5 surface">
          <form action="" onSubmit={onSubmit}>
            <label className="block">
              <p className="text-sm font-medium">Login</p>
              <input
                className="w-full h-8 rounded-lg mt-1 p-1 text-black"
                type="text"
                placeholder="Login"
                name="login"
                minLength={1}
                maxLength={50}
                required
              />
            </label>
            <label className="block mt-2">
              <p className="text-sm font-medium">Password</p>
              <input
                className="w-full h-8 rounded-lg mt-1 p-1 text-black"
                type="password"
                placeholder="Password"
                name="password"
                minLength={1}
                maxLength={50}
                required
              />
            </label>
            <div className="flex justify-between mt-3">
              <Link
                className="rounded px-3 py-1 font-medium secondary"
                href={"/register"}
              >
                Sign up
              </Link>
              <button className="rounded px-3 py-1 font-medium primary">
                Sign in
              </button>
            </div>
          </form>
        </div>
      </main>
      <ToastContainer
        containerId="login"
        position="bottom-left"
        theme="colored"
      />
    </>
  );
}
