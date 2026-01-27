import React from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";

const mockSignup = vi.hoisted(() => vi.fn());
const mockPush = vi.hoisted(() => vi.fn());

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

vi.mock("./actions", () => ({
  signup: mockSignup,
}));

import { SignupForm } from "./signup-form";

describe("SignupForm", () => {
  it("サインアップフォームを表示する", () => {
    render(<SignupForm />);

    expect(screen.getByLabelText("メールアドレス")).toBeInTheDocument();
    expect(screen.getByLabelText("パスワード")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "サインアップ" })).toBeInTheDocument();
  });

  it("サインアップ失敗時にエラーメッセージを表示する", async () => {
    const user = userEvent.setup();
    mockSignup.mockResolvedValueOnce({
      success: false,
      error: "サインアップに失敗しました",
    });

    render(<SignupForm />);

    await user.type(screen.getByLabelText("メールアドレス"), "test@example.com");
    await user.type(screen.getByLabelText("パスワード"), "password123");
    await user.click(screen.getByRole("button", { name: "サインアップ" }));

    expect(await screen.findByText("サインアップに失敗しました")).toBeInTheDocument();
  });

  it("重複ユーザーの場合にエラーメッセージを表示する", async () => {
    const user = userEvent.setup();
    mockSignup.mockResolvedValueOnce({
      success: false,
      error: "既に登録されています",
    });

    render(<SignupForm />);

    await user.type(screen.getByLabelText("メールアドレス"), "test@example.com");
    await user.type(screen.getByLabelText("パスワード"), "password123");
    await user.click(screen.getByRole("button", { name: "サインアップ" }));

    expect(await screen.findByText("既に登録されています")).toBeInTheDocument();
  });

  it("サインアップ成功時にTODO一覧へ遷移する", async () => {
    const user = userEvent.setup();
    mockSignup.mockResolvedValueOnce({
      success: true,
      data: { id: 1, email: "test@example.com", role: "user" },
    });

    render(<SignupForm />);

    await user.type(screen.getByLabelText("メールアドレス"), "test@example.com");
    await user.type(screen.getByLabelText("パスワード"), "password123");
    await user.click(screen.getByRole("button", { name: "サインアップ" }));

    expect(mockPush).toHaveBeenCalledWith("/todos");
  });
});
