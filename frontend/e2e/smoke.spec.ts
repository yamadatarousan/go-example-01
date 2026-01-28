import { test, expect } from "@playwright/test";

test("トップページが表示される", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "TODO App" })).toBeVisible();
});

test("ログイン後にTODO一覧へ遷移する", async ({ page }) => {
  await page.goto("/login");

  await page.getByLabel("メールアドレス").fill("user-test@example.com");
  await page.getByLabel("パスワード").fill("password123");
  await page.getByRole("button", { name: "ログイン" }).click();

  await expect(page).toHaveURL("/todos");
});

test("サインアップ→ログイン→TODO一覧へ遷移する", async ({ page }) => {
  const email = `test-${Date.now()}@example.com`;
  const password = "password123";

  await page.goto("/signup");

  await page.getByLabel("メールアドレス").fill(email);
  await page.getByLabel("パスワード").fill(password);
  await page.getByRole("button", { name: "サインアップ" }).click();

  // サインアップ後に自動ログインされ、/todosへ遷移する
  await expect(page).toHaveURL("/todos");
});

test("TODO一覧が表示される", async ({ page }) => {
  await page.goto("/todos");
  await expect(page.getByText("サンプルTODO 1")).toBeVisible();
  await expect(page.getByText("サンプルTODO 2")).toBeVisible();
});

test("TODOが0件のときに空状態を表示する", async ({ page }) => {
  await page.context().addCookies([
    { name: "mock_todos", value: "empty", url: "http://localhost:3010" },
  ]);

  await page.goto("/todos");

  await expect(page.getByText("TODOがまだありません")).toBeVisible();
});
