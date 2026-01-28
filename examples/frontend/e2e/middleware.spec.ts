import { test, expect } from "@playwright/test";

test.beforeEach(async ({ context }) => {
  // テスト間のCookieをクリアして独立させる
  await context.clearCookies();
});

test("未ログインで /todos にアクセスすると /login にリダイレクトされる", async ({ page }) => {
  await page.goto("/todos");
  await expect(page).toHaveURL("/login");
});

test("ログイン済みで /login にアクセスすると /todos にリダイレクトされる", async ({ page }) => {
  await page.context().addCookies([
    { name: "token", value: "dummy-token", url: "http://localhost:3010" },
  ]);

  await page.goto("/login");
  await expect(page).toHaveURL("/todos");
});
