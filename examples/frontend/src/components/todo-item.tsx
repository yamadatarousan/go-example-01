// TODOアイテムコンポーネント

import type { Todo } from "@/types";

type TodoItemProps = {
  todo: Todo;
};

export function TodoItem({ todo }: TodoItemProps) {
  return (
    <li className="rounded-md border p-4 space-y-1">
      <p className="font-medium">{todo.name}</p>
      {todo.description && (
        <p className="text-sm text-gray-500">{todo.description}</p>
      )}
      <div className="text-xs text-gray-500">
        <span>ステータス: {todo.status}</span>
        <span className="mx-2">/</span>
        <span>優先度: {todo.priority}</span>
      </div>
    </li>
  );
}
