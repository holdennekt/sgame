import { NextRequest, NextResponse } from "next/server";

export const SESSION_ID_COOKIE_NAME = "sessionId";
export const USER_HEADER_NAME = "user";

export type User = { id: string; name: string; avatar: string | null };

export type ErrorBody = { error: string };
export const isError = (obj: unknown): obj is ErrorBody =>
  (obj as ErrorBody).error !== undefined;

const unprotectedPages = ["/register", "/login", "/about"];

export const config = {
  matcher: ["/((?!transport|_next/static|_next/image|favicon.ico).*)"],
};

export async function middleware(request: NextRequest) {
  const sessionId = request.cookies.get(SESSION_ID_COOKIE_NAME)?.value;

  const path = request.nextUrl.pathname;

  if (unprotectedPages.some((prefix) => path.startsWith(prefix))) {
    if (
      sessionId &&
      (path.startsWith("/login") || path.startsWith("/register"))
    )
      return Response.redirect(new URL("/", request.url));
    return NextResponse.next();
  }

  if (!sessionId) return Response.redirect(new URL("/login", request.url));

  const resp = await fetch(`http://${process.env.BACKEND_HOST}/api/user`, {
    headers: { cookie: request.cookies.toString() },
  });
  const user: User | ErrorBody = await resp?.json();

  if (!resp.ok || isError(user)) {
    const response = NextResponse.redirect(new URL("/login", request.url));
    response.cookies.delete(SESSION_ID_COOKIE_NAME);
    return response;
  }
  // const user = {"id":"3","name":"nikita","avatar":null,"isConnected":true}

  const clonedRequest = request.clone();
  clonedRequest.headers.set(USER_HEADER_NAME, JSON.stringify(user));
  return NextResponse.rewrite(request.url.toString(), {
    request: clonedRequest,
  });
}
