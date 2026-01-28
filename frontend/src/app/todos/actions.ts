// TODO取得用のServer Action

"use server";

import { fetchWithAuth } from "@/lib/server-api";
import type { Todo } from "@/types";

// TODO一覧を取得する
export async function getTodos(): Promise<Todo[]> {
  return fetchWithAuth<Todo[]>("/api/v1/todos");
}
