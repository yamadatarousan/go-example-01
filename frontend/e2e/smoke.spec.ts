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
