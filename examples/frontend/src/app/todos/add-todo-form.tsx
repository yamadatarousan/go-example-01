// TODO 追加フォーム
// Client Component: 新しい TODO を作成

'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useTodoStore } from '@/stores/todo-store'
import { createTodoAction } from './actions'

// バリデーションスキーマ
const addTodoSchema = z.object({
  title: z
    .string()
    .min(1, 'タイトルを入力してください')
    .max(100, 'タイトルは100文字以内で入力してください'),
})

type AddTodoFormData = z.infer<typeof addTodoSchema>

export function AddTodoForm() {
  const [isLoading, setIsLoading] = useState(false)
  const { addTodo, setError } = useTodoStore()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AddTodoFormData>({
    resolver: zodResolver(addTodoSchema),
  })

  const onSubmit = async (data: AddTodoFormData) => {
    setIsLoading(true)
    setError(null)

    try {
      // Server Action を呼び出し
      const result = await createTodoAction(data.title)

      if (result.error) {
        setError(result.error)
      } else if (result.todo) {
        // 成功: ストアに追加してフォームをリセット
        addTodo(result.todo)
        reset()
      }
    } catch (err) {
      setError('TODO の作成に失敗しました')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex gap-2">
      <div className="flex-1">
        <Input
          placeholder="新しいタスクを入力..."
          {...register('title')}
          disabled={isLoading}
        />
        {errors.title && (
          <p className="text-sm text-destructive mt-1">{errors.title.message}</p>
        )}
      </div>
      <Button type="submit" disabled={isLoading}>
        {isLoading ? '追加中...' : '追加'}
      </Button>
    </form>
  )
}
