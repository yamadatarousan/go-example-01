// ============================================================================
// サインアップの Server Action
// ============================================================================

"use server";

import { cookies } from "next/headers";
import { fetchWithoutAuth } from "@/lib/server-api";
import type { LoginInput, SignupInput, User, ActionResult } from "@/types";

// サインアップAPIのレスポンス型
// バックエンドの /signup エンドポイントが返す形式
type SignupResponse = User;

// ログインAPIのレスポンス型
type LoginResponse = {
  token: string;
  user: User;
};

// signup Server Action
export async function signup(input: SignupInput): Promise<ActionResult<User>> {
  try {
    const response = await fetchWithoutAuth<SignupResponse>("/signup", {
      method: "POST",
      body: JSON.stringify(input),
    });

    // サインアップ後に自動ログイン
    const loginInput: LoginInput = {
      email: input.email,
      password: input.password,
    };
    const loginResponse = await fetchWithoutAuth<LoginResponse>("/login", {
      method: "POST",
      body: JSON.stringify(loginInput),
    });

    const cookieStore = await cookies();
    cookieStore.set("token", loginResponse.token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: 60 * 60 * 24 * 7,
    });

    return {
      success: true,
      data: response,
    };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "サインアップに失敗しました",
    };
  }
}
