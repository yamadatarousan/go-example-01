import React from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";

const mockLogin = vi.hoisted(() => vi.fn());
const mockPush = vi.hoisted(() => vi.fn());

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

vi.mock("./actions", () => ({
  login: mockLogin,
}));

import { LoginForm } from "./login-form";

describe("LoginForm", () => {
  it("ログインフォームを表示する", () => {
    render(<LoginForm />);

    expect(screen.getByLabelText("メールアドレス")).toBeInTheDocument();
    expect(screen.getByLabelText("パスワード")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "ログイン" })).toBeInTheDocument();
  });

  it("ログイン失敗時にエラーメッセージを表示する", async () => {
    const user = userEvent.setup();
    mockLogin.mockResolvedValueOnce({
      success: false,
      error: "ログインに失敗しました",
    });

    render(<LoginForm />);

    await user.type(screen.getByLabelText("メールアドレス"), "test@example.com");
    await user.type(screen.getByLabelText("パスワード"), "password123");
    await user.click(screen.getByRole("button", { name: "ログイン" }));

    expect(await screen.findByText("ログインに失敗しました")).toBeInTheDocument();
  });

  it("ログイン成功時にTODO一覧へ遷移する", async () => {
    const user = userEvent.setup();
    mockLogin.mockResolvedValueOnce({
      success: true,
      data: { id: 1, email: "test@example.com", role: "user" },
    });

    render(<LoginForm />);

    await user.type(screen.getByLabelText("メールアドレス"), "test@example.com");
    await user.type(screen.getByLabelText("パスワード"), "password123");
    await user.click(screen.getByRole("button", { name: "ログイン" }));

    expect(mockPush).toHaveBeenCalledWith("/todos");
  });
});
