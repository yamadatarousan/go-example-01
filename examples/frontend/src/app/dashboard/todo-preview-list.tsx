// TODO プレビューリストコンポーネント
// ダッシュボードで TODO の概要を表示する

import Link from 'next/link'
import type { TodayTodo } from './actions'

interface TodoPreviewListProps {
  todos: TodayTodo[]
  emptyMessage: string
}

// 優先度に応じた色を返す
function getPriorityColor(priority: string): string {
  switch (priority) {
    case 'high':
      return 'bg-red-100 text-red-800'
    case 'medium':
      return 'bg-yellow-100 text-yellow-800'
    case 'low':
      return 'bg-green-100 text-green-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

// 優先度の日本語表示
function getPriorityLabel(priority: string): string {
  switch (priority) {
    case 'high':
      return '高'
    case 'medium':
      return '中'
    case 'low':
      return '低'
    default:
      return priority
  }
}

// 日付のフォーマット
function formatDate(dateString: string): string {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleDateString('ja-JP', {
    month: 'short',
    day: 'numeric',
  })
}

export function TodoPreviewList({ todos, emptyMessage }: TodoPreviewListProps) {
  if (todos.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        <svg
          className="mx-auto h-12 w-12 text-gray-400 mb-3"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
          />
        </svg>
        <p>{emptyMessage}</p>
      </div>
    )
  }

  return (
    <ul className="divide-y divide-gray-200">
      {todos.slice(0, 5).map((todo) => (
        <li key={todo.id}>
          <Link
            href={`/todos/${todo.id}`}
            className="block py-3 hover:bg-gray-50 -mx-2 px-2 rounded transition-colors"
          >
            <div className="flex items-center justify-between">
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 truncate">
                  {todo.name}
                </p>
                {todo.due_date && (
                  <p className="text-xs text-gray-500 mt-1">
                    期限: {formatDate(todo.due_date)}
                  </p>
                )}
              </div>
              <span
                className={`ml-3 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${getPriorityColor(
                  todo.priority
                )}`}
              >
                {getPriorityLabel(todo.priority)}
              </span>
            </div>
          </Link>
        </li>
      ))}
      {todos.length > 5 && (
        <li className="py-3 text-center text-sm text-gray-500">
          他 {todos.length - 5} 件
        </li>
      )}
    </ul>
  )
}
