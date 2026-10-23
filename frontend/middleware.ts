import { NextRequest, NextResponse } from "next/server";

export const SESSION_ID_COOKIE_NAME = "sessionId";
export const USER_HEADER_NAME = "user";

export type User = {
  id: string;
  name: string;
  avatar: string | null;
  isGuest?: boolean;
  isDeleted?: boolean;
};

export type ErrorBody = { error: string; status?: number };
export const isError = (obj: unknown): obj is ErrorBody =>
  obj != null && (obj as ErrorBody).error !== undefined;

export const isConflictError = (obj: unknown): obj is ErrorBody =>
  isError(obj) && obj.status === 409;

export type ValidationError = { path: string; message: string };
export type ValidationErrors = { errors: ValidationError[] };
export const isValidationErrors = (obj: unknown): obj is ValidationErrors =>
  obj != null && Array.isArray((obj as ValidationErrors).errors);

const unprotectedPages = ["/register", "/login", "/about"];

export const config = {
  matcher: [
    "/((?!api|_next/static|_next/image|favicon.ico|transport|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)",
  ],
};

const fetchUser = (request: NextRequest) =>
  fetch(`http://${process.env.BACKEND_HOST}/api/user`, {
    headers: { cookie: request.cookies.toString() },
  });

const rewriteWithUser = (request: NextRequest, userJson: string) => {
  const clonedRequest = request.clone();
  clonedRequest.headers.set(USER_HEADER_NAME, encodeURIComponent(userJson));
  return NextResponse.rewrite(request.url.toString(), {
    request: clonedRequest,
  });
};

export async function middleware(request: NextRequest) {
  const sessionId = request.cookies.get(SESSION_ID_COOKIE_NAME)?.value;

  const path = request.nextUrl.pathname;

  if (unprotectedPages.some((prefix) => path.startsWith(prefix))) {
    if (
      sessionId &&
      (path.startsWith("/login") || path.startsWith("/register"))
    )
      return Response.redirect(new URL("/lobby", request.url));
    if (sessionId) {
      const resp = await fetchUser(request);
      if (resp.ok) return rewriteWithUser(request, await resp.text());
    }
    return NextResponse.next();
  }

  if (!sessionId) return Response.redirect(new URL("/login", request.url));

  const resp = await fetchUser(request);

  if (!resp.ok) {
    const response = NextResponse.redirect(new URL("/login", request.url));
    response.cookies.delete(SESSION_ID_COOKIE_NAME);
    return response;
  }

  return rewriteWithUser(request, await resp.text());
}
