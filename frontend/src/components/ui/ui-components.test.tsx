import { render, screen } from "@testing-library/react";

import { Button } from "./button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./card";
import { Input } from "./input";
import { Label } from "./label";

describe("UIコンポーネント", () => {
  it("Buttonが表示される", () => {
    render(<Button>送信</Button>);
    expect(screen.getByRole("button", { name: "送信" })).toBeInTheDocument();
  });

  it("Inputが表示される", () => {
    render(<Input placeholder="入力してください" />);
    expect(screen.getByPlaceholderText("入力してください")).toBeInTheDocument();
  });

  it("LabelがInputと紐付く", () => {
    render(
      <div>
        <Label htmlFor="email">メールアドレス</Label>
        <Input id="email" />
      </div>
    );

    expect(screen.getByLabelText("メールアドレス")).toBeInTheDocument();
  });

  it("Cardが表示される", () => {
    render(
      <Card>
        <CardHeader>
          <CardTitle>タイトル</CardTitle>
          <CardDescription>説明文</CardDescription>
        </CardHeader>
        <CardContent>本文</CardContent>
        <CardFooter>フッター</CardFooter>
      </Card>
    );

    expect(screen.getByText("タイトル")).toBeInTheDocument();
    expect(screen.getByText("説明文")).toBeInTheDocument();
    expect(screen.getByText("本文")).toBeInTheDocument();
    expect(screen.getByText("フッター")).toBeInTheDocument();
  });
});
