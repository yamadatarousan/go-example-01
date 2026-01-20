'use client'

// メンバー一覧コンポーネント
// プロジェクトメンバーの表示と削除機能を提供

import { useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { removeProjectMemberAction, type ProjectMember } from '../actions'

interface MemberListProps {
  members: ProjectMember[]
  projectId: number
}

// 役割の日本語表示
function getRoleLabel(role: string): string {
  switch (role) {
    case 'owner':
      return 'オーナー'
    case 'admin':
      return '管理者'
    case 'member':
      return 'メンバー'
    default:
      return role
  }
}

// 役割に応じたバッジ色
function getRoleBadgeColor(role: string): string {
  switch (role) {
    case 'owner':
      return 'bg-purple-100 text-purple-800'
    case 'admin':
      return 'bg-blue-100 text-blue-800'
    case 'member':
      return 'bg-gray-100 text-gray-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

export function MemberList({ members, projectId }: MemberListProps) {
  const [isPending, startTransition] = useTransition()
  const router = useRouter()

  const handleRemove = async (userId: number, email: string) => {
    if (!confirm(`${email} をプロジェクトから削除しますか？`)) {
      return
    }

    startTransition(async () => {
      const result = await removeProjectMemberAction(projectId, userId)

      if (result.error) {
        toast.error(result.error)
      } else {
        toast.success('メンバーを削除しました')
        router.refresh()
      }
    })
  }

  if (members.length === 0) {
    return (
      <p className="text-gray-500 text-sm">メンバーがいません</p>
    )
  }

  return (
    <ul className="divide-y divide-gray-200">
      {members.map((member) => (
        <li key={member.user_id} className="py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              {/* アバター */}
              <div className="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center">
                <span className="text-sm font-medium text-gray-600">
                  {member.email.charAt(0).toUpperCase()}
                </span>
              </div>
              <div>
                <p className="text-sm font-medium text-gray-900 truncate max-w-[150px]">
                  {member.email}
                </p>
                <span
                  className={`inline-block text-xs px-2 py-0.5 rounded ${getRoleBadgeColor(
                    member.role
                  )}`}
                >
                  {getRoleLabel(member.role)}
                </span>
              </div>
            </div>
            {/* オーナー以外は削除可能 */}
            {member.role !== 'owner' && (
              <Button
                variant="ghost"
                size="sm"
                className="text-red-600 hover:text-red-700 hover:bg-red-50"
                onClick={() => handleRemove(member.user_id, member.email)}
                disabled={isPending}
              >
                <svg
                  className="w-4 h-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </Button>
            )}
          </div>
        </li>
      ))}
    </ul>
  )
}
