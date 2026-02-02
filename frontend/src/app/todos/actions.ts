// TODO取得用のServer Action

"use server";

import { fetchWithAuth } from "@/lib/server-api";
import type { components } from "@/openapi/types";

type Todo = components["schemas"]["Todo"];

// TODO一覧を取得する
export async function getTodos(): Promise<Todo[]> {
  const response = await fetchWithAuth<unknown>("/api/v1/todos");
  if (!Array.isArray(response)) {
    throw new Error("TODO一覧の取得に失敗しました");
  }
  return response as Todo[];
}
