# レイヤードアーキテクチャガイド（PHP MVC経験者向け）

## PHP MVCとの対応

```
PHP MVC              →  レイヤードアーキテクチャ
────────────────────────────────────────────
Controller           →  Handler層
Service              →  Service層
Model                →  Repository層 + Domain層 ← ★ここが新しい
```

**ポイント: Modelが2つに分離される**

## Domain層とは？

**「データベース、HTTP、JSONを一切知らない、純粋なビジネスの概念」**

- このアプリで扱う"もの"は何か？ → **エンティティ**（Todo、User）
- それは何ができる/できない？ → **ビジネスルール**
- どんな値が正しい/間違い？ → **バリデーション**

## MVCのModelとの違い

### Laravel Eloquentの問題（責務が混在）

```php
class Todo extends Model {
    // ① DB操作（Repository的）
    public static function findByUser($userId) {
        return self::where('user_id', $userId)->get();
    }

    // ② ビジネスルール（Domain的）
    public function canBeCompleted() {
        return $this->status !== 'archived';
    }

    // ③ バリデーション（Domain的）
    public static function rules() {
        return ['name' => 'required|max:100'];
    }
}
```

**①と②③が混在 → テストが遅い、責務が不明確**

### レイヤードアーキテクチャの解決

```go
// Domain層: ビジネスの概念だけ（DB不要）
type Todo struct {
    ID     int
    Name   string
    Status string
}

func (t *Todo) CanBeCompleted() bool {
    return t.Status != "archived"
}

func (t *Todo) Validate() error {
    if len(t.Name) == 0 || len(t.Name) > 100 {
        return ErrInvalidName
    }
    return nil
}

// Repository層: DB操作だけ
type TodoRepository interface {
    FindByUser(userID int) ([]Todo, error)
    Save(todo Todo) error
}
```

## Domain層に何を入れるか

### ✅ 入れるもの

1. **エンティティ**: `type Todo struct { ... }`
2. **ビジネスルール**: `func (t *Todo) CanBeCompleted() bool`
3. **バリデーション**: `func (t *Todo) Validate() error`
4. **ドメインエラー**: `var ErrNotFound = errors.New(...)`
5. **入力DTO**: `type SignupInput struct { ... }`

### ❌ 入れてはいけないもの

- SQL文
- データベース接続（`*sql.DB`）
- HTTPリクエスト/レスポンス（`gin.Context`）
- 外部API呼び出し

## なぜ分離するのか

### 1. 高速テスト

```go
// DB不要、一瞬で完了
func TestTodoValidation(t *testing.T) {
    todo := Todo{Name: ""}
    err := todo.Validate()
    assert.Error(t, err)
}
```

**Eloquent: 0.5〜2秒 vs Domain: 0.001秒未満**

### 2. ビジネスルールが一箇所に集約

```php
// 悪い例: ルールが散らばる
if ($todo->status === 'archived') { ... } // Controller
if ($todo->status === 'archived') { ... } // Service
@if($todo->status !== 'archived') // View
```

```go
// 良い例: 1箇所に集約
func (t *Todo) CanBeCompleted() bool {
    return t.Status != "archived" // ← ここだけ
}
```

### 3. 技術から独立

**Domain層は知らない:**
- PostgreSQLかMySQLか
- GinかEchoか
- JSON か gRPC か

**効果:**
- Web API、CLI、バッチで同じコードを使える
- DB変更してもDomain層は無修正

## まとめ表

| 観点 | Eloquent Model | Domain層 |
|------|----------------|----------|
| 役割 | DB操作 + ビジネス概念 | ビジネス概念のみ |
| 依存 | DB、ORM | なし |
| テスト | DB必要（遅い） | DB不要（速い） |
| 再利用 | Web専用 | どこでも |

**Domain層 = MVCのModelから「データベース操作」を除いた「ビジネスの概念」だけを抽出したもの**
