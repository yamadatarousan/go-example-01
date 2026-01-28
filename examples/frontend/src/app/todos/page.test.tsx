import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

const mockGetTodos = vi.hoisted(() => vi.fn());

vi.mock("./actions", () => ({
  getTodos: mockGetTodos,
}));

import TodosPage from "./page";

describe("TodosPage", () => {
  it("TODOが0件のときに空状態を表示する", async () => {
    mockGetTodos.mockResolvedValueOnce([]);

    render(await TodosPage());

    expect(screen.getByText("TODOがまだありません")).toBeInTheDocument();
  });

  it("取得失敗時に簡易エラーを表示する", async () => {
    mockGetTodos.mockRejectedValueOnce(new Error("取得エラー"));

    render(await TodosPage());

    expect(screen.getByText("取得エラー")).toBeInTheDocument();
  });

  it("TODO一覧を表示する", async () => {
    mockGetTodos.mockResolvedValueOnce([
      {
        id: 1,
        name: "テストTODO",
        description: "説明文",
        status: "todo",
        priority: "medium",
        user_id: 1,
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      },
    ]);

    render(await TodosPage());

    expect(screen.getByText("テストTODO")).toBeInTheDocument();
    expect(screen.getByText("説明文")).toBeInTheDocument();
    expect(screen.getByText("ステータス: todo")).toBeInTheDocument();
    expect(screen.getByText("優先度: medium")).toBeInTheDocument();
  });
});
