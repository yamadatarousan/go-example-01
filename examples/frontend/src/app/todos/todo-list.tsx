// TODO リストコンポーネント
// Client Component: TODO の CRUD 操作を処理

'use client'

import { useEffect } from 'react'
import { useTodoStore } from '@/stores/todo-store'
import { TodoItem } from './todo-item'
import { AddTodoForm } from './add-todo-form'
import type { Todo } from '@/types'

interface TodoListProps {
  // Server Component から渡される初期データ
  initialTodos: Todo[]
}

export function TodoList({ initialTodos }: TodoListProps) {
  const { todos, setTodos, error } = useTodoStore()

  // 初期データをストアに設定
  useEffect(() => {
    setTodos(initialTodos)
  }, [initialTodos, setTodos])

  return (
    <div className="space-y-6">
      {/* TODO 追加フォーム */}
      <AddTodoForm />

      {/* エラーメッセージ */}
      {error && (
        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
          {error}
        </div>
      )}

      {/* TODO リスト */}
      <div className="space-y-2">
        {todos.length === 0 ? (
          <p className="text-center text-muted-foreground py-8">
            TODO がありません。新しいタスクを追加してください。
          </p>
        ) : (
          todos.map((todo) => <TodoItem key={todo.id} todo={todo} />)
        )}
      </div>

      {/* 統計 */}
      {todos.length > 0 && (
        <div className="text-sm text-muted-foreground text-center">
          {todos.filter((t) => t.completed).length} / {todos.length} 完了
        </div>
      )}
    </div>
  )
}
