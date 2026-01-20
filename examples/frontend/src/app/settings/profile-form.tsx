'use client'

// プロフィール編集フォームコンポーネント

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { updateProfileAction, type UserProfile } from './actions'

interface ProfileFormProps {
  profile: UserProfile
}

export function ProfileForm({ profile }: ProfileFormProps) {
  const [isPending, startTransition] = useTransition()
  const [isEditing, setIsEditing] = useState(false)
  const [username, setUsername] = useState(profile.username)
  const [bio, setBio] = useState(profile.bio || '')
  const router = useRouter()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!username.trim()) {
      toast.error('ユーザー名を入力してください')
      return
    }

    startTransition(async () => {
      const result = await updateProfileAction({
        username: username.trim(),
        bio: bio.trim() || undefined,
      })

      if (result.error) {
        toast.error(result.error)
      } else {
        toast.success('プロフィールを更新しました')
        setIsEditing(false)
        router.refresh()
      }
    })
  }

  const handleCancel = () => {
    setUsername(profile.username)
    setBio(profile.bio || '')
    setIsEditing(false)
  }

  if (!isEditing) {
    return (
      <div>
        <div className="flex items-start space-x-4">
          {/* アバター */}
          <div className="w-16 h-16 rounded-full bg-gray-200 flex items-center justify-center flex-shrink-0">
            {profile.image ? (
              <img
                src={profile.image}
                alt={profile.username}
                className="w-16 h-16 rounded-full object-cover"
              />
            ) : (
              <span className="text-2xl font-medium text-gray-600">
                {profile.username.charAt(0).toUpperCase()}
              </span>
            )}
          </div>

          {/* 情報 */}
          <div className="flex-1">
            <h3 className="text-lg font-medium text-gray-900">
              {profile.username}
            </h3>
            {profile.bio ? (
              <p className="text-gray-600 mt-1">{profile.bio}</p>
            ) : (
              <p className="text-gray-400 mt-1 italic">自己紹介が設定されていません</p>
            )}
          </div>

          {/* 編集ボタン */}
          <Button variant="outline" onClick={() => setIsEditing(true)}>
            編集
          </Button>
        </div>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit}>
      <div className="space-y-4">
        <div className="flex items-start space-x-4">
          {/* アバター */}
          <div className="w-16 h-16 rounded-full bg-gray-200 flex items-center justify-center flex-shrink-0">
            <span className="text-2xl font-medium text-gray-600">
              {username.charAt(0).toUpperCase()}
            </span>
          </div>

          {/* フォーム */}
          <div className="flex-1 space-y-4">
            <div>
              <Label htmlFor="username">ユーザー名</Label>
              <Input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={isPending}
                placeholder="ユーザー名"
              />
            </div>

            <div>
              <Label htmlFor="bio">自己紹介</Label>
              <textarea
                id="bio"
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                disabled={isPending}
                placeholder="自己紹介を入力（任意）"
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                rows={3}
              />
            </div>
          </div>
        </div>

        <div className="flex justify-end space-x-3">
          <Button
            type="button"
            variant="outline"
            onClick={handleCancel}
            disabled={isPending}
          >
            キャンセル
          </Button>
          <Button type="submit" disabled={isPending}>
            {isPending ? '保存中...' : '保存'}
          </Button>
        </div>
      </div>
    </form>
  )
}
