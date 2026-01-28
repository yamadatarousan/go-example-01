// ============================================================================
// Server Action とは
// ============================================================================
// Server Action は Next.js の機能で、クライアントから直接呼び出せるサーバー側関数。
// 従来の流れ: クライアント → fetch("/api/login") → API Route → 処理
// Server Action: クライアント → login() → 処理（直接呼び出し）
//
// API Routeを別途作る必要がなく、関数を直接importして呼べる。
// ただし実行は必ずサーバー側で行われる（クライアントにコードは送られない）。

// ============================================================================
// "use server" ディレクティブ
// ============================================================================
// このファイル内の全ての関数を Server Action として扱う宣言。
// これがないと、クライアントで実行されてしまう可能性がある。
//
// 書ける場所:
// 1. ファイルの先頭（ファイル内の全関数が対象）← 今回はこれ
// 2. 関数の中（その関数だけが対象）
//    async function login() {
//      "use server";
//      ...
//    }

"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { fetchWithoutAuth } from "@/lib/server-api";
import type { LoginInput, ActionResult } from "@/types";

// ============================================================================
// ログインAPIのレスポンス型
// ============================================================================
// バックエンドの /login エンドポイントが返す形式

type LoginResponse = {
  access_token: string;
  refresh_token: string;
};

// ============================================================================
// login Server Action
// ============================================================================
// クライアント（LoginForm）から呼び出されるサーバー側関数。
//
// Server Action の制約:
// - 引数はシリアライズ可能な値のみ（関数やクラスインスタンスは不可）
// - 戻り値もシリアライズ可能な値のみ
// - async 関数である必要がある

export async function login(input: LoginInput): Promise<ActionResult<null>> {
  try {
    // ------------------------------------------------------------------------
    // 1. バックエンドAPIの呼び出し
    // ------------------------------------------------------------------------
    // Server Action内なのでサーバー側で実行される。
    // fetchWithoutAuthは認証不要のAPI呼び出し用ユーティリティ。

    const response = await fetchWithoutAuth<LoginResponse>("/login", {
      method: "POST",
      body: JSON.stringify(input),
    });

    // ------------------------------------------------------------------------
    // 2. JWTトークンをHttpOnly Cookieに保存
    // ------------------------------------------------------------------------
    // cookies() は Next.js が提供するサーバー側専用の関数。
    // Server Action内でしか使えない（Client Componentでは使えない）。
    //
    // HttpOnly: true にすることで、ブラウザのJavaScriptからは
    // このCookieにアクセスできなくなる（XSS対策）。
    //
    // Secure: 本番環境ではHTTPS通信でのみCookieを送信。
    // SameSite: CSRF対策。"lax"は同一サイトからのリクエストのみCookieを送信。
    // Path: "/"で全てのパスでCookieが有効。
    // MaxAge: Cookieの有効期限（秒）。7日間 = 60 * 60 * 24 * 7

    const cookieStore = await cookies();
    cookieStore.set("token", response.access_token, {
      httpOnly: true,                                    // JSからアクセス不可
      secure: process.env.NODE_ENV === "production",    // 本番はHTTPSのみ
      sameSite: "lax",                                   // CSRF対策
      path: "/",                                         // 全パスで有効
      maxAge: 60 * 60 * 24 * 7,                         // 7日間
    });

    // ------------------------------------------------------------------------
    // 3. 結果を返す
    // ------------------------------------------------------------------------
    // Server Action から Client Component に値を返す。
    // ここでは成功/失敗を明示的に返し、クライアント側で処理を分岐する。
    //
    // 注意: redirect() はここでは使わない。
    // Server Action内でredirect()を呼ぶと、戻り値が返らずに
    // 直接リダイレクトしてしまう。クライアント側でエラー表示等の
    // 処理をしたい場合は、戻り値を返してクライアントで判断する。

    return {
      success: true,
      data: null,
    };
  } catch (error) {
    // ------------------------------------------------------------------------
    // エラーハンドリング
    // ------------------------------------------------------------------------
    // Server Action内でエラーが発生した場合、そのままthrowすると
    // クライアント側で詳細なエラー情報が見えてしまう可能性がある。
    // セキュリティのため、汎用的なエラーメッセージを返すのが一般的。

    return {
      success: false,
      error: error instanceof Error ? error.message : "ログインに失敗しました",
    };
  }
}

// ============================================================================
// Server Action の呼び出し方（参考）
// ============================================================================
// Client Component から以下のように呼び出す:
//
// import { login } from "./actions";
//
// const onSubmit = async (data: LoginFormData) => {
//   const result = await login(data);  // ← 普通の関数呼び出しに見える
//
//   if (result.success) {
//     router.push("/dashboard");  // クライアント側でリダイレクト
//   } else {
//     setError(result.error);     // エラー表示
//   }
// };
//
// 見た目は普通の関数呼び出しだが、実際には:
// 1. Next.jsが内部的にHTTPリクエストを生成
// 2. サーバーで login() が実行される
// 3. 結果がシリアライズされてクライアントに返る
