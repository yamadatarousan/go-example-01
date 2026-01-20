// TODO アイテムコンポーネント
// Client Component: 個別の TODO の表示と操作

'use client'

import { useState } from 'react'
import { Checkbox } from '@/components/ui/checkbox'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useTodoStore } from '@/stores/todo-store'
import { updateTodoAction, deleteTodoAction } from './actions'
import { cn } from '@/lib/utils'
import type { Todo } from '@/types'

interface TodoItemProps {
  todo: Todo
}

export function TodoItem({ todo }: TodoItemProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editTitle, setEditTitle] = useState(todo.title)
  const [isLoading, setIsLoading] = useState(false)
  const { updateTodo, deleteTodo, setError } = useTodoStore()

  // 完了状態の切り替え
  const handleToggleComplete = async () => {
    setIsLoading(true)
    try {
      const result = await updateTodoAction(todo.id, {
        completed: !todo.completed,
      })

      if (result.error) {
        setError(result.error)
      } else if (result.todo) {
        updateTodo(todo.id, result.todo)
      }
    } catch (err) {
      setError('更新に失敗しました')
    } finally {
      setIsLoading(false)
    }
  }

  // タイトルの更新
  const handleUpdateTitle = async () => {
    if (editTitle.trim() === '' || editTitle === todo.title) {
      setIsEditing(false)
      setEditTitle(todo.title)
      return
    }

    setIsLoading(true)
    try {
      const result = await updateTodoAction(todo.id, {
        title: editTitle.trim(),
      })

      if (result.error) {
        setError(result.error)
        setEditTitle(todo.title)
      } else if (result.todo) {
        updateTodo(todo.id, result.todo)
      }
    } catch (err) {
      setError('更新に失敗しました')
      setEditTitle(todo.title)
    } finally {
      setIsLoading(false)
      setIsEditing(false)
    }
  }

  // TODO の削除
  const handleDelete = async () => {
    setIsLoading(true)
    try {
      const result = await deleteTodoAction(todo.id)

      if (result.error) {
        setError(result.error)
      } else {
        deleteTodo(todo.id)
      }
    } catch (err) {
      setError('削除に失敗しました')
    } finally {
      setIsLoading(false)
    }
  }

  // Enter キーで編集を確定
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleUpdateTitle()
    } else if (e.key === 'Escape') {
      setIsEditing(false)
      setEditTitle(todo.title)
    }
  }

  return (
    <div
      className={cn(
        'flex items-center gap-3 p-3 rounded-lg border bg-card',
        isLoading && 'opacity-50 pointer-events-none'
      )}
    >
      {/* チェックボックス */}
      <Checkbox
        checked={todo.completed}
        onCheckedChange={handleToggleComplete}
        disabled={isLoading}
      />

      {/* タイトル（編集モード切り替え） */}
      {isEditing ? (
        <Input
          value={editTitle}
          onChange={(e) => setEditTitle(e.target.value)}
          onBlur={handleUpdateTitle}
          onKeyDown={handleKeyDown}
          autoFocus
          className="flex-1"
        />
      ) : (
        <span
          className={cn(
            'flex-1 cursor-pointer',
            todo.completed && 'line-through text-muted-foreground'
          )}
          onClick={() => setIsEditing(true)}
        >
          {todo.title}
        </span>
      )}

      {/* 削除ボタン */}
      <Button
        variant="ghost"
        size="sm"
        onClick={handleDelete}
        disabled={isLoading}
        className="text-destructive hover:text-destructive hover:bg-destructive/10"
      >
        削除
      </Button>
    </div>
  )
}
