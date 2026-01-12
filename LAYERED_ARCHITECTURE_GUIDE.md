# レイヤードアーキテクチャガイド（PHP MVC経験者向け）

このドキュメントは、PHP MVCフレームワーク（Laravel、Symfonyなど）の経験者が、Goのレイヤードアーキテクチャを理解するためのガイドです。

## 目次
- [PHP MVCとの対応関係](#php-mvcとの対応関係)
- [Domain層とは何か？](#domain層とは何か)
- [MVCのModelとの決定的な違い](#mvcのmodelとの決定的な違い)
- [Domain層に切り出すべきもの](#domain層に切り出すべきもの)
- [なぜDomain層を分離するのか？](#なぜdomain層を分離するのか)
- [実践例：PHPとGoの比較](#実践例phpとgoの比較)

---

## PHP MVCとの対応関係

### 基本的な対応表

```
PHPのMVC          →  レイヤードアーキテクチャ
─────────────────────────────────────────────
Controller        →  Handler層
Service（※）      →  Service層
Model             →  Repository層 + Domain層（← ここがポイント！）
View              →  （フロントエンドで別途実装）
```

**※ Laravel/Symfonyなどで「Service」という概念を使っている場合**

### 視覚的な対応

```
┌─────────────────┐         ┌─────────────────┐
│   Controller    │         │   Handler層      │
├─────────────────┤         ├─────────────────┤
│   Service       │  ≒     │   Service層      │
├─────────────────┤         ├─────────────────┤
│                 │         │  Repository層    │
│     Model       │         ├─────────────────┤
│                 │         │   Domain層       │ ← 新しい概念！
└─────────────────┘         └─────────────────┘
   PHP MVC              レイヤードアーキテクチャ
```

---

## Domain層とは何か？

### 一言で言うと

**「データベースのことを一切知らない、純粋なビジネスの概念」**

### もう少し詳しく

Domain層は、以下の質問に答えるレイヤーです：

- **「このアプリケーションで扱う"もの"は何か？」** → エンティティ（Todo、Userなど）
- **「それらは何ができて、何ができないのか？」** → ビジネスルール
- **「どんな値が正しくて、どんな値が間違っているのか？」** → バリデーション
- **「どんなエラーが発生するのか？」** → ドメインエラー

**重要なのは、これらを定義する際に「データベース」「HTTP」「JSON」といった技術的な詳細を一切考えないこと**です。

---

## MVCのModelとの決定的な違い

### 問題：MVCのModelは責務が曖昧

PHPのMVCでは、Modelが「データベース操作」と「ビジネスの概念」の両方を担当することが多いです。

**LaravelのEloquent Modelの例:**

```php
class Todo extends Model
{
    protected $fillable = ['name', 'user_id', 'status'];

    // ========================================
    // ① データベース操作（Repository的な役割）
    // ========================================
    public static function findByUser($userId) {
        return self::where('user_id', $userId)->get();
    }

    public function scopeCompleted($query) {
        return $query->where('status', 'completed');
    }

    // ② リレーション定義（Repository的な役割）
    public function user() {
        return $this->belongsTo(User::class);
    }

    // ========================================
    // ③ ビジネスルール（Domain的な役割）
    // ========================================
    public function canBeCompleted() {
        return $this->status !== 'archived';
    }

    public function isOverdue() {
        return $this->due_date && $this->due_date->isPast();
    }

    // ④ バリデーション（Domain的な役割）
    public static function rules() {
        return [
            'name' => 'required|max:100',
            'status' => 'in:todo,in_progress,completed,archived',
        ];
    }

    // ⑤ アクセサ/ミューテータ（Domain的な役割）
    public function getFormattedDueDateAttribute() {
        return $this->due_date?->format('Y-m-d');
    }
}
```

この「①②データベース操作」と「③④⑤ビジネスの概念」が**混在している**のが問題です。

### 解決：Repository層とDomain層に分離

レイヤードアーキテクチャでは、これを明確に分離します。

**Domain層（ビジネスの概念のみ）:**

```go
// ビジネスの「概念」そのもの
// データベースのことは一切知らない
type Todo struct {
    ID      int
    Name    string
    UserID  int
    Status  string    // "todo", "in_progress", "completed", "archived"
    DueDate *time.Time
}

// ビジネスルール
func (t *Todo) CanBeCompleted() bool {
    return t.Status != "archived"
}

func (t *Todo) IsOverdue() bool {
    if t.DueDate == nil {
        return false
    }
    return t.DueDate.Before(time.Now())
}

// バリデーション
func (t *Todo) Validate() error {
    if len(t.Name) == 0 {
        return ErrInvalidName
    }
    if len(t.Name) > 100 {
        return ErrNameTooLong
    }
    validStatuses := []string{"todo", "in_progress", "completed", "archived"}
    if !contains(validStatuses, t.Status) {
        return ErrInvalidStatus
    }
    return nil
}
```

**Repository層（データベース操作のみ）:**

```go
// データベース操作のインターフェース
type TodoRepository interface {
    FindByUser(userID int) ([]Todo, error)
    FindCompleted(userID int) ([]Todo, error)
    Save(todo Todo) error
    Delete(todoID int) error
}

// PostgreSQL実装
type todoRepository struct {
    db *sql.DB
}

func (r *todoRepository) FindByUser(userID int) ([]Todo, error) {
    rows, err := r.db.Query(
        "SELECT id, name, user_id, status, due_date FROM todos WHERE user_id = $1",
        userID,
    )
    // ... SQL処理
}

func (r *todoRepository) FindCompleted(userID int) ([]Todo, error) {
    rows, err := r.db.Query(
        "SELECT id, name, user_id, status, due_date FROM todos WHERE user_id = $1 AND status = $2",
        userID, "completed",
    )
    // ... SQL処理
}
```

### 何が良くなったのか？

| 観点 | Eloquent Model | Repository + Domain |
|------|----------------|---------------------|
| **責務** | データベース操作 + ビジネス概念が混在 | 明確に分離 |
| **テスト** | データベースが必要（遅い） | Domain層はDB不要（速い） |
| **再利用** | Webアプリ専用 | Web、CLI、gRPCどこでも使える |
| **変更影響** | DB変更でビジネスルールも変更リスク | Repository層のみ変更 |
| **理解しやすさ** | 何でも詰め込まれて肥大化しがち | 各レイヤーの役割が明確 |

---

## Domain層に切り出すべきもの

### 1. エンティティ（ビジネスの核となる概念）

**「このアプリケーションで扱う"もの"は何か？」**

**Go（Domain層）:**
```go
type Todo struct {
    ID          int
    Name        string
    Description string
    UserID      int
    Status      string
    Priority    string
    DueDate     *time.Time
}

type User struct {
    ID           int
    Email        string
    PasswordHash string
    Role         string
    CreatedAt    time.Time
}
```

**PHP（POPO: Plain Old PHP Object として定義可能）:**
```php
class TodoEntity {
    public function __construct(
        public int $id,
        public string $name,
        public string $description,
        public int $userId,
        public string $status,
        public string $priority,
        public ?DateTimeImmutable $dueDate = null,
    ) {}
}
```

### 2. バリューオブジェクト（値そのものに意味がある概念）

例えば「メールアドレス」は、単なる文字列ではなく、ビジネス上の意味を持ちます。

**Go（Domain層）:**
```go
type Email struct {
    value string
}

func NewEmail(email string) (Email, error) {
    // メールアドレスのバリデーション
    if !strings.Contains(email, "@") {
        return Email{}, ErrInvalidEmail
    }
    if len(email) > 255 {
        return Email{}, ErrEmailTooLong
    }
    return Email{value: email}, nil
}

func (e Email) String() string {
    return e.value
}
```

**PHP:**
```php
class Email {
    private string $value;

    public function __construct(string $email) {
        if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
            throw new InvalidEmailException('Invalid email format');
        }
        if (strlen($email) > 255) {
            throw new EmailTooLongException('Email too long');
        }
        $this->value = $email;
    }

    public function toString(): string {
        return $this->value;
    }
}
```

### 3. ビジネスルール（「〜してよい/してはいけない」の判定）

**Go（Domain層）:**
```go
// TODOを完了できるか？
func (t *Todo) CanBeCompleted() bool {
    // アーカイブ済みのTODOは完了できない
    return t.Status != "archived"
}

// TODOを削除できるか？
func (t *Todo) CanBeDeleted(user User) bool {
    // 自分のTODOしか削除できない
    return t.UserID == user.ID
}

// 管理者か？
func (u *User) IsAdmin() bool {
    return u.Role == "admin"
}

// このTODOは高優先度か？
func (t *Todo) IsHighPriority() bool {
    return t.Priority == "high"
}
```

**PHP:**
```php
class TodoEntity {
    public function canBeCompleted(): bool {
        return $this->status !== 'archived';
    }

    public function canBeDeleted(UserEntity $user): bool {
        return $this->userId === $user->id;
    }

    public function isHighPriority(): bool {
        return $this->priority === 'high';
    }
}

class UserEntity {
    public function isAdmin(): bool {
        return $this->role === 'admin';
    }
}
```

### 4. バリデーションルール

**Go（Domain層）:**
```go
func (t *Todo) Validate() error {
    if len(t.Name) == 0 {
        return ErrTodoNameRequired
    }
    if len(t.Name) > 100 {
        return ErrTodoNameTooLong
    }
    validStatuses := []string{"todo", "in_progress", "completed", "archived"}
    if !contains(validStatuses, t.Status) {
        return ErrInvalidStatus
    }
    return nil
}
```

**PHP（FormRequestに相当）:**
```php
class TodoEntity {
    public function validate(): array {
        $errors = [];

        if (empty($this->name)) {
            $errors[] = 'Todo name is required';
        }
        if (strlen($this->name) > 100) {
            $errors[] = 'Todo name is too long';
        }
        if (!in_array($this->status, ['todo', 'in_progress', 'completed', 'archived'])) {
            $errors[] = 'Invalid status';
        }

        return $errors;
    }
}
```

### 5. ドメイン固有のエラー定義

**Go（Domain層）:**
```go
var (
    // リソース関連
    ErrNotFound     = errors.New("resource not found")
    ErrAlreadyExists = errors.New("resource already exists")

    // 認証・認可関連
    ErrUnauthorized = errors.New("unauthorized")
    ErrForbidden    = errors.New("forbidden")

    // バリデーション関連
    ErrTodoNameRequired = errors.New("todo name is required")
    ErrTodoNameTooLong  = errors.New("todo name is too long")
    ErrInvalidStatus    = errors.New("invalid status")
    ErrInvalidEmail     = errors.New("invalid email format")
)
```

**PHP（カスタム例外）:**
```php
// リソース関連
class ResourceNotFoundException extends DomainException {}
class ResourceAlreadyExistsException extends DomainException {}

// 認証・認可関連
class UnauthorizedException extends DomainException {}
class ForbiddenException extends DomainException {}

// バリデーション関連
class TodoNameRequiredException extends DomainException {}
class TodoNameTooLongException extends DomainException {}
class InvalidStatusException extends DomainException {}
class InvalidEmailException extends DomainException {}
```

### 6. 入力用のDTO（Data Transfer Object）

**Go（Domain層）:**
```go
type SignupInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type CreateTodoInput struct {
    Name        string  `json:"name" binding:"required,max=100"`
    Description string  `json:"description"`
    Priority    string  `json:"priority" binding:"oneof=low medium high"`
    DueDate     *string `json:"due_date"`
}
```

**PHP（FormRequest）:**
```php
class SignupRequest extends FormRequest {
    public function rules(): array {
        return [
            'email' => 'required|email',
            'password' => 'required|min:8',
        ];
    }
}

class CreateTodoRequest extends FormRequest {
    public function rules(): array {
        return [
            'name' => 'required|max:100',
            'description' => 'nullable|string',
            'priority' => 'required|in:low,medium,high',
            'due_date' => 'nullable|date',
        ];
    }
}
```

---

## Domain層に入れてはいけないもの

### ❌ NG例

**1. SQL文**
```go
// ❌ これはDomain層に書いてはいけない
func (t *Todo) Save() error {
    db.Exec("INSERT INTO todos (name, user_id) VALUES ($1, $2)", t.Name, t.UserID)
}
```
→ Repository層に書くべき

**2. データベース接続**
```go
// ❌ Domain層はDBを知ってはいけない
type Todo struct {
    ID   int
    Name string
    db   *sql.DB // ← NG！
}
```

**3. HTTPリクエスト/レスポンス**
```go
// ❌ Domain層はHTTPを知ってはいけない
func (t *Todo) HandleRequest(c *gin.Context) {
    c.JSON(200, t)
}
```
→ Handler層に書くべき

**4. 外部APIの呼び出し**
```go
// ❌ Domain層は外部依存を持ってはいけない
func (u *User) SendEmail() {
    http.Post("https://api.sendgrid.com/...", ...)
}
```
→ Service層やRepository層で外部サービスのインターフェースを定義

---

## なぜDomain層を分離するのか？

### 1. データベースに依存しない = 高速テスト

**Eloquent Modelのテスト（遅い）:**
```php
public function test_todo_validation() {
    // データベースが必要
    $todo = new Todo(['name' => '']);
    $todo->save(); // ← DB接続が必要

    $this->assertFalse($todo->isValid());
}
```

**Domain層のテスト（速い）:**
```go
func TestTodoValidation(t *testing.T) {
    // データベース不要！メモリ上で完結
    todo := Todo{Name: ""}
    err := todo.Validate()

    assert.Error(t, err) // 一瞬で完了
}
```

**速度の違い:**
- Eloquent Model: 0.5秒〜2秒（DBセットアップ含む）
- Domain層: 0.001秒未満

テストが1000個あれば、**数分 vs 数秒**の差になります。

### 2. ビジネスルールが一箇所に集約される

**悪い例（ビジネスルールが散らばる）:**

```php
// Controller
class TodoController {
    public function complete(Todo $todo) {
        if ($todo->status === 'archived') { // ← ビジネスルール①
            return response()->json(['error' => 'Cannot complete'], 400);
        }
        $todo->status = 'completed';
        $todo->save();
    }
}

// Service
class TodoService {
    public function archive(Todo $todo) {
        if ($todo->status === 'completed') { // ← ビジネスルール②
            $todo->status = 'archived';
            $todo->save();
        }
    }
}

// Blade View
@if($todo->status !== 'archived') {{-- ← ビジネスルール③ --}}
    <button>Complete</button>
@endif
```

同じルール「アーカイブ済みは完了できない」が3箇所に散らばっています。

**良い例（Domain層に集約）:**

```go
// Domain層
func (t *Todo) CanBeCompleted() bool {
    return t.Status != "archived" // ← ルールは1箇所のみ
}

// Handler層
func (h *Handler) CompleteTodo(c *gin.Context) {
    if !todo.CanBeCompleted() {
        c.JSON(400, gin.H{"error": "Cannot complete"})
        return
    }
    // ...
}

// Service層
func (s *Service) ArchiveTodo(todo Todo) error {
    if todo.Status != "completed" {
        return ErrCannotArchive
    }
    // ...
}

// Template（フロントエンド）
{{ if .Todo.CanBeCompleted }}
    <button>Complete</button>
{{ end }}
```

### 3. 技術的な詳細から独立している

Domain層は以下のことを**知りません**：
- ✅ データベースがPostgreSQLかMySQLか
- ✅ HTTPフレームワークがGinかEchoか
- ✅ JSONでデータをやり取りするのか、gRPCか
- ✅ ファイルに保存するのか、メモリに保存するのか

これにより：

**再利用性が高い:**
```go
// 同じDomain層を、Web API、CLI、バッチ処理で使い回せる

// Web API
func (h *Handler) CreateTodo(c *gin.Context) {
    todo := domain.Todo{Name: "Buy milk"}
    if err := todo.Validate(); err != nil {
        // ...
    }
}

// CLI
func main() {
    todo := domain.Todo{Name: "Buy milk"}
    if err := todo.Validate(); err != nil {
        fmt.Println(err)
    }
}

// バッチ処理
func ProcessTodos() {
    todos := []domain.Todo{...}
    for _, todo := range todos {
        if todo.IsOverdue() {
            // 通知を送る
        }
    }
}
```

**変更に強い:**
```go
// データベースをPostgreSQL → MySQLに変更
// → Repository層のみ修正、Domain層は無修正

// HTTPフレームワークをGin → Echoに変更
// → Handler層のみ修正、Domain層は無修正
```

### 4. ビジネスロジックの可視化

Domain層を見れば、「このアプリケーションが何をするものか」が一目でわかります。

```go
// domain/todo.go を見るだけで、TODOアプリの仕様がわかる
type Todo struct {
    ID          int
    Name        string
    Status      string    // "todo", "in_progress", "completed", "archived"
    Priority    string    // "low", "medium", "high"
    DueDate     *time.Time
}

func (t *Todo) CanBeCompleted() bool { ... }
func (t *Todo) CanBeArchived() bool { ... }
func (t *Todo) IsOverdue() bool { ... }
func (t *Todo) IsHighPriority() bool { ... }
```

一方、Eloquent Modelだとデータベーススキーマと混ざって見づらいです。

---

## 実践例：PHPとGoの比較

### シナリオ：「TODOを完了する」機能

#### Laravel（MVC）の実装

```php
// routes/web.php
Route::post('/todos/{todo}/complete', [TodoController::class, 'complete']);

// app/Http/Controllers/TodoController.php
class TodoController extends Controller
{
    public function complete(Todo $todo)
    {
        // ビジネスルール①: アーカイブ済みは完了できない
        if ($todo->status === 'archived') {
            return response()->json([
                'error' => 'Cannot complete archived todo'
            ], 400);
        }

        // ビジネスルール②: 自分のTODOしか完了できない
        if ($todo->user_id !== auth()->id()) {
            return response()->json([
                'error' => 'Forbidden'
            ], 403);
        }

        $todo->status = 'completed';
        $todo->completed_at = now();
        $todo->save();

        return response()->json($todo);
    }
}

// app/Models/Todo.php
class Todo extends Model
{
    protected $fillable = ['name', 'user_id', 'status', 'completed_at'];

    public function user()
    {
        return $this->belongsTo(User::class);
    }
}
```

**問題点:**
- ビジネスルールがControllerに書かれている（再利用できない）
- テストにはHTTPリクエストとDBが必要
- Modelにはビジネスロジックがほとんどない

#### Go（レイヤードアーキテクチャ）の実装

```go
// Domain層: ビジネスの概念
// domain/todo.go
type Todo struct {
    ID          int
    Name        string
    UserID      int
    Status      string
    CompletedAt *time.Time
}

// ビジネスルール
func (t *Todo) CanBeCompleted() bool {
    return t.Status != "archived"
}

func (t *Todo) CanBeAccessedBy(user User) bool {
    return t.UserID == user.ID
}

func (t *Todo) Complete() error {
    if !t.CanBeCompleted() {
        return ErrCannotComplete
    }
    t.Status = "completed"
    now := time.Now()
    t.CompletedAt = &now
    return nil
}

// Repository層: データ永続化
// repository/todo_repository.go
type TodoRepository interface {
    FindByID(id int) (domain.Todo, error)
    Update(todo domain.Todo) error
}

type todoRepository struct {
    db *sql.DB
}

func (r *todoRepository) Update(todo domain.Todo) error {
    _, err := r.db.Exec(
        "UPDATE todos SET status = $1, completed_at = $2 WHERE id = $3",
        todo.Status, todo.CompletedAt, todo.ID,
    )
    return err
}

// Service層: ビジネスロジックの組み合わせ
// service/todo_service.go
type TodoService struct {
    todoRepo repository.TodoRepository
}

func (s *TodoService) CompleteTodo(todoID int, user domain.User) (domain.Todo, error) {
    // 1. TODOを取得
    todo, err := s.todoRepo.FindByID(todoID)
    if err != nil {
        return domain.Todo{}, err
    }

    // 2. 権限チェック
    if !todo.CanBeAccessedBy(user) {
        return domain.Todo{}, domain.ErrForbidden
    }

    // 3. ビジネスロジック実行
    if err := todo.Complete(); err != nil {
        return domain.Todo{}, err
    }

    // 4. 永続化
    if err := s.todoRepo.Update(todo); err != nil {
        return domain.Todo{}, err
    }

    return todo, nil
}

// Handler層: HTTPリクエスト/レスポンス
// handler/todo_handler.go
func (h *TodoHandler) CompleteTodo(c *gin.Context) error {
    todoID, _ := strconv.Atoi(c.Param("id"))
    claims := c.MustGet("claims").(*AppClaims)
    userID, _ := strconv.Atoi(claims.Subject)
    user := domain.User{ID: userID}

    todo, err := h.todoService.CompleteTodo(todoID, user)
    if err != nil {
        return err // errorHandlerが適切なHTTPステータスに変換
    }

    c.JSON(http.StatusOK, todo)
    return nil
}
```

**改善点:**
- ビジネスルールがDomain層に集約（`CanBeCompleted()`, `Complete()`）
- 各レイヤーが明確な責務を持つ
- テストが容易：
  - Domain層: DBなしでテスト可能
  - Service層: Repository をモックしてテスト
  - Handler層: Service をモックしてテスト

---

## まとめ

### Domain層とは

**データベース、HTTP、JSONといった技術的な詳細を一切知らない、純粋なビジネスの概念**

### MVCのModelとの違い

| 観点 | MVCのModel | Domain層 |
|------|-----------|----------|
| 役割 | データベース操作 + ビジネス概念 | ビジネス概念のみ |
| 依存 | DB、ORM（Eloquent等） | 何にも依存しない |
| テスト | DB必要（遅い） | DB不要（速い） |
| 再利用 | Webアプリ専用 | どこでも使える |
| 責務 | 曖昧（何でも入る） | 明確（ビジネスのみ） |

### Domain層に切り出すべきもの

✅ エンティティ（Todo、User）
✅ バリューオブジェクト（Email、Password）
✅ ビジネスルール（「〜できる/できない」）
✅ バリデーション
✅ ドメイン固有のエラー
✅ 入力/出力DTO

### Domain層に入れてはいけないもの

❌ SQL文
❌ データベース接続
❌ HTTPリクエスト/レスポンス
❌ 外部API呼び出し

### なぜDomain層を分離するのか？

1. **高速テスト**: DB不要で一瞬でテスト完了
2. **ビジネスルールの集約**: ロジックが散らばらない
3. **技術的独立性**: DB/フレームワーク変更に強い
4. **再利用性**: Web、CLI、バッチで同じコードを使える
5. **可読性**: ビジネス要件が一目でわかる

---

## 参考：実際のディレクトリ構造

```
examples/
├── cmd/api/
│   └── main.go                    # エントリーポイント
├── internal/
│   ├── domain/                    # ★ Domain層（ビジネスの概念）
│   │   ├── todo.go               # TODOエンティティ + ビジネスルール
│   │   ├── user.go               # Userエンティティ + ビジネスルール
│   │   └── errors.go             # ドメインエラー定義
│   ├── repository/                # Repository層（DB操作）
│   │   ├── repository.go         # インターフェース定義
│   │   ├── todo_repository.go    # TODO用リポジトリ実装
│   │   └── user_repository.go    # User用リポジトリ実装
│   ├── service/                   # Service層（ビジネスロジック）
│   │   ├── auth_service.go       # 認証サービス
│   │   ├── todo_service.go       # TODOサービス
│   │   └── admin_service.go      # 管理者サービス
│   ├── handler/                   # Handler層（HTTP処理）
│   │   ├── user_handler.go       # ユーザー関連ハンドラ
│   │   ├── todo_handler.go       # TODO関連ハンドラ
│   │   └── admin_handler.go      # 管理者関連ハンドラ
│   ├── middleware/                # ミドルウェア
│   │   ├── auth.go
│   │   └── admin.go
│   └── config/                    # 設定管理
│       └── config.go
```

この構造により、**ビジネスロジックが`internal/domain/`に集約**され、他の層から独立して理解・テスト・変更できるようになります。
