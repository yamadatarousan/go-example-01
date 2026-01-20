// Next.js ミドルウェア
// 認証が必要なページへのアクセスを制御する
//
// ミドルウェアはリクエストが完了する前に実行され、
// レスポンスを変更したり、リダイレクトしたりできる

import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// 認証が必要なパス（これらのパスにアクセスするにはログインが必要）
const protectedPaths = ['/todos', '/dashboard', '/projects', '/settings']

// 認証済みユーザーがアクセスすべきでないパス（ログイン後はリダイレクト）
const authPaths = ['/login', '/signup']

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Cookie から JWT トークンを取得
  const token = request.cookies.get('token')?.value

  // 認証が必要なページへのアクセス
  if (protectedPaths.some((path) => pathname.startsWith(path))) {
    // トークンがない場合はログインページへリダイレクト
    if (!token) {
      const loginUrl = new URL('/login', request.url)
      // リダイレクト後に元のページに戻れるよう、元の URL を保存
      loginUrl.searchParams.set('callbackUrl', pathname)
      return NextResponse.redirect(loginUrl)
    }
  }

  // 認証済みユーザーが認証ページにアクセスした場合
  if (authPaths.some((path) => pathname.startsWith(path))) {
    // トークンがある場合は TODO 一覧へリダイレクト
    if (token) {
      return NextResponse.redirect(new URL('/todos', request.url))
    }
  }

  // それ以外はそのまま続行
  return NextResponse.next()
}

// ミドルウェアを適用するパスの設定
// 静的ファイル（_next, favicon 等）は除外
export const config = {
  matcher: [
    /*
     * 以下で始まるパスを除外:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
}
