'use client'

// プロジェクトヘッダーコンポーネント
// プロジェクトの編集・削除機能を提供

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { updateProjectAction, deleteProjectAction, type Project } from '../actions'

interface ProjectHeaderProps {
  project: Project
}

export function ProjectHeader({ project }: ProjectHeaderProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [name, setName] = useState(project.name)
  const [description, setDescription] = useState(project.description || '')
  const router = useRouter()

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      toast.error('プロジェクト名を入力してください')
      return
    }

    startTransition(async () => {
      const result = await updateProjectAction(project.id, {
        name: name.trim(),
        description: description.trim() || undefined,
      })

      if (result.error) {
        toast.error(result.error)
      } else {
        toast.success('プロジェクトを更新しました')
        setIsEditing(false)
        router.refresh()
      }
    })
  }

  const handleDelete = async () => {
    startTransition(async () => {
      const result = await deleteProjectAction(project.id)

      if (result.error) {
        toast.error(result.error)
      } else {
        toast.success('プロジェクトを削除しました')
        router.push('/projects')
      }
    })
  }

  if (isEditing) {
    return (
      <form onSubmit={handleUpdate}>
        <div className="space-y-4">
          <div>
            <Label htmlFor="name">プロジェクト名</Label>
            <Input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={isPending}
            />
          </div>
          <div>
            <Label htmlFor="description">説明</Label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              disabled={isPending}
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
              rows={3}
            />
          </div>
          <div className="flex space-x-2">
            <Button type="submit" disabled={isPending}>
              {isPending ? '保存中...' : '保存'}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setIsEditing(false)
                setName(project.name)
                setDescription(project.description || '')
              }}
              disabled={isPending}
            >
              キャンセル
            </Button>
          </div>
        </div>
      </form>
    )
  }

  return (
    <div className="flex justify-between items-start">
      <div>
        <h2 className="text-xl font-semibold text-gray-900">{project.name}</h2>
      </div>
      <div className="flex space-x-2">
        <Button variant="outline" size="sm" onClick={() => setIsEditing(true)}>
          編集
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="text-red-600 hover:text-red-700"
          onClick={() => setIsDeleting(true)}
        >
          削除
        </Button>
      </div>

      {/* 削除確認モーダル */}
      {isDeleting && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div
            className="absolute inset-0 bg-black/50"
            onClick={() => setIsDeleting(false)}
          />
          <div className="relative bg-white rounded-lg shadow-xl w-full max-w-sm mx-4 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              プロジェクトを削除
            </h3>
            <p className="text-gray-600 mb-4">
              「{project.name}」を削除しますか？この操作は取り消せません。
            </p>
            <div className="flex justify-end space-x-3">
              <Button
                variant="outline"
                onClick={() => setIsDeleting(false)}
                disabled={isPending}
              >
                キャンセル
              </Button>
              <Button
                variant="default"
                className="bg-red-600 hover:bg-red-700"
                onClick={handleDelete}
                disabled={isPending}
              >
                {isPending ? '削除中...' : '削除'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
