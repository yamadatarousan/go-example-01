// TODO一覧ページ

import { getTodos } from "./actions";
import type { Todo } from "@/types";

export default async function TodosPage() {
  let todos: Todo[] = [];
  let errorMessage: string | null = null;

  try {
    todos = await getTodos();
  } catch (error) {
    errorMessage = error instanceof Error ? error.message : "TODOの取得に失敗しました";
  }

  return (
    <main className="mx-auto w-full max-w-3xl p-6 space-y-6">
      <header className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">TODO一覧</h1>
      </header>

      {errorMessage && (
        <p className="text-sm text-red-500">{errorMessage}</p>
      )}

      {!errorMessage && todos.length === 0 && (
        <p className="text-sm text-gray-500">TODOがまだありません</p>
      )}

      {!errorMessage && todos.length > 0 && (
        <ul className="space-y-3">
          {todos.map((todo) => (
            <li key={todo.id} className="rounded-md border p-4">
              <p className="font-medium">{todo.name}</p>
              {todo.description && (
                <p className="text-sm text-gray-500">{todo.description}</p>
              )}
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
