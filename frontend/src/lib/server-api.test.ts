// server-api.ts のユニットテスト
// fetch / cookies / 環境変数をモックして、通信・認証・エラー分岐を検証する

import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";

// ---------------------------------------------------------------------------
// グローバルfetchのモック
// ---------------------------------------------------------------------------
// server-api.ts は fetch を直接呼ぶため、ここで差し替えて挙動を制御する
const mockFetch = vi.fn();

vi.stubGlobal("fetch", mockFetch);

// ---------------------------------------------------------------------------
// next/headers の cookies() モック
// ---------------------------------------------------------------------------
// server-api.ts の fetchWithAuth は cookies().get("token") を使うため、
// cookies().get をテストから制御できるようにする
const mockCookiesGet = vi.fn();
vi.mock("next/headers", () => ({
  cookies: async () => ({
    get: mockCookiesGet,
  }),
}));

// ---------------------------------------------------------------------------
// モジュール再読み込みヘルパー
// ---------------------------------------------------------------------------
// server-api.ts はモジュール読み込み時に環境変数（API_BASE_URL / USE_MOCK）を読む。
// そのため、テストごとに env を差し替えてから import し直す必要がある。
// vi.resetModules() でキャッシュを捨て、最新の env を反映する。
const loadModule = async (env: Record<string, string>) => {
  const originalEnv = { ...process.env };
  process.env = { ...process.env, ...env };
  vi.resetModules();
  const serverApi = await import("./server-api");
  process.env = originalEnv;
  return serverApi;
};

describe("server-api", () => {
  beforeEach(() => {
    // テスト間で呼び出し履歴が混ざらないようにリセット
    mockFetch.mockReset();
    mockCookiesGet.mockReset();
  });

  afterEach(() => {
    mockFetch.mockReset();
  });

  it("fetchWithoutAuth: 成功時にJSONを返す", async () => {
    // USE_MOCK=false で実際の fetch を通す設定
    const { fetchWithoutAuth } = await loadModule({
      API_BASE_URL: "http://localhost:8080",
      USE_MOCK: "false",
    });

    // fetch の成功レスポンスをスタブ
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ message: "ok" }),
    });

    const result = await fetchWithoutAuth<{ message: string }>("/ping");

    // 戻り値と fetch 呼び出し内容を検証
    expect(result).toEqual({ message: "ok" });
    expect(mockFetch).toHaveBeenCalledWith("http://localhost:8080/ping", {
      headers: { "Content-Type": "application/json" },
    });
  });

  it("fetchWithoutAuth: 失敗時にエラーを投げる", async () => {
    const { fetchWithoutAuth } = await loadModule({
      API_BASE_URL: "http://localhost:8080",
      USE_MOCK: "false",
    });

    // エラーレスポンス（errorフィールドあり）を返す
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: async () => ({ error: "Unauthorized" }),
    });

    await expect(fetchWithoutAuth("/login")).rejects.toThrow("Unauthorized");
  });

  it("fetchWithAuth: トークンがない場合はエラー", async () => {
    const { fetchWithAuth } = await loadModule({
      API_BASE_URL: "http://localhost:8080",
      USE_MOCK: "false",
    });

    // cookies().get("token") が undefined を返す想定
    mockCookiesGet.mockReturnValueOnce(undefined);

    await expect(fetchWithAuth("/api/v1/todos")).rejects.toThrow("認証が必要です");
  });

  it("fetchWithAuth: 成功時にAuthorizationヘッダーを付与する", async () => {
    const { fetchWithAuth } = await loadModule({
      API_BASE_URL: "http://localhost:8080",
      USE_MOCK: "false",
    });

    // Cookie からトークンが取得できる想定
    mockCookiesGet.mockReturnValueOnce({ value: "token-value" });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ todos: [] }),
    });

    const result = await fetchWithAuth<{ todos: unknown[] }>("/api/v1/todos");

    // Authorization ヘッダーが付与されていることを検証
    expect(result).toEqual({ todos: [] });
    expect(mockFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/todos", {
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer token-value",
      },
    });
  });

  it("fetchWithAuth: 失敗時にエラーを投げる", async () => {
    const { fetchWithAuth } = await loadModule({
      API_BASE_URL: "http://localhost:8080",
      USE_MOCK: "false",
    });

    // 認証トークンはあるが、API側が失敗する想定
    mockCookiesGet.mockReturnValueOnce({ value: "token-value" });
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: async () => ({ error: "Server Error" }),
    });

    await expect(fetchWithAuth("/api/v1/todos")).rejects.toThrow("Server Error");
  });
});
