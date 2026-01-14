# 完了タスク

## Phase 1: コード構造のリファクタリング（完了: 2026-01-12）

### 実施内容

**1.1 ディレクトリ構造の再編成**
- ✅ レイヤードアーキテクチャの導入
- ✅ examples/ディレクトリ配下に新しい構造を作成

```
examples/
├── cmd/
│   └── api/
│       └── main.go           # エントリーポイント
├── internal/
│   ├── domain/               # ドメインモデル
│   │   ├── user.go
│   │   ├── todo.go
│   │   └── errors.go
│   ├── handler/              # HTTPハンドラー
│   │   ├── error_handler.go
│   │   ├── user_handler.go
│   │   ├── todo_handler.go
│   │   └── admin_handler.go
│   ├── repository/           # データアクセス
│   │   ├── repository.go     # インターフェース定義
│   │   ├── user_repository.go
│   │   └── todo_repository.go
│   ├── service/              # ビジネスロジック
│   │   ├── auth_service.go
│   │   ├── todo_service.go
│   │   └── admin_service.go
│   ├── middleware/           # ミドルウェア
│   │   ├── auth.go
│   │   ├── admin.go
│   │   └── request_id.go
│   └── config/               # 設定管理
│       └── config.go
└── pkg/                      # (今回は省略)
```

**1.2 レイヤー分離の実装**
- ✅ **Domain層**: エンティティ（Todo, User）とビジネスルール（エラー定義）を実装
- ✅ **Repository層**: データベース操作の抽象化とインターフェース定義
- ✅ **Service層**: ビジネスロジックの集約（認証、TODO管理、管理者機能）
- ✅ **Handler層**: HTTPリクエスト/レスポンス処理とエラーハンドリング
- ✅ **Middleware層**: 認証、管理者権限チェック、リクエストID生成
- ✅ **インターフェースによる依存性注入**: 各レイヤーは上位レイヤーに依存せず、インターフェースを通じて疎結合

**1.3 設定管理の改善**
- ✅ 環境変数の一元管理（config.Config構造体）
- ✅ データベース、サーバー、JWT設定を分離
- ✅ デフォルト値のサポート

### 主な改善点

1. **保守性の向上**
   - 単一ファイル（main.go 516行）から複数ファイルに分割
   - 各レイヤーの責務が明確化
   - テストしやすい構造

2. **依存性の明確化**
   - Repository → Service → Handler の明確な依存関係
   - インターフェースによる抽象化

3. **エラーハンドリングの改善**
   - ドメイン固有のエラー定義（domain/errors.go）
   - 一元化されたエラーハンドラー（handler/error_handler.go）

4. **設定の一元管理**
   - 環境変数の管理がconfig.Loadで一箇所に集約
   - DSN生成などのロジックも設定層に集約

### ファイル一覧

#### Domain層（3ファイル）
- `internal/domain/todo.go`: Todoエンティティ
- `internal/domain/user.go`: User エンティティ、SignupInput、LoginInput
- `internal/domain/errors.go`: ドメイン固有のエラー定義

#### Repository層（3ファイル）
- `internal/repository/repository.go`: TodoRepository、UserRepositoryインターフェース
- `internal/repository/todo_repository.go`: TODO関連のDB操作（トランザクション対応）
- `internal/repository/user_repository.go`: ユーザー関連のDB操作

#### Service層（3ファイル）
- `internal/service/auth_service.go`: 認証ロジック（JWT生成、パスワード検証）
- `internal/service/todo_service.go`: TODO管理ロジック
- `internal/service/admin_service.go`: 管理者機能ロジック

#### Handler層（4ファイル）
- `internal/handler/error_handler.go`: エラーハンドリングの統一
- `internal/handler/user_handler.go`: ユーザー登録・ログインハンドラー
- `internal/handler/todo_handler.go`: TODO CRUD操作ハンドラー
- `internal/handler/admin_handler.go`: 管理者機能ハンドラー

#### Middleware層（3ファイル）
- `internal/middleware/auth.go`: JWT認証ミドルウェア
- `internal/middleware/admin.go`: 管理者権限チェックミドルウェア
- `internal/middleware/request_id.go`: リクエストID生成ミドルウェア

#### Config層（1ファイル）
- `internal/config/config.go`: 環境変数管理と設定読み込み

#### エントリーポイント（1ファイル）
- `cmd/api/main.go`: アプリケーションのエントリーポイント（DI、ルーティング、Graceful Shutdown）

**合計: 18ファイル**

### 技術的なハイライト

1. **依存性注入（DI）パターン**
   ```go
   // Repository層
   todoRepo := repository.NewTodoRepository(db)
   userRepo := repository.NewUserRepository(db)

   // Service層（Repositoryを注入）
   authService := service.NewAuthService(userRepo, cfg.JWT.Secret)
   todoService := service.NewTodoService(todoRepo)

   // Handler層（Serviceを注入）
   userHandler := handler.NewUserHandler(authService)
   todoHandler := handler.NewTodoHandler(todoService)
   ```

2. **インターフェースベースの設計**
   ```go
   type TodoRepository interface {
       FindAll(userID int) ([]domain.Todo, error)
       FindByID(todoID, userID int) (domain.Todo, error)
       CreateTodoWithAudit(ctx context.Context, todo domain.Todo) (domain.Todo, error)
       // ...
   }
   ```

3. **ドメイン駆動エラー**
   ```go
   var (
       ErrNotFound = errors.New("resource not found")
       ErrUnauthorized = errors.New("unauthorized")
       ErrForbidden = errors.New("forbidden")
       // ...
   )
   ```

---

## Phase 2: TODO機能の拡張（完了: 2026-01-13）

### 実施内容

**2.1 データベースマイグレーション**
- ✅ TODOテーブルの拡張（優先度、期限、ステータス、説明、サブタスク）
- ✅ カテゴリーテーブルの作成
- ✅ タグテーブルと中間テーブルの作成（Many-to-Many）
- ✅ パフォーマンス向上のためのインデックス追加

**マイグレーションファイル（6ファイル）**
- `db/migrations/000008_add_todo_extensions.up/down.sql`
- `db/migrations/000009_create_categories_table.up/down.sql`
- `db/migrations/000010_create_tags_table.up/down.sql`

**2.2 ドメインモデルの拡張**
- ✅ Todoエンティティの拡張（Priority, Status, DueDate, Description, CategoryID, ParentTodoID）
- ✅ Categoryエンティティの追加
- ✅ Tagエンティティの追加
- ✅ CreateTodoInput / UpdateTodoInput の追加

**2.3 Repository層の拡張**
- ✅ TodoRepositoryに新しいメソッド追加
  - `UpdateStatus`: ステータス更新
  - `FindOverdue`: 期限切れTODO取得
  - `FindToday`: 今日が期限のTODO取得
  - `FindThisWeek`: 今週が期限のTODO取得
- ✅ CategoryRepository の実装（CRUD操作）
- ✅ TagRepository の実装（作成・紐付け・検索）

**2.4 Service層の拡張**
- ✅ CategoryService の実装
- ✅ TodoServiceに新しいメソッド追加
  - `CompleteTodo`: TODO完了
  - `ReopenTodo`: TODO再開
  - `GetOverdueTodos`: 期限切れTODO取得
  - `GetTodayTodos`: 今日のTODO取得
  - `GetThisWeekTodos`: 今週のTODO取得

**2.5 Handler層の拡張**
- ✅ CategoryHandler の実装（カテゴリーCRUD）
- ✅ TodoHandlerに新しいエンドポイント追加
  - `POST /api/v1/todos/:id/complete`: TODO完了
  - `POST /api/v1/todos/:id/reopen`: TODO再開
  - `GET /api/v1/todos/overdue`: 期限切れTODO一覧
  - `GET /api/v1/todos/today`: 今日のTODO一覧
  - `GET /api/v1/todos/week`: 今週のTODO一覧

**2.6 カテゴリーエンドポイント**
- ✅ `POST /api/v1/categories`: カテゴリー作成
- ✅ `GET /api/v1/categories`: カテゴリー一覧
- ✅ `GET /api/v1/categories/:id`: カテゴリー取得
- ✅ `PUT /api/v1/categories/:id`: カテゴリー更新
- ✅ `DELETE /api/v1/categories/:id`: カテゴリー削除

### 追加されたファイル

#### Domain層（+2ファイル）
- `internal/domain/category.go`: Categoryエンティティ
- `internal/domain/tag.go`: Tagエンティティ
- `internal/domain/todo.go`: 拡張（Priority, Status, DueDate等を追加）

#### Repository層（+2ファイル）
- `internal/repository/category_repository.go`: カテゴリーCRUD実装
- `internal/repository/tag_repository.go`: タグ管理実装
- `internal/repository/repository.go`: インターフェース拡張
- `internal/repository/todo_repository.go`: 新メソッド追加

#### Service層（+1ファイル）
- `internal/service/category_service.go`: カテゴリー管理ビジネスロジック
- `internal/service/todo_service.go`: 新メソッド追加

#### Handler層（+1ファイル）
- `internal/handler/category_handler.go`: カテゴリーHTTPハンドラー
- `internal/handler/todo_handler.go`: 新エンドポイント追加

#### エントリーポイント
- `cmd/api/main.go`: 新しい依存性注入とルーティング追加

**Phase 2で追加されたファイル数: 6ファイル**
**Phase 2で修正されたファイル数: 5ファイル**

### 主な改善点

1. **TODO管理の実用性向上**
   - 優先度（high/medium/low）でタスクの重要度を管理
   - 期限（due_date）でデッドラインを設定
   - ステータス（todo/in_progress/done）で進捗を管理
   - 詳細説明（description）で詳細情報を記録
   - サブタスク（parent_todo_id）で階層的なタスク管理

2. **カテゴリー機能**
   - TODOをカテゴリー別に分類可能
   - カラーコード付きで視覚的に管理
   - ユーザーごとに独立したカテゴリー

3. **タグ機能（Phase 3で活用予定）**
   - Many-to-Many関係でTODOに複数タグを付与可能
   - タグによる横断的な分類

4. **期限管理機能**
   - 期限切れTODOの一覧取得
   - 今日・今週のTODO取得
   - タスクの優先度付け

### データベーススキーマの変更

**TODOテーブルの拡張**
```sql
ALTER TABLE todos ADD COLUMN priority VARCHAR(20) DEFAULT 'medium';
ALTER TABLE todos ADD COLUMN due_date TIMESTAMPTZ;
ALTER TABLE todos ADD COLUMN status VARCHAR(20) DEFAULT 'todo';
ALTER TABLE todos ADD COLUMN description TEXT;
ALTER TABLE todos ADD COLUMN parent_todo_id INT REFERENCES todos(id);
ALTER TABLE todos ADD COLUMN category_id INT REFERENCES categories(id);
```

**新規テーブル**
- `categories`: カテゴリー管理（id, name, color, user_id）
- `tags`: タグ管理（id, name）
- `todo_tags`: TODO-タグ中間テーブル（todo_id, tag_id）

**インデックス追加**
```sql
CREATE INDEX idx_todos_status ON todos(status);
CREATE INDEX idx_todos_priority ON todos(priority);
CREATE INDEX idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL;
CREATE INDEX idx_todos_parent ON todos(parent_todo_id) WHERE parent_todo_id IS NOT NULL;
CREATE INDEX idx_todos_category ON todos(category_id) WHERE category_id IS NOT NULL;
```

### 技術的なハイライト

1. **条件付きインデックス（Partial Index）**
   ```sql
   CREATE INDEX idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL;
   ```
   NULL値を除外してインデックスサイズを削減

2. **Many-to-Many関係の実装**
   ```sql
   CREATE TABLE todo_tags (
       todo_id INT REFERENCES todos(id) ON DELETE CASCADE,
       tag_id INT REFERENCES tags(id) ON DELETE CASCADE,
       PRIMARY KEY (todo_id, tag_id)
   );
   ```

3. **日付範囲クエリの最適化**
   ```go
   // 今週のTODO取得
   query := `
       SELECT ...
       FROM todos
       WHERE user_id = $1
         AND due_date >= CURRENT_DATE
         AND due_date < CURRENT_DATE + INTERVAL '7 days'
         AND status != 'done'
   `
   ```

4. **NULL許容フィールドの適切な型使用**
   ```go
   type Todo struct {
       Description  *string    `json:"description"`
       DueDate      *time.Time `json:"due_date"`
       CategoryID   *int       `json:"category_id"`
       ParentTodoID *int       `json:"parent_todo_id"`
   }
   ```

### context.Contextの全面適用（追加修正）

**実施日**: 2026-01-13

**背景**: 当初、書き込み系のみにcontextを使用していたが、読み取り系にもcontextを使うべきというフィードバックを受けて、全DB操作にcontextを適用。

**修正内容**:
- ✅ 全Repository層のインターフェースにcontext追加
- ✅ 全Repository層の実装で `db.Query` → `db.QueryContext` に変更
- ✅ 全Repository層の実装で `db.QueryRow` → `db.QueryRowContext` に変更
- ✅ 全Repository層の実装で `db.Exec` → `db.ExecContext` に変更
- ✅ 全Service層のメソッドにcontext引数を追加
- ✅ 全Handler層で `c.Request.Context()` をServiceに渡すように修正

**修正したメソッド**:
- `TodoRepository`: FindAll, FindByID, Create, FindOverdue, FindToday, FindThisWeek
- `UserRepository`: CreateUser, FindUserByEmail, FindAllUsers
- `CategoryRepository`: FindAll, FindByID
- `TagRepository`: FindAll, FindByTodoID
- `TodoService`: GetTodos, GetTodo, GetOverdueTodos, GetTodayTodos, GetThisWeekTodos
- `AuthService`: Signup, Login
- `AdminService`: GetAllUsers
- `CategoryService`: GetCategories, GetCategory

**効果**:
1. **タイムアウト制御**: 長時間かかるクエリを自動でキャンセル可能
2. **キャンセル伝播**: HTTPリクエストがキャンセルされたら、DB操作も即座に中断
3. **リクエストスコープの情報伝播**: トレーシングID、ユーザー情報などをcontextで伝播可能
4. **一貫性**: すべてのDB操作で統一的なAPIを使用

---

### 重要な注意事項

すべてのファイルに以下のコメントを追加：
```go
// ★★★ 重要: プロジェクトディレクトリ直下に写経する際は、
// import pathから "examples/" を削除してください ★★★
```

---

## Phase 2.5: 全エンドポイントの統合テスト（完了: 2026-01-14）

### 実施内容

**背景**: Phase 3に進む前に、Phase 1およびPhase 2で実装された全エンドポイントが正しく動作することを確認するため、統合テストを作成しました。

**2.5.1 統合テストの実装**
- ✅ テストフレームワークの構築（TestMain、setupTestRouter）
- ✅ 全21個のエンドポイントのテスト実装
- ✅ テスト用DBのセットアップ（docker-compose.test.yml使用）
- ✅ マイグレーション自動実行
- ✅ シードデータのロード

**テスト対象エンドポイント**:

1. **ユーザー認証** (3テスト)
   - TestSignup: ユーザー登録
   - TestLoginSuccess: ログイン成功
   - TestLoginFailure: ログイン失敗

2. **TODO基本CRUD** (6テスト)
   - TestGetTodos: TODO一覧取得
   - TestGetTodo: TODO詳細取得
   - TestGetTodoNotFound: 存在しないTODO取得（404）
   - TestCreateTodoWithExtendedFields: 拡張フィールド含むTODO作成
   - TestUpdateTodo: TODO更新
   - TestDeleteTodo: TODO削除

3. **TODO拡張機能**（Phase 2追加、5テスト）
   - TestCompleteTodo: TODO完了
   - TestReopenTodo: TODO再開
   - TestGetOverdueTodos: 期限切れTODO一覧
   - TestGetTodayTodos: 今日のTODO一覧
   - TestGetThisWeekTodos: 今週のTODO一覧

4. **カテゴリーCRUD**（Phase 2追加、5テスト）
   - TestCreateCategory: カテゴリー作成
   - TestGetCategories: カテゴリー一覧取得
   - TestGetCategory: カテゴリー詳細取得
   - TestUpdateCategory: カテゴリー更新
   - TestDeleteCategory: カテゴリー削除

5. **管理者機能** (2テスト)
   - TestGetAllUsersAsAdmin: 管理者による全ユーザー取得
   - TestGetAllUsersAsForbidden: 一般ユーザーによるアクセス拒否（403）

**合計: 21テスト、全てPASS**

**2.5.2 実装中に発見・修正した問題**

1. **created_at/updated_atカラムの不足**
   - 問題: TODOsテーブルにタイムスタンプカラムが存在しなかった
   - 修正: マイグレーション000011を追加してカラムを作成

2. **category_handler.goの認証エラー**
   - 問題: `c.Get("user_id")`を使用していたが、ミドルウェアは`claims`をセット
   - 修正: 他のハンドラーと同様に`c.MustGet("claims").(*service.AppClaims)`パターンに変更

3. **Complete/Reopenエンドポイントのレスポンス**
   - 問題: メッセージのみ返していたため、テストでstatusフィールドを確認できない
   - 修正: 更新後のTODOオブジェクト全体を返すように変更

4. **TODO作成時の拡張フィールド**
   - 問題: INSERTで`name`と`user_id`のみ挿入していた
   - 修正: Phase 2フィールド（priority, status, description, due_date, category_id, parent_todo_id）も挿入するように変更

5. **TODO取得時の拡張フィールド**
   - 問題: FindByID/FindAllで`id, name, user_id`のみ取得していた
   - 修正: Phase 2フィールドと`created_at, updated_at`も取得するように変更

**2.5.3 作成ファイル**
- `examples/cmd/api/main_test.go`: 全エンドポイントの統合テスト（約680行）
- `db/migrations/000011_add_timestamps_to_todos.up.sql`: タイムスタンプカラム追加
- `db/migrations/000011_add_timestamps_to_todos.down.sql`: ロールバック用

**2.5.4 修正ファイル**
- `examples/internal/handler/category_handler.go`: 認証方法をclaims取得に統一
- `examples/internal/handler/todo_handler.go`: Complete/Reopenレスポンスを更新
- `examples/internal/repository/todo_repository.go`: Create/FindByID/FindAllをPhase 2フィールド対応

### 主な改善点

1. **テストの自動化**
   - 手動での動作確認が不要になった
   - 全エンドポイントの動作を数秒で検証可能
   - リグレッション（機能の退行）を防ぐ

2. **バグの早期発見**
   - 実装中に5つの問題を発見・修正
   - Phase 2で追加したフィールドが正しく保存・取得されることを確認

3. **コードの品質向上**
   - データベース操作の一貫性を確保（全操作でPhase 2フィールドを扱う）
   - エラーハンドリングの統一（401, 403, 404のテスト）

4. **開発速度の向上**
   - 新機能追加時にテストを追加することで、既存機能への影響を即座に確認可能
   - テストがドキュメントの役割も果たす（エンドポイントの使い方が明確）

### テスト実行方法

```bash
# examples/cmd/api ディレクトリで実行
cd examples/cmd/api
go test -v -timeout 5m

# 特定のテストのみ実行
go test -v -run TestComplete
go test -v -run "TestGet.*Todos"
```

### 次のステップ

Phase 3: 検索・フィルタリング機能の実装準備が整いました。
- クエリパラメータによる高度な検索
- 全文検索の実装
- ページネーション
- 統計情報API
