// ============================================================================
// サインアップの Server Action
// ============================================================================

"use server";

import { fetchWithoutAuth } from "@/lib/server-api";
import type { SignupInput, User, ActionResult } from "@/types";

// サインアップAPIのレスポンス型
// バックエンドの /signup エンドポイントが返す形式
type SignupResponse = User;

// signup Server Action
export async function signup(input: SignupInput): Promise<ActionResult<User>> {
  try {
    const response = await fetchWithoutAuth<SignupResponse>("/signup", {
      method: "POST",
      body: JSON.stringify(input),
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
