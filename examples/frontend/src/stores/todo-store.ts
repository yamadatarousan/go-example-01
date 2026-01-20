// TODO の状態管理ストア
// Zustand を使用してクライアントサイドの状態を管理

import { create } from 'zustand'
import type { Todo } from '@/types'

// ストアの型定義
interface TodoStore {
  // 状態
  todos: Todo[]
  isLoading: boolean
  error: string | null

  // アクション
  setTodos: (todos: Todo[]) => void
  addTodo: (todo: Todo) => void
  updateTodo: (id: number, updates: Partial<Todo>) => void
  deleteTodo: (id: number) => void
  setLoading: (isLoading: boolean) => void
  setError: (error: string | null) => void
}

// Zustand ストアの作成
export const useTodoStore = create<TodoStore>((set) => ({
  // 初期状態
  todos: [],
  isLoading: false,
  error: null,

  // TODO リストを設定
  setTodos: (todos) => set({ todos }),

  // TODO を追加
  addTodo: (todo) =>
    set((state) => ({
      todos: [...state.todos, todo],
    })),

  // TODO を更新
  updateTodo: (id, updates) =>
    set((state) => ({
      todos: state.todos.map((todo) =>
        todo.id === id ? { ...todo, ...updates } : todo
      ),
    })),

  // TODO を削除
  deleteTodo: (id) =>
    set((state) => ({
      todos: state.todos.filter((todo) => todo.id !== id),
    })),

  // ローディング状態を設定
  setLoading: (isLoading) => set({ isLoading }),

  // エラーを設定
  setError: (error) => set({ error }),
}))
