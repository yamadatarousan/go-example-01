import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const protectedPaths = ["/todos", "/dashboard", "/projects", "/settings"];
const authOnlyPaths = ["/login", "/signup"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get("token")?.value;

  // 公開ページはそのまま許可
  if (pathname === "/") {
    return NextResponse.next();
  }

  // ログイン必須ページ
  if (protectedPaths.some((path) => pathname.startsWith(path))) {
    if (!token) {
      const url = request.nextUrl.clone();
      url.pathname = "/login";
      return NextResponse.redirect(url);
    }
    return NextResponse.next();
  }

  // 未ログイン専用ページ（ログイン済みは /todos へ）
  if (authOnlyPaths.some((path) => pathname.startsWith(path))) {
    if (token) {
      const url = request.nextUrl.clone();
      url.pathname = "/todos";
      return NextResponse.redirect(url);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
