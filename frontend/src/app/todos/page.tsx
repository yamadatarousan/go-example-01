// TODO一覧ページ

import { getTodos } from "./actions";
import type { Todo } from "@/types";
import { TodoItem } from "@/components/todo-item";

export default async function TodosPage() {
  let todos: Todo[] = [];
  let errorMessage: string | null = null;

  try {
    todos = await getTodos();
  } catch (error) {
    errorMessage = error instanceof Error ? error.message : "TODOの取得に失敗しました";
  }

  const safeTodos = Array.isArray(todos) ? todos : [];

  return (
    <main className="mx-auto w-full max-w-3xl p-6 space-y-6">
      <header className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">TODO一覧</h1>
        <div className="text-sm text-gray-500">ログイン済み</div>
      </header>

      {errorMessage && (
        <p className="text-sm text-red-500">{errorMessage}</p>
      )}

      {!errorMessage && safeTodos.length === 0 && (
        <p className="text-sm text-gray-500">TODOがまだありません</p>
      )}

      {!errorMessage && safeTodos.length > 0 && (
        <ul className="space-y-3">
          {safeTodos.map((todo) => (
            <TodoItem key={todo.id} todo={todo} />
          ))}
        </ul>
      )}
    </main>
  );
}
