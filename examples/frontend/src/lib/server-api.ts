// API通信基盤
// Server ActionsからバックエンドAPIを呼び出すためのユーティリティ

import { cookies } from "next/headers";

// 環境変数からAPI設定を取得
const API_BASE_URL = process.env.API_BASE_URL || "http://localhost:8080";
const USE_MOCK = process.env.USE_MOCK === "true";

// ============================================================================
// 認証なしのfetch（ログイン・サインアップ用）
// ============================================================================

export async function fetchWithoutAuth<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  // モックモードの場合はモックデータを返す
  if (USE_MOCK) {
    return getMockData<T>(endpoint, options.method || "GET");
  }

  const url = `${API_BASE_URL}${endpoint}`;

  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: "Unknown error" }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
}

// ============================================================================
// 認証付きのfetch（認証が必要なAPI用）
// ============================================================================

export async function fetchWithAuth<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  // モックモードの場合はモックデータを返す
  if (USE_MOCK) {
    const cookieStore = await cookies();
    const mockTodosMode = cookieStore.get("mock_todos")?.value;
    return getMockData<T>(endpoint, options.method || "GET", {
      mockTodosMode,
    });
  }

  // CookieからJWTトークンを取得
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;

  if (!token) {
    throw new Error("認証が必要です");
  }

  const url = `${API_BASE_URL}${endpoint}`;

  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: "Unknown error" }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
}

// ============================================================================
// モックデータ（開発・テスト用）
// ============================================================================

// モックデータを返す関数
// バックエンドが未実装の場合や、フロントエンド単体でテストしたい場合に使用
function getMockData<T>(
  endpoint: string,
  method: string,
  options?: { mockTodosMode?: string }
): T {
  // エンドポイントとメソッドに応じたモックデータを返す
  const mockResponses: Record<string, Record<string, unknown>> = {
    "/login": {
      POST: { access_token: "mock-jwt-token", refresh_token: "mock-refresh-token" },
    },
    "/signup": {
      POST: { id: 1, email: "test@example.com", role: "user" },
    },
    "/api/v1/todos": {
      GET: [
        { id: 1, name: "サンプルTODO 1", status: "todo", priority: "medium", user_id: 1, created_at: "2024-01-01T00:00:00Z", updated_at: "2024-01-01T00:00:00Z" },
        { id: 2, name: "サンプルTODO 2", status: "in_progress", priority: "high", user_id: 1, created_at: "2024-01-01T00:00:00Z", updated_at: "2024-01-01T00:00:00Z" },
      ],
      POST: { id: 3, name: "新規TODO", status: "todo", priority: "medium", user_id: 1, created_at: "2024-01-01T00:00:00Z", updated_at: "2024-01-01T00:00:00Z" },
    },
    "/api/v1/todos/stats": {
      GET: { total_count: 10, status_counts: { todo: 5, in_progress: 3, done: 2 }, priority_counts: { high: 2, medium: 5, low: 3 }, overdue_count: 1, due_today_count: 2, due_this_week_count: 5 },
    },
    "/api/v1/projects": {
      GET: [
        { id: 1, name: "サンプルプロジェクト", description: "説明文", owner_id: 1, created_at: "2024-01-01T00:00:00Z", updated_at: "2024-01-01T00:00:00Z" },
      ],
      POST: { id: 2, name: "新規プロジェクト", owner_id: 1, created_at: "2024-01-01T00:00:00Z", updated_at: "2024-01-01T00:00:00Z" },
    },
  };

  if (endpoint === "/api/v1/todos" && options?.mockTodosMode === "empty") {
    mockResponses["/api/v1/todos"].GET = [];
  }

  const endpointMocks = mockResponses[endpoint];
  if (endpointMocks && endpointMocks[method]) {
    return endpointMocks[method] as T;
  }

  // 動的ルート（/api/v1/todos/123 など）のモック
  if (endpoint.match(/^\/api\/v1\/todos\/\d+$/)) {
    if (method === "GET" || method === "PUT") {
      return { id: 1, name: "TODO詳細", status: "todo", priority: "medium", user_id: 1, created_at: "2024-01-01T00:00:00Z", updated_at: "2024-01-01T00:00:00Z" } as T;
    }
    if (method === "DELETE") {
      return {} as T;
    }
  }

  throw new Error(`モックデータが定義されていません: ${method} ${endpoint}`);
}
