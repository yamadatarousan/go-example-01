// プロジェクト管理用 Server Actions

'use server'

import { revalidatePath } from 'next/cache'
import { fetchWithAuth } from '@/lib/server-api'

// プロジェクトの型定義
export interface Project {
  id: number
  name: string
  description: string
  owner_id: number
  created_at: string
  updated_at: string
}

// プロジェクトメンバーの型定義
export interface ProjectMember {
  user_id: number
  email: string
  role: string
  joined_at: string
}

// プロジェクト作成リクエスト
interface CreateProjectRequest {
  name: string
  description?: string
}

// プロジェクト更新リクエスト
interface UpdateProjectRequest {
  name?: string
  description?: string
}

/**
 * プロジェクト一覧を取得する Server Action
 */
export async function getProjectsAction(): Promise<{
  data?: Project[]
  error?: string
}> {
  const result = await fetchWithAuth<{ projects: Project[] }>('/api/v1/projects')

  if (result.error) {
    return { error: result.error }
  }

  return { data: result.data?.projects || [] }
}

/**
 * プロジェクト詳細を取得する Server Action
 */
export async function getProjectAction(id: number): Promise<{
  data?: Project
  error?: string
}> {
  const result = await fetchWithAuth<Project>(`/api/v1/projects/${id}`)
  return result
}

/**
 * プロジェクトを作成する Server Action
 */
export async function createProjectAction(data: CreateProjectRequest): Promise<{
  data?: Project
  error?: string
}> {
  const result = await fetchWithAuth<Project>('/api/v1/projects', {
    method: 'POST',
    body: JSON.stringify(data),
  })

  if (result.data) {
    revalidatePath('/projects')
  }

  return result
}

/**
 * プロジェクトを更新する Server Action
 */
export async function updateProjectAction(
  id: number,
  data: UpdateProjectRequest
): Promise<{
  data?: Project
  error?: string
}> {
  const result = await fetchWithAuth<Project>(`/api/v1/projects/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })

  if (result.data) {
    revalidatePath('/projects')
    revalidatePath(`/projects/${id}`)
  }

  return result
}

/**
 * プロジェクトを削除する Server Action
 */
export async function deleteProjectAction(id: number): Promise<{
  success?: boolean
  error?: string
}> {
  const result = await fetchWithAuth<void>(`/api/v1/projects/${id}`, {
    method: 'DELETE',
  })

  if (result.error) {
    return { error: result.error }
  }

  revalidatePath('/projects')
  return { success: true }
}

/**
 * プロジェクトメンバー一覧を取得する Server Action
 */
export async function getProjectMembersAction(projectId: number): Promise<{
  data?: ProjectMember[]
  error?: string
}> {
  const result = await fetchWithAuth<{ members: ProjectMember[] }>(
    `/api/v1/projects/${projectId}/members`
  )

  if (result.error) {
    return { error: result.error }
  }

  return { data: result.data?.members || [] }
}

/**
 * プロジェクトにメンバーを追加する Server Action
 */
export async function addProjectMemberAction(
  projectId: number,
  userId: number,
  role: string = 'member'
): Promise<{
  success?: boolean
  error?: string
}> {
  const result = await fetchWithAuth<void>(`/api/v1/projects/${projectId}/members`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId, role }),
  })

  if (result.error) {
    return { error: result.error }
  }

  revalidatePath(`/projects/${projectId}`)
  return { success: true }
}

/**
 * プロジェクトからメンバーを削除する Server Action
 */
export async function removeProjectMemberAction(
  projectId: number,
  userId: number
): Promise<{
  success?: boolean
  error?: string
}> {
  const result = await fetchWithAuth<void>(
    `/api/v1/projects/${projectId}/members/${userId}`,
    {
      method: 'DELETE',
    }
  )

  if (result.error) {
    return { error: result.error }
  }

  revalidatePath(`/projects/${projectId}`)
  return { success: true }
}
