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

---

## Phase 3: 検索・フィルタリング機能（完了: 2026-01-15）

### 実施内容

**3.1 全文検索機能の実装**
- ✅ PostgreSQLのGINインデックスを使用した全文検索
- ✅ tsvectorカラムとトリガー関数の追加
- ✅ 名前と説明文に対する全文検索

**マイグレーションファイル**
- `db/migrations/000012_add_fulltext_search_index.up.sql`: GINインデックスとトリガー追加
- `db/migrations/000012_add_fulltext_search_index.down.sql`: ロールバック用

**3.2 高度な検索・フィルタリングAPI**
- ✅ Repository層に`Search`メソッド実装
- ✅ Service層に`SearchTodos`メソッド実装
- ✅ Handler層に`SearchTodos`ハンドラー実装
- ✅ ルーティングに`GET /api/v1/todos/search`追加

**検索パラメータ**:
- `status`: ステータスフィルター（todo/in_progress/done）
- `priority`: 優先度フィルター（low/medium/high）
- `category_id`: カテゴリーIDフィルター
- `tag_ids`: タグIDフィルター（配列）
- `search`: 全文検索キーワード
- `due_from`: 期限開始日
- `due_to`: 期限終了日
- `sort`: ソート項目（due_date/priority/created_at/updated_at）
- `order`: ソート順（asc/desc）
- `page`: ページ番号
- `limit`: 1ページあたりの件数

**レスポンス形式**:
```json
{
  "todos": [...],
  "total": 100,
  "page": 1,
  "limit": 10,
  "total_pages": 10
}
```

**3.3 統計情報API**
- ✅ Repository層に`GetStatistics`メソッド実装
- ✅ Service層に`GetStatistics`メソッド実装
- ✅ Handler層に`GetStatistics`ハンドラー実装
- ✅ ルーティングに`GET /api/v1/todos/statistics`追加

**統計情報内容**:
```json
{
  "total_count": 100,
  "status_counts": {
    "todo": 30,
    "in_progress": 20,
    "done": 50
  },
  "priority_counts": {
    "low": 30,
    "medium": 40,
    "high": 30
  },
  "overdue_count": 10,
  "due_today_count": 5,
  "due_this_week_count": 15
}
```

**3.4 Domain層の拡張**
- ✅ `SearchFilters`型を追加（検索条件を表現）
- ✅ `SearchResult`型を追加（検索結果とページネーション情報）
- ✅ `Statistics`型を追加（統計情報）

**3.5 統合テストの追加**
- ✅ TestSearchTodos: 検索機能のテスト（5つのサブテスト）
  - Filter by priority: 優先度フィルター
  - Filter by status: ステータスフィルター
  - Full-text search: 全文検索
  - Pagination: ページネーション
  - Sort order: ソート順
- ✅ TestGetStatistics: 統計情報のテスト

**合計: 23テスト、全てPASS**

### 追加されたファイル

#### マイグレーション（+2ファイル）
- `db/migrations/000012_add_fulltext_search_index.up.sql`
- `db/migrations/000012_add_fulltext_search_index.down.sql`

#### Domain層の拡張
- `internal/domain/todo.go`: SearchFilters, SearchResult, Statistics型を追加

#### Repository層の拡張
- `internal/repository/repository.go`: TodoRepositoryインターフェースにSearch, GetStatisticsを追加
- `internal/repository/todo_repository.go`: Search, GetStatisticsメソッドを実装

#### Service層の拡張
- `internal/service/todo_service.go`: SearchTodos, GetStatisticsメソッドを追加

#### Handler層の拡張
- `internal/handler/todo_handler.go`: SearchTodos, GetStatisticsハンドラーを追加

#### エントリーポイント
- `cmd/api/main.go`: 新しいルーティング追加

#### テスト
- `cmd/api/main_test.go`: TestSearchTodos, TestGetStatisticsを追加

**Phase 3で追加されたコード行数: 約350行**

### 主な改善点

1. **高度な検索機能**
   - 複数のフィルター条件を組み合わせた検索が可能
   - ステータス、優先度、カテゴリー、タグ、期限範囲による絞り込み
   - 動的SQLクエリビルダーによる柔軟な条件構築

2. **全文検索の実装**
   - PostgreSQLのGINインデックスを活用した高速な全文検索
   - 名前と説明文を対象とした検索（重み付けA/B）
   - トリガー関数による自動インデックス更新

3. **ページネーション対応**
   - 大量データの効率的な表示
   - ページ番号、件数、総ページ数の情報提供
   - OFFSETとLIMITによるページング実装

4. **ソート機能**
   - 複数のソート項目（due_date, priority, created_at, updated_at）
   - 昇順・降順の指定
   - デフォルト値（created_at DESC）

5. **統計情報の提供**
   - ダッシュボード作成に必要な集計情報
   - ステータス別・優先度別のカウント
   - 期限関連の集計（期限切れ、今日期限、今週期限）

### 技術的なハイライト

1. **PostgreSQL GINインデックスによる全文検索**
   ```sql
   -- tsvector列とトリガー関数
   ALTER TABLE todos ADD COLUMN search_vector tsvector;

   CREATE OR REPLACE FUNCTION todos_search_vector_update() RETURNS trigger AS $$
   BEGIN
     NEW.search_vector :=
       setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
       setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B');
     RETURN NEW;
   END;
   $$ LANGUAGE plpgsql;

   CREATE TRIGGER todos_search_vector_trigger
   BEFORE INSERT OR UPDATE ON todos
   FOR EACH ROW
   EXECUTE FUNCTION todos_search_vector_update();

   CREATE INDEX idx_todos_search_vector ON todos USING GIN(search_vector);
   ```

2. **動的SQLクエリビルディング**
   ```go
   func (r *todoRepository) Search(ctx context.Context, userID int, filters domain.SearchFilters) (domain.SearchResult, error) {
       // 条件を動的に構築
       conditions := []string{"t.user_id = $1"}
       args := []interface{}{userID}
       argCount := 1

       if filters.Status != nil {
           argCount++
           conditions = append(conditions, fmt.Sprintf("t.status = $%d", argCount))
           args = append(args, *filters.Status)
       }

       if filters.Search != "" {
           argCount++
           conditions = append(conditions, fmt.Sprintf("t.search_vector @@ plainto_tsquery('english', $%d)", argCount))
           args = append(args, filters.Search)
       }
       // ... 他のフィルター条件
   }
   ```

3. **集計クエリのパフォーマンス最適化**
   ```go
   // 1つのクエリで複数のカウントを取得
   statusQuery := `
       SELECT
           COUNT(*) as total,
           COUNT(CASE WHEN status = 'todo' THEN 1 END) as todo_count,
           COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress_count,
           COUNT(CASE WHEN status = 'done' THEN 1 END) as done_count
       FROM todos
       WHERE user_id = $1
   `
   ```

4. **ページネーション計算**
   ```go
   offset := (filters.Page - 1) * filters.Limit
   totalPages := (total + filters.Limit - 1) / filters.Limit

   return domain.SearchResult{
       Todos:      todos,
       Total:      total,
       Page:       filters.Page,
       Limit:      filters.Limit,
       TotalPages: totalPages,
   }, nil
   ```

### API使用例

**検索API**:
```bash
# 優先度がhighで未完了のTODOを検索
GET /api/v1/todos/search?priority=high&status=todo

# 全文検索でキーワードを含むTODOを検索
GET /api/v1/todos/search?search=meeting

# 複数条件での検索＋ページネーション
GET /api/v1/todos/search?status=todo&priority=high&page=1&limit=10&sort=due_date&order=asc

# 期限範囲での検索
GET /api/v1/todos/search?due_from=2026-01-01&due_to=2026-12-31

# カテゴリーでの絞り込み
GET /api/v1/todos/search?category_id=1
```

**統計情報API**:
```bash
# 統計情報を取得
GET /api/v1/todos/statistics
```

### テスト結果

```bash
=== RUN   TestSearchTodos
=== RUN   TestSearchTodos/Filter_by_priority
=== RUN   TestSearchTodos/Filter_by_status
=== RUN   TestSearchTodos/Full-text_search
=== RUN   TestSearchTodos/Pagination
=== RUN   TestSearchTodos/Sort_order
--- PASS: TestSearchTodos (0.15s)
    --- PASS: TestSearchTodos/Filter_by_priority (0.00s)
    --- PASS: TestSearchTodos/Filter_by_status (0.00s)
    --- PASS: TestSearchTodos/Full-text_search (0.01s)
    --- PASS: TestSearchTodos/Pagination (0.00s)
    --- PASS: TestSearchTodos/Sort_order (0.00s)
=== RUN   TestGetStatistics
--- PASS: TestGetStatistics (0.20s)
PASS
ok  	gin-quickstart/examples/cmd/api	6.485s
```

**全23テスト、全てPASS**

### 次のステップ

Phase 3が完了しました。以下の機能が実装され、完全に動作しています：

1. ✅ **Phase 1**: レイヤードアーキテクチャへのリファクタリング
2. ✅ **Phase 2**: TODO機能の拡張（優先度、期限、ステータス、カテゴリー）
3. ✅ **Phase 3**: 検索・フィルタリング・統計情報

今後の拡張可能性：
- タグ機能の完全実装（Phase 2で準備済み）
- サブタスク機能のUI対応
- リアルタイム通知機能
- ファイル添付機能
- コメント機能

---

## Phase 4: 通知・リマインダー機能（完了: 2026-01-15）

### 実施内容

**4.1 データベースマイグレーション**
- ✅ 通知テーブルの作成（notifications）
- ✅ リマインダーテーブルの作成（reminders）
- ✅ パフォーマンス向上のためのインデックス追加

**マイグレーションファイル（4ファイル）**
- `db/migrations/000013_create_notifications_table.up/down.sql`
- `db/migrations/000014_create_reminders_table.up/down.sql`

**4.2 Domain層の拡張**
- ✅ Notificationエンティティの追加
- ✅ Reminderエンティティの追加
- ✅ CreateNotificationInput / CreateReminderInputの追加

**4.3 Repository層の実装**
- ✅ NotificationRepository の実装（CRUD操作）
  - `Create`: 通知作成
  - `FindAll`: ユーザーの全通知取得
  - `FindUnread`: 未読通知取得
  - `MarkAsRead`: 通知を既読にする
  - `MarkAllAsRead`: 全通知を既読にする
  - `Delete`: 通知削除
- ✅ ReminderRepository の実装
  - `Create`: リマインダー作成
  - `FindByTodoID`: TODOに紐づくリマインダー取得
  - `FindPending`: 送信待ちリマインダー取得
  - `MarkAsSent`: リマインダーを送信済みにする
  - `Delete`: リマインダー削除

**4.4 Service層の実装**
- ✅ NotificationService の実装
  - 通知のCRUD操作をラップ
- ✅ ReminderService の実装
  - `ProcessPendingReminders`: バックグラウンドワーカー用のリマインダー処理メソッド

**4.5 Handler層の実装**
- ✅ NotificationHandler の実装（6つのエンドポイント）
  - `GET /api/v1/notifications`: 通知一覧取得
  - `GET /api/v1/notifications/unread`: 未読通知取得
  - `GET /api/v1/notifications/stream`: SSEリアルタイム通知配信
  - `PUT /api/v1/notifications/:id/read`: 通知を既読にする
  - `PUT /api/v1/notifications/read-all`: 全通知を既読にする
  - `DELETE /api/v1/notifications/:id`: 通知削除
- ✅ ReminderHandler の実装（3つのエンドポイント）
  - `POST /api/v1/todos/:id/reminders`: リマインダー作成
  - `GET /api/v1/todos/:id/reminders`: リマインダー一覧取得
  - `DELETE /api/v1/reminders/:id`: リマインダー削除

**4.6 バックグラウンドワーカーの実装**
- ✅ `startReminderWorker`: 1分ごとにリマインダーをチェックして通知を作成
- ✅ Graceful Shutdown対応
- ✅ `ProcessPendingReminders`の定期実行

**4.7 SSE (Server-Sent Events) の実装**
- ✅ `StreamNotifications`: 5秒ごとに未読通知をプッシュ
- ✅ リアルタイム通知配信機能
- ✅ クライアント切断時の適切なクリーンアップ

**4.8 システム権限の実装**
- ✅ `FindByID`にシステム権限ロジック追加（userID=0で全TODOアクセス可能）
- ✅ バックグラウンドワーカーがユーザー所有権を超えてTODOにアクセス可能

**4.9 統合テストの追加**
- ✅ TestGetNotifications: 通知一覧取得のテスト
- ✅ TestGetUnreadNotifications: 未読通知取得のテスト
- ✅ TestMarkNotificationAsRead: 通知既読化のテスト
- ✅ TestMarkAllNotificationsAsRead: 全通知既読化のテスト
- ✅ TestDeleteNotification: 通知削除のテスト
- ✅ TestMarkNotificationAsReadUnauthorized: 他ユーザーの通知へのアクセス拒否テスト
- ✅ TestDeleteNotificationUnauthorized: 他ユーザーの通知削除拒否テスト
- ✅ TestCreateReminder: リマインダー作成のテスト
- ✅ TestGetRemindersByTodoID: リマインダー一覧取得のテスト
- ✅ TestDeleteReminder: リマインダー削除のテスト
- ✅ TestProcessPendingReminders: バックグラウンドワーカーの動作テスト

**合計: 35テスト、全てPASS**

### 追加されたファイル

#### マイグレーション（+4ファイル）
- `db/migrations/000013_create_notifications_table.up.sql`
- `db/migrations/000013_create_notifications_table.down.sql`
- `db/migrations/000014_create_reminders_table.up.sql`
- `db/migrations/000014_create_reminders_table.down.sql`

#### Domain層（+2ファイル）
- `internal/domain/notification.go`: Notificationエンティティ、CreateNotificationInput
- `internal/domain/reminder.go`: Reminderエンティティ、CreateReminderInput

#### Repository層（+2ファイル）
- `internal/repository/notification_repository.go`: 通知CRUD実装
- `internal/repository/reminder_repository.go`: リマインダー管理実装
- `internal/repository/repository.go`: NotificationRepository, ReminderRepositoryインターフェース追加

#### Service層（+2ファイル）
- `internal/service/notification_service.go`: 通知ビジネスロジック
- `internal/service/reminder_service.go`: リマインダービジネスロジック（ProcessPendingReminders含む）

#### Handler層（+2ファイル）
- `internal/handler/notification_handler.go`: 通知HTTPハンドラー（SSE含む）
- `internal/handler/reminder_handler.go`: リマインダーHTTPハンドラー

#### エントリーポイント
- `cmd/api/main.go`: 新しい依存性注入、ルーティング、バックグラウンドワーカー追加

#### Repository層の修正
- `internal/repository/todo_repository.go`: FindByIDにシステム権限ロジック追加

#### テスト
- `cmd/api/main_test.go`: 11個のテストを追加（通知7個、リマインダー3個、バックグラウンドワーカー1個）

**Phase 4で追加されたファイル数: 12ファイル**
**Phase 4で修正されたファイル数: 4ファイル**

### 主な改善点

1. **通知システムの実装**
   - 3種類の通知タイプ（deadline_reminder, todo_assigned, todo_completed）
   - ユーザーごとの通知管理
   - 既読/未読ステータスの管理
   - TODO紐付けによる関連情報の管理

2. **リマインダー機能の基盤**
   - TODOに対するリマインダー設定
   - 送信待ちリマインダーの管理
   - バックグラウンドワーカー用のProcessPendingRemindersメソッド
   - 自動通知作成機能

3. **セキュリティ**
   - ユーザー所有権の検証（他ユーザーの通知にアクセス不可）
   - JWTクレームを使用した認証
   - 404エラーによる情報漏洩防止

4. **データベース設計**
   - 通知とTODOの外部キー制約（CASCADE削除）
   - リマインダーの効率的な検索（部分インデックス）
   - GINインデックスによる高速な通知タイプ検索

### データベーススキーマ

**通知テーブル（notifications）**
```sql
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    todo_id INT REFERENCES todos(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,  -- 'deadline_reminder', 'todo_assigned', 'todo_completed'
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_todo ON notifications(todo_id) WHERE todo_id IS NOT NULL;
```

**リマインダーテーブル（reminders）**
```sql
CREATE TABLE IF NOT EXISTS reminders (
    id SERIAL PRIMARY KEY,
    todo_id INT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    remind_at TIMESTAMPTZ NOT NULL,
    is_sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_reminders_pending ON reminders(is_sent, remind_at) WHERE is_sent = FALSE;
CREATE INDEX idx_reminders_todo ON reminders(todo_id);
```

### 技術的なハイライト

1. **部分インデックスによる最適化**
   ```sql
   -- 送信待ちリマインダーのみをインデックス化
   CREATE INDEX idx_reminders_pending ON reminders(is_sent, remind_at) WHERE is_sent = FALSE;

   -- TODO紐付けのある通知のみをインデックス化
   CREATE INDEX idx_notifications_todo ON notifications(todo_id) WHERE todo_id IS NOT NULL;
   ```

2. **バックグラウンドワーカー用の設計**
   ```go
   // ProcessPendingReminders は送信待ちリマインダーを処理して通知を作成
   func (s *ReminderService) ProcessPendingReminders(ctx context.Context) error {
       reminders, err := s.reminderRepo.FindPending(ctx)
       // 各リマインダーに対して:
       // 1. TODOを取得
       // 2. 通知を作成
       // 3. リマインダーを送信済みにする
       // エラーが発生しても処理を継続
   }
   ```

3. **ユーザー所有権の検証**
   ```go
   // NotificationRepositoryでは全操作にuser_id条件を含める
   func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID, userID int) error {
       query := `UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`
       // rowsAffected == 0 の場合は ErrNotFound を返す
   }
   ```

4. **複合インデックスによる高速クエリ**
   ```sql
   -- 未読通知の取得を高速化
   CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);
   ```

5. **SSE (Server-Sent Events) によるリアルタイム通知**
   ```go
   func (h *NotificationHandler) StreamNotifications(c *gin.Context) {
       c.Header("Content-Type", "text/event-stream")
       c.Header("Cache-Control", "no-cache")
       c.Header("Connection", "keep-alive")

       ticker := time.NewTicker(5 * time.Second)
       for {
           select {
           case <-ticker.C:
               notifications, _ := h.service.GetUnreadNotifications(...)
               if len(notifications) > 0 {
                   c.SSEvent("notification", notifications)
                   c.Writer.Flush()
               }
           case <-c.Request.Context().Done():
               return
           }
       }
   }
   ```

6. **バックグラウンドワーカーの実装**
   ```go
   func startReminderWorker(ctx context.Context, reminderService *service.ReminderService) {
       ticker := time.NewTicker(1 * time.Minute)
       for {
           select {
           case <-ticker.C:
               reminderService.ProcessPendingReminders(ctx)
           case <-ctx.Done():
               return // Graceful Shutdown
           }
       }
   }
   ```

7. **システム権限の実装**
   ```go
   func (r *todoRepository) FindByID(ctx context.Context, todoID, userID int) (domain.Todo, error) {
       if userID == 0 {
           // システム権限: user_idチェックなし（バックグラウンドワーカー用）
           query = `SELECT ... FROM todos WHERE id = $1`
       } else {
           // 通常のユーザー権限: user_idチェックあり
           query = `SELECT ... FROM todos WHERE id = $1 AND user_id = $2`
       }
   }
   ```

### API使用例

**通知API**:
```bash
# 全通知を取得
GET /api/v1/notifications
Authorization: Bearer <token>

# 未読通知のみ取得
GET /api/v1/notifications/unread
Authorization: Bearer <token>

# 特定の通知を既読にする
PUT /api/v1/notifications/123/read
Authorization: Bearer <token>

# 全通知を既読にする
PUT /api/v1/notifications/read-all
Authorization: Bearer <token>

# 通知を削除
DELETE /api/v1/notifications/123
Authorization: Bearer <token>

# SSEでリアルタイム通知を受信
GET /api/v1/notifications/stream
Authorization: Bearer <token>
```

**リマインダーAPI**:
```bash
# リマインダーを作成
POST /api/v1/todos/123/reminders
Authorization: Bearer <token>
Content-Type: application/json

{
  "remind_at": "2026-12-31T10:00:00Z"
}

# TODOのリマインダー一覧を取得
GET /api/v1/todos/123/reminders
Authorization: Bearer <token>

# リマインダーを削除
DELETE /api/v1/reminders/456
Authorization: Bearer <token>
```

**SSE使用例（フロントエンド）**:
```javascript
const eventSource = new EventSource('/api/v1/notifications/stream', {
    headers: { 'Authorization': 'Bearer ' + token }
});

eventSource.addEventListener('notification', (event) => {
    const notifications = JSON.parse(event.data);
    console.log('新しい通知:', notifications);
    showNotificationBadge(notifications.length);
});

eventSource.onerror = (error) => {
    console.error('SSE接続エラー:', error);
    eventSource.close();
};
```

### テスト結果

```bash
=== RUN   TestGetNotifications
--- PASS: TestGetNotifications (0.08s)
=== RUN   TestGetUnreadNotifications
--- PASS: TestGetUnreadNotifications (0.09s)
=== RUN   TestMarkNotificationAsRead
--- PASS: TestMarkNotificationAsRead (0.09s)
=== RUN   TestMarkAllNotificationsAsRead
--- PASS: TestMarkAllNotificationsAsRead (0.09s)
=== RUN   TestDeleteNotification
--- PASS: TestDeleteNotification (0.09s)
=== RUN   TestMarkNotificationAsReadUnauthorized
--- PASS: TestMarkNotificationAsReadUnauthorized (0.08s)
=== RUN   TestDeleteNotificationUnauthorized
--- PASS: TestDeleteNotificationUnauthorized (0.08s)
=== RUN   TestCreateReminder
--- PASS: TestCreateReminder (0.15s)
=== RUN   TestGetRemindersByTodoID
--- PASS: TestGetRemindersByTodoID (0.14s)
=== RUN   TestDeleteReminder
--- PASS: TestDeleteReminder (0.16s)
=== RUN   TestProcessPendingReminders
--- PASS: TestProcessPendingReminders (0.12s)
PASS
ok  	gin-quickstart/examples/cmd/api	7.747s
```

**全35テスト、全てPASS**

### 次のステップ

Phase 4が完了しました。以下の機能が実装され、完全に動作しています：

1. ✅ **Phase 1**: レイヤードアーキテクチャへのリファクタリング
2. ✅ **Phase 2**: TODO機能の拡張（優先度、期限、ステータス、カテゴリー）
3. ✅ **Phase 3**: 検索・フィルタリング・統計情報
4. ✅ **Phase 4**: 通知・リマインダー機能

今後の拡張可能性：
- メール/Push通知の送信（Firebase Cloud Messaging等）
- タグ機能の完全実装（Phase 2で準備済み）
- サブタスク機能のUI対応
- ファイル添付機能

---

## Phase 5: 共有・コラボレーション機能 **（2026-01-16完了）**

### 実装内容

#### 1. プロジェクト機能
- プロジェクトのCRUD操作
- オーナー権限管理
- プロジェクトメンバー自動登録

**エンドポイント:**
```
POST   /api/v1/projects                      # プロジェクト作成
GET    /api/v1/projects                      # プロジェクト一覧
GET    /api/v1/projects/:id                  # プロジェクト取得
PUT    /api/v1/projects/:id                  # プロジェクト更新
DELETE /api/v1/projects/:id                  # プロジェクト削除
```

#### 2. メンバー管理
- メンバーの追加・削除
- 役割管理（owner/admin/member）
- メンバー一覧取得
- 役割更新機能

**エンドポイント:**
```
POST   /api/v1/projects/:id/members          # メンバー追加
GET    /api/v1/projects/:id/members          # メンバー一覧
DELETE /api/v1/projects/:id/members/:userId  # メンバー削除
PUT    /api/v1/projects/:id/members/:userId/role  # 役割更新
```

#### 3. TODO担当者機能
- TODOへの担当者割り当て
- プロジェクトメンバーのみ割り当て可能
- 担当者一覧取得
- 担当解除

**エンドポイント:**
```
POST   /api/v1/todos/:id/assignments         # 担当者割り当て
GET    /api/v1/todos/:id/assignments         # 担当者一覧
DELETE /api/v1/todos/:id/assignments/:userId # 担当者解除
```

#### 4. コメント機能
- TODOへのコメント追加
- コメントの編集・削除
- コメント一覧取得
- プロジェクトメンバーもアクセス可能

**エンドポイント:**
```
POST   /api/v1/todos/:id/comments            # コメント作成
GET    /api/v1/todos/:id/comments            # コメント一覧
GET    /api/v1/comments/:commentId           # コメント取得
PUT    /api/v1/comments/:commentId           # コメント更新
DELETE /api/v1/comments/:commentId           # コメント削除
```

### データベース設計

#### テーブル構成
```sql
-- projects: プロジェクト情報
-- project_members: プロジェクトメンバー（複合主キー）
-- todo_assignments: TODO担当者（複合主キー）
-- comments: TODOコメント
```

### アクセス制御

#### プロジェクト
- オーナーのみ: 更新、削除、メンバー追加、役割変更
- 管理者: メンバー追加
- メンバー: 閲覧のみ

#### TODO
- オーナー: 全権限
- プロジェクトメンバー: 閲覧、コメント追加、担当者割り当て

#### コメント
- 作成者のみ: 編集、削除
- プロジェクトメンバー: 閲覧

### テスト結果

```bash
=== RUN   TestCreateProject
--- PASS: TestCreateProject (0.08s)
=== RUN   TestGetProjects
--- PASS: TestGetProjects (0.09s)
=== RUN   TestGetProject
--- PASS: TestGetProject (0.08s)
=== RUN   TestUpdateProject
--- PASS: TestUpdateProject (0.08s)
=== RUN   TestDeleteProject
--- PASS: TestDeleteProject (0.08s)
=== RUN   TestAddMember
--- PASS: TestAddMember (0.09s)
=== RUN   TestGetMembers
--- PASS: TestGetMembers (0.08s)
=== RUN   TestRemoveMember
--- PASS: TestRemoveMember (0.09s)
=== RUN   TestUpdateMemberRole
--- PASS: TestUpdateMemberRole (0.09s)
=== RUN   TestAssignUser
--- PASS: TestAssignUser (0.08s)
=== RUN   TestGetAssignments
--- PASS: TestGetAssignments (0.08s)
=== RUN   TestUnassignUser
--- PASS: TestUnassignUser (0.08s)
=== RUN   TestCreateComment
--- PASS: TestCreateComment (0.08s)
=== RUN   TestGetCommentsByTodoID
--- PASS: TestGetCommentsByTodoID (0.09s)
=== RUN   TestGetComment
--- PASS: TestGetComment (0.09s)
=== RUN   TestUpdateComment
--- PASS: TestUpdateComment (0.10s)
=== RUN   TestDeleteComment
--- PASS: TestDeleteComment (0.10s)
PASS
```

**全51テスト、全てPASS**（Phase 1-5の累計）

### ファイル構成

**マイグレーション:**
- `000015_create_projects_table.up/down.sql`
- `000016_create_project_members_table.up/down.sql`
- `000017_create_todo_assignments_table.up/down.sql`
- `000018_create_comments_table.up/down.sql`

**Domain層:**
- `internal/domain/project.go`
- `internal/domain/project_member.go`
- `internal/domain/todo_assignment.go`
- `internal/domain/comment.go`

**Repository層:**
- `internal/repository/project_repository.go`
- `internal/repository/comment_repository.go`
- `internal/repository/todo_assignment_repository.go`

**Service層:**
- `internal/service/project_service.go`
- `internal/service/comment_service.go`
- `internal/service/todo_assignment_service.go`

**Handler層:**
- `internal/handler/project_handler.go`
- `internal/handler/comment_handler.go`
- `internal/handler/todo_assignment_handler.go`

### 次のステップ

Phase 5が完了しました。以下の機能が実装され、完全に動作しています：

1. ✅ **Phase 1**: レイヤードアーキテクチャへのリファクタリング
2. ✅ **Phase 2**: TODO機能の拡張（優先度、期限、ステータス、カテゴリー）
3. ✅ **Phase 3**: 検索・フィルタリング・統計情報
4. ✅ **Phase 4**: 通知・リマインダー機能
5. ✅ **Phase 5**: 共有・コラボレーション機能

今後の拡張可能性：
- セキュリティ・パフォーマンス強化（Phase 6）
- リアルタイム機能（WebSocket）
- ファイル添付機能
- タグ機能の完全実装
