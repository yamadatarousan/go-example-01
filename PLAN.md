# go-example-01 プロジェクト拡張プラン

## 前提
- AIはexamplesディレクトリ配下しか手を付けない
- AIはプロジェクトディレクトリの直下を原則として編集しない(README.mdやPLAN.mdなどのドキュメントを除く)
- 終わったタスクについてと進捗状況について切りの良いタイミングでDONE.mdに書き込んでいく
- コメントは日本語で書くこと
- **🐳 Docker・テスト関連ファイルの例外ルール**:
  - **Docker関連ファイルは常にプロジェクトルート直下のものを使用する**
    - `docker-compose.test.yml`
    - `testdata/seed.sql`
  - examples配下のテストコードも、プロジェクトルート直下のDocker設定・テストデータを参照する
  - これらのファイルを修正する際は、**例外的にプロジェクトルート直下を直接編集してよい**
  - 理由: 相対パスの不一致によるバグを防ぐため、Docker設定とテストデータは1箇所に集約する
  - **📝 main_test.goの実装方式**:
    - `getProjectRoot()`関数でgo.modを探してプロジェクトルートを自動検出
    - `filepath.Join(projectRoot, "docker-compose.test.yml")` でパス構築
    - これにより`examples/cmd/api/`でも`cmd/api/`でも**同じコード**で動作する
    - **そのまま写経しても正しく動作する**設計
- **🧪 テストの原則**:
  - **新しいエンドポイントを追加するたびに、そのエンドポイントの動作確認テストも必ず追加する**
  - 手動での動作確認に頼らず、自動テストで品質を保証する
  - 各フェーズの完了時には、そのフェーズで追加された全エンドポイントのテストが存在することを確認する
  - テストは統合テスト形式で実装し、正常系・異常系の両方をカバーする

## 📌 現状分析

### 現在の機能
- ✅ ユーザー認証（JWT）
- ✅ ユーザー登録・ログイン
- ✅ ロールベースアクセス制御（user/admin）
- ✅ TODO CRUD操作
- ✅ 監査ログ（TODO作成のみ）
- ✅ PostgreSQL + マイグレーション管理
- ✅ 統合テスト（カバレッジ向上）
- ✅ Docker対応

### 現在の技術スタック
- **言語**: Go 1.24.3
- **フレームワーク**: Gin
- **DB**: PostgreSQL (Docker)
- **認証**: JWT (golang-jwt/jwt)
- **テスト**: testify
- **マイグレーション**: golang-migrate

### プロジェクト構造
```
go-example-01/
├── main.go                   # 単一ファイル実装（426行）
├── repository.go             # データアクセス層（139行）
├── main_test.go              # JWTテスト
├── integration_test.go       # 統合テスト
├── db/migrations/            # マイグレーションファイル（7個）
├── testdata/                 # テスト用シードデータ
├── examples/                 # 実装例（PUT/DELETE追加済み）
└── docker-compose.yml        # PostgreSQL設定
```

---

## 🎯 拡張の目的

現在は**学習用のシンプルな実装**ですが、以下の目標に向けて段階的に拡張します：

1. **保守性の向上** - モジュール化とレイヤー分離
2. **スケーラビリティ** - 複数機能の追加に対応
3. **本番環境対応** - セキュリティ、パフォーマンス、監視
4. **チーム開発対応** - コード規約、ドキュメント、CI/CD

---

## 🏗️ アーキテクチャ設計

### レイヤードアーキテクチャとは

レイヤードアーキテクチャ（Layered Architecture）は、アプリケーションを複数の層（レイヤー）に分割し、各層が明確な責務を持つ設計パターンです。各レイヤーは下位レイヤーにのみ依存し、上位レイヤーには依存しないという**単方向の依存関係**を持ちます。

#### 基本原則

1. **関心の分離（Separation of Concerns）**: 各レイヤーは特定の責務のみを担当
2. **依存関係の方向**: 上位レイヤー → 下位レイヤーの単方向依存
3. **抽象化**: インターフェースを通じて層間を疎結合に保つ
4. **置き換え可能性**: 各レイヤーの実装を独立して変更可能

### 本アプリケーションのレイヤー構成

このプロジェクトでは、以下の4層構造を採用しています：

```
┌─────────────────────────────────────┐
│   Handler層（プレゼンテーション層）  │  ← HTTPリクエスト/レスポンス処理
├─────────────────────────────────────┤
│   Service層（ビジネスロジック層）    │  ← ビジネスルール、トランザクション制御
├─────────────────────────────────────┤
│   Repository層（データアクセス層）   │  ← データベース操作の抽象化
├─────────────────────────────────────┤
│   Domain層（ドメインモデル層）       │  ← エンティティ、ドメインルール
└─────────────────────────────────────┘
```

#### 各レイヤーの責務

**1. Domain層（最下層）**
- **役割**: ビジネスの核となるエンティティとルールを定義
- **内容**:
  - エンティティ（Todo, User）
  - ドメイン固有のエラー定義（ErrNotFound, ErrUnauthorizedなど）
  - 入力検証用の構造体（SignupInput, LoginInput）
- **依存**: なし（他のレイヤーに依存しない）
- **例**: `domain/todo.go`, `domain/user.go`, `domain/errors.go`

**2. Repository層**
- **役割**: データの永続化と取得を抽象化
- **内容**:
  - データベース操作のインターフェース定義
  - SQLクエリの実行
  - トランザクション管理
- **依存**: Domain層のみ
- **例**: `repository/todo_repository.go`, `repository/user_repository.go`

**3. Service層**
- **役割**: ビジネスロジックの実装
- **内容**:
  - 複数のRepositoryを組み合わせた処理
  - 認証・認可のロジック（JWT生成、パスワード検証）
  - トランザクションの制御
- **依存**: Domain層、Repository層のインターフェース
- **例**: `service/auth_service.go`, `service/todo_service.go`

**4. Handler層（最上層）**
- **役割**: HTTPリクエストとレスポンスの処理
- **内容**:
  - リクエストのバインディング（JSON → 構造体）
  - Serviceの呼び出し
  - レスポンスの生成（構造体 → JSON）
  - HTTPステータスコードの設定
- **依存**: Service層、Domain層
- **例**: `handler/todo_handler.go`, `handler/user_handler.go`

#### 依存関係の流れ（例: TODO作成）

```
HTTPリクエスト
    ↓
Handler層: todoHandler.CreateTodo()
    ↓ （Serviceを呼び出し）
Service層: todoService.CreateTodo()
    ↓ （Repositoryを呼び出し）
Repository層: todoRepo.CreateTodoWithAudit()
    ↓ （データベース操作）
PostgreSQL
    ↓
レスポンスを逆順に返却
```

### レイヤードアーキテクチャの採用理由

#### 1. **保守性の向上**
- **問題**: 単一ファイル（main.go 516行）では、コードの場所を見つけにくい
- **解決**: 各レイヤーごとにファイルを分割し、責務が明確化
- **効果**: バグ修正や機能追加時に、影響範囲を特定しやすい

#### 2. **テスタビリティの向上**
- **問題**: データベースに依存したテストは遅く、セットアップが複雑
- **解決**: インターフェースを使うことで、モック実装に差し替え可能
- **効果**:
  - Serviceのユニットテストでは、RepositoryをモックDB で代替
  - Handlerのユニットテストでは、Serviceをモックで代替

```go
// テスト例: Serviceのユニットテスト
mockRepo := &MockTodoRepository{
    FindAllFunc: func(userID int) ([]domain.Todo, error) {
        return []domain.Todo{{ID: 1, Name: "Test"}}, nil
    },
}
service := service.NewTodoService(mockRepo)
// serviceのテストを実行
```

#### 3. **拡張性の確保**
- **問題**: 新機能追加時に既存コードへの影響が大きい
- **解決**: 各レイヤーが疎結合なため、影響を局所化
- **効果**:
  - 新しいエンドポイント追加 → Handlerのみ追加
  - ビジネスロジック変更 → Serviceのみ修正
  - データベース変更 → Repositoryのみ修正

#### 4. **チーム開発の効率化**
- **問題**: 複数人が同じファイルを編集すると競合が発生
- **解決**: レイヤーごと、機能ごとにファイルが分かれる
- **効果**:
  - フロントエンド担当 → Handler層を編集
  - ビジネスロジック担当 → Service層を編集
  - DB担当 → Repository層を編集
  - マージ競合が減少

#### 5. **技術的な置き換えが容易**
- **問題**: 将来的にデータベースやフレームワークを変更したい
- **解決**: インターフェースで抽象化されているため、実装の差し替えが可能
- **効果**:
  - PostgreSQL → MySQL に変更 → Repository層のみ修正
  - Gin → Echo に変更 → Handler層のみ修正
  - Service層、Domain層は影響を受けない

#### 6. **ビジネスロジックの再利用**
- **問題**: HTTPエンドポイント以外（CLI、gRPC、バッチ処理）でも同じロジックを使いたい
- **解決**: Service層がHTTPに依存しないため、どこからでも呼び出し可能
- **効果**:
  - REST API、gRPC、CLI ツールで同じServiceを共有
  - ビジネスロジックの重複を防ぐ

### 実装上の工夫

#### 依存性注入（DI: Dependency Injection）

各レイヤーは、必要な依存関係をコンストラクタで受け取ります。

```go
// Repository層の作成
todoRepo := repository.NewTodoRepository(db)
userRepo := repository.NewUserRepository(db)

// Service層の作成（Repositoryを注入）
authService := service.NewAuthService(userRepo, jwtSecret)
todoService := service.NewTodoService(todoRepo)

// Handler層の作成（Serviceを注入）
userHandler := handler.NewUserHandler(authService)
todoHandler := handler.NewTodoHandler(todoService)
```

この設計により、各レイヤーは具体的な実装ではなく、インターフェースに依存します（依存性逆転の原則：DIP）。

#### インターフェースの活用

Repository層はインターフェースとして定義されています。

```go
// インターフェース定義
type TodoRepository interface {
    FindAll(userID int) ([]domain.Todo, error)
    CreateTodoWithAudit(ctx context.Context, todo domain.Todo) (domain.Todo, error)
    // ...
}

// 実装（PostgreSQL版）
type todoRepository struct {
    db *sql.DB
}

// 将来的にはMySQL版、MongoDB版なども作成可能
```

これにより、Service層は「どのデータベースを使うか」を知る必要がありません。

---

## 📈 拡張ロードマップ

### Phase 1: コード構造のリファクタリング（Week 1-2）✅ **完了: 2026-01-12**

#### 目標
単一ファイルからレイヤードアーキテクチャへ移行

#### タスク

**1.1 ディレクトリ構造の再編成**
```
go-example-01/
├── cmd/
│   └── api/
│       └── main.go           # エントリーポイント
├── internal/
│   ├── domain/               # ドメインモデル
│   │   ├── user.go
│   │   ├── todo.go
│   │   └── errors.go
│   ├── handler/              # HTTPハンドラー
│   │   ├── user_handler.go
│   │   ├── todo_handler.go
│   │   └── admin_handler.go
│   ├── repository/           # データアクセス
│   │   ├── user_repository.go
│   │   └── todo_repository.go
│   ├── service/              # ビジネスロジック
│   │   ├── auth_service.go
│   │   └── todo_service.go
│   ├── middleware/           # ミドルウェア
│   │   ├── auth.go
│   │   ├── admin.go
│   │   └── request_id.go
│   └── config/               # 設定管理
│       └── config.go
├── pkg/                      # 外部公開可能なパッケージ
│   ├── jwt/
│   │   └── jwt.go
│   └── validator/
│       └── validator.go
├── db/
│   └── migrations/
├── tests/
│   ├── integration/
│   └── fixtures/
└── docs/
    ├── api/                  # API仕様書
    └── architecture/         # アーキテクチャドキュメント
```

**1.2 レイヤー分離の実装** ✅
- ✅ Domain層: エンティティとビジネスルール
- ✅ Repository層: データベース操作の抽象化
- ✅ Service層: ビジネスロジックの集約
- ✅ Handler層: HTTPリクエスト/レスポンス処理
- ✅ インターフェースによる依存性注入

**1.3 設定管理の改善** ✅
- ✅ 環境変数の一元管理
- ✅ config.Loadによる設定読み込み

#### 成果物
- ✅ リファクタリング後のコード（18ファイル）
- ✅ レイヤードアーキテクチャへの移行完了

---

### Phase 2: TODO機能の拡張（Week 3-4）✅ **完了: 2026-01-13**

#### 目標
TODOアプリとしての実用性を高める

#### タスク

**2.1 TODO機能の充実**
- [ ] **優先度**: `priority` (high/medium/low)
- [ ] **期限**: `due_date` (TIMESTAMPTZ)
- [ ] **ステータス**: `status` (todo/in_progress/done)
- [ ] **カテゴリー**: `category_id` (FK to categories)
- [ ] **タグ機能**: `tags` (many-to-many)
- [ ] **サブタスク**: `parent_todo_id` (自己参照)

**マイグレーション追加**
```sql
-- 000008_add_todo_extensions.up.sql
ALTER TABLE todos ADD COLUMN priority VARCHAR(20) DEFAULT 'medium';
ALTER TABLE todos ADD COLUMN due_date TIMESTAMPTZ;
ALTER TABLE todos ADD COLUMN status VARCHAR(20) DEFAULT 'todo';
ALTER TABLE todos ADD COLUMN description TEXT;
ALTER TABLE todos ADD COLUMN parent_todo_id INT REFERENCES todos(id);
```

**2.2 カテゴリー機能**
```sql
-- 000009_create_categories_table.up.sql
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7),
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE todos ADD COLUMN category_id INT REFERENCES categories(id);
```

**2.3 タグ機能**
```sql
-- 000010_create_tags_table.up.sql
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS todo_tags (
    todo_id INT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    tag_id INT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (todo_id, tag_id)
);
```

**2.4 新しいエンドポイント**
```
POST   /api/v1/todos/:id/complete     # TODO完了
POST   /api/v1/todos/:id/reopen       # TODO再開
GET    /api/v1/todos/overdue          # 期限切れTODO一覧
GET    /api/v1/todos/today            # 今日のTODO
GET    /api/v1/todos/week             # 今週のTODO
POST   /api/v1/categories             # カテゴリー作成
GET    /api/v1/categories             # カテゴリー一覧
PUT    /api/v1/categories/:id         # カテゴリー更新
DELETE /api/v1/categories/:id         # カテゴリー削除
```

**2.5 全エンドポイントの統合テスト**
Phase 3に進む前に、Phase 1およびPhase 2で実装された全エンドポイントが正しく動作することを確認するテストを作成します。これまでに実装された全てのエンドポイントを網羅的にテストします。

**テスト対象エンドポイント**:
- [ ] **ユーザー認証**
  - POST /signup - ユーザー登録
  - POST /login - ログイン
- [ ] **TODO基本操作**
  - GET /api/v1/todos - TODO一覧取得
  - GET /api/v1/todos/:id - TODO詳細取得
  - POST /api/v1/todos - TODO作成（優先度、期限、ステータス、説明等を含む）
  - PUT /api/v1/todos/:id - TODO更新
  - DELETE /api/v1/todos/:id - TODO削除
- [ ] **TODO拡張機能**（Phase 2で追加）
  - POST /api/v1/todos/:id/complete - TODO完了
  - POST /api/v1/todos/:id/reopen - TODO再開
  - GET /api/v1/todos/overdue - 期限切れTODO一覧
  - GET /api/v1/todos/today - 今日のTODO一覧
  - GET /api/v1/todos/week - 今週のTODO一覧
- [ ] **カテゴリー機能**（Phase 2で追加）
  - POST /api/v1/categories - カテゴリー作成
  - GET /api/v1/categories - カテゴリー一覧取得
  - GET /api/v1/categories/:id - カテゴリー詳細取得
  - PUT /api/v1/categories/:id - カテゴリー更新
  - DELETE /api/v1/categories/:id - カテゴリー削除
- [ ] **管理者機能**
  - GET /api/v1/admin/users - 全ユーザー取得（管理者のみ）

**テスト方針**:
- 統合テストとして実装（実際のDBを使用）
- 各エンドポイントの正常系と異常系をカバー
- 認証が必要なエンドポイントでは、JWTトークンの検証もテスト
- ロールベースアクセス制御（管理者権限）のテスト

**テストファイル**:
- `examples/tests/integration/endpoint_test.go` - 全エンドポイントのテスト
- 既存の `integration_test.go` を拡張する形でも可

#### 成果物
- ✅ 拡張されたTODOモデル（優先度、期限、ステータス、説明等）
- ✅ カテゴリー機能の実装
- ✅ タグ機能の基盤実装（テーブル作成済み）
- ✅ 新しいエンドポイントのテスト（21テスト）
- ✅ **全エンドポイントの統合テスト**（Phase 2.5で追加）
- ✅ context.Contextの全DB操作への適用

---

### Phase 3: 検索・フィルタリング機能（Week 5）✅ **完了: 2026-01-15**

#### 目標
大量のTODOを効率的に管理できるようにする

#### タスク

**3.1 高度な検索機能**
```
GET /api/v1/todos?status=done&priority=high&category=work&tag=urgent&sort=due_date&order=asc
```

**クエリパラメータ**
- `status`: ステータスフィルター
- `priority`: 優先度フィルター
- `category`: カテゴリーフィルター
- `tag`: タグフィルター
- `search`: 名前・説明での全文検索
- `due_from`, `due_to`: 期限範囲
- `sort`: ソート項目（created_at, due_date, priority）
- `order`: ソート順序（asc, desc)
- `page`, `limit`: ページネーション

**3.2 全文検索の実装**
```sql
-- PostgreSQLの全文検索インデックス
CREATE INDEX idx_todos_fulltext ON todos
USING gin(to_tsvector('english', name || ' ' || COALESCE(description, '')));
```

**3.3 統計情報エンドポイント**
```
GET /api/v1/todos/stats

Response:
{
  "total": 150,
  "by_status": {
    "todo": 50,
    "in_progress": 30,
    "done": 70
  },
  "by_priority": {
    "high": 20,
    "medium": 80,
    "low": 50
  },
  "overdue": 15,
  "due_today": 5,
  "due_this_week": 12
}
```

#### 成果物
- ✅ 検索・フィルタリング実装（動的SQLクエリビルディング）
- ✅ 全文検索機能（PostgreSQL GINインデックス）
- ✅ 統計情報API（ステータス別・優先度別カウント）
- ✅ ページネーション機能
- ✅ ソート機能（複数項目対応）
- ✅ **全エンドポイントの統合テスト**（23テスト、全てPASS）

---

### Phase 4: 通知・リマインダー機能（Week 6）✅ **完了: 2026-01-15**

#### 目標
期限管理を支援する通知機能

#### タスク

**4.1 通知モデル** ✅
```sql
-- 000013_create_notifications_table.up.sql (実装済み)
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

**4.2 リマインダー設定** ✅
```sql
-- 000014_create_reminders_table.up.sql (実装済み)
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

**4.3 通知エンドポイント** ✅
```
GET    /api/v1/notifications              # 通知一覧
GET    /api/v1/notifications/unread       # 未読通知
GET    /api/v1/notifications/stream       # SSEリアルタイム通知配信
PUT    /api/v1/notifications/:id/read     # 既読にする
PUT    /api/v1/notifications/read-all     # 全て既読
DELETE /api/v1/notifications/:id          # 通知削除
```

**4.4 リマインダーエンドポイント** ✅
```
POST   /api/v1/todos/:id/reminders        # リマインダー作成
GET    /api/v1/todos/:id/reminders        # リマインダー一覧
DELETE /api/v1/reminders/:id              # リマインダー削除
```

**4.5 バックグラウンドワーカー** ✅
- ✅ 定期的に期限をチェックするジョブ（1分ごと）
- ✅ リマインダー送信処理（ProcessPendingReminders）
- ✅ 通知の自動生成
- ✅ Graceful Shutdown対応

**4.6 SSE (Server-Sent Events)** ✅
- ✅ リアルタイム通知配信（5秒ごとにプッシュ）
- ✅ クライアント切断時の適切なクリーンアップ

#### 成果物
- ✅ 通知システム
- ✅ リマインダー機能
- ✅ バックグラウンドジョブ
- ✅ SSEリアルタイム通知
- ✅ システム権限ロジック（userID=0でバックグラウンドワーカー用）
- ✅ **全エンドポイントの統合テスト**（通知・リマインダーAPI）
- ✅ **35テスト、全てPASS**

---

### Phase 5: 共有・コラボレーション機能（Week 7-8）

#### 目標
複数ユーザーでのTODO管理

#### タスク

**5.1 プロジェクト機能**
```sql
-- 000013_create_projects_table.up.sql
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    owner_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE todos ADD COLUMN project_id INT REFERENCES projects(id);
```

**5.2 プロジェクトメンバー管理**
```sql
-- 000014_create_project_members_table.up.sql
CREATE TABLE IF NOT EXISTS project_members (
    project_id INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member',  -- 'owner', 'admin', 'member'
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);
```

**5.3 TODO担当者**
```sql
-- 000015_create_todo_assignments_table.up.sql
CREATE TABLE IF NOT EXISTS todo_assignments (
    todo_id INT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (todo_id, user_id)
);
```

**5.4 コメント機能**
```sql
-- 000016_create_comments_table.up.sql
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    todo_id INT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**5.5 新しいエンドポイント**
```
# プロジェクト管理
POST   /api/v1/projects
GET    /api/v1/projects
GET    /api/v1/projects/:id
PUT    /api/v1/projects/:id
DELETE /api/v1/projects/:id

# メンバー管理
POST   /api/v1/projects/:id/members       # メンバー追加
GET    /api/v1/projects/:id/members       # メンバー一覧
DELETE /api/v1/projects/:id/members/:uid  # メンバー削除

# TODO担当
POST   /api/v1/todos/:id/assign           # 担当者割り当て
DELETE /api/v1/todos/:id/assign/:uid      # 担当解除

# コメント
POST   /api/v1/todos/:id/comments         # コメント追加
GET    /api/v1/todos/:id/comments         # コメント一覧
PUT    /api/v1/comments/:id               # コメント編集
DELETE /api/v1/comments/:id               # コメント削除
```

#### 成果物
- ✅ プロジェクト機能 **（2026-01-16完了）**
- ✅ メンバー管理
- ✅ TODO担当機能
- ✅ コメント機能
- ✅ **全エンドポイントの統合テスト**（プロジェクト・メンバー・担当・コメントAPI）
- ✅ **51テスト、全てPASS**

---

### Phase 6: セキュリティ・パフォーマンス強化（Week 9-10）

#### 目標
本番環境に耐える品質

#### タスク

**6.1 セキュリティ強化**
- [ ] リフレッシュトークンの実装
- [ ] レート制限（100req/min）
- [ ] セキュリティヘッダーの追加
- [ ] 監査ログの拡充（全CRUD操作）
- [ ] HTTPS強制

**6.2 パフォーマンス最適化**
- [ ] データベースインデックスの追加
```sql
CREATE INDEX idx_todos_user_status ON todos(user_id, status);
CREATE INDEX idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL;
```
- [ ] N+1クエリの排除
- [ ] Redisキャッシュの導入
- [ ] ページネーション改善

**6.3 監視・ロギング**
- [ ] 構造化ロギング（zap）
- [ ] Prometheusメトリクス
- [ ] ヘルスチェックエンドポイント

#### 成果物
- [ ] セキュリティ強化実装
- [ ] パフォーマンス最適化
- [ ] 監視基盤

---

### Phase 7: CI/CD自動化（Week 11-12）

#### 目標
自動化されたテスト・デプロイパイプラインの構築

#### タスク

**7.1 GitHub Actions**
```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go test -v -coverprofile=coverage.out ./...
```

**7.2 Docker最適化**
- [ ] マルチステージビルド
- [ ] docker-compose.yml更新
- [ ] 本番用Dockerfile

**7.3 自動デプロイ**
- [ ] ステージング環境へのデプロイ設定
- [ ] 本番環境へのデプロイ設定（手動承認）
- [ ] ロールバック機能

#### 成果物
- [ ] CI/CDパイプライン
- [ ] 自動デプロイ環境

---

### Phase 8: フロントエンド開発（Week 13-16）

#### 目標
Webフロントエンドの実装

#### タスク

**8.1 技術選定**
- **言語**: TypeScript
  - 型安全性によるバグの早期発見
  - IDEの強力な補完・リファクタリング支援
  - APIレスポンスの型定義
- **フレームワーク**: Next.js 14 (App Router)
  - React 18+ Server Components
  - Server Actions for mutations
  - App Router (Pages Routerは使用しない)
- **データ取得戦略**
  - **Server Components**: GETリクエスト（参照系）はRSCで直接fetch
  - **Server Actions**: POST/PUT/DELETE（更新系）はServer Actionsで実装
  - **TanStack Query**: クライアント側の楽観的更新・リアルタイム更新のみ（補助的）
  - **HTTPクライアント**: fetch API（Axiosは使用しない）
- **状態管理**: Zustand
  - 軽量でシンプル、RSCと相性が良い
  - グローバルなUI状態（モーダル、トースト等）に使用
- **UIライブラリ**: Tailwind CSS + shadcn/ui
  - Tailwind CSS: ユーティリティファーストCSS
  - shadcn/ui: 再利用可能なコンポーネント
- **フォーム**: React Hook Form + Zod
  - React Hook Form: パフォーマンス重視
  - Zod: TypeScript型安全なバリデーション
- **認証戦略**
  - **JWT保存**: httpOnly cookie（XSS攻撃対策）
  - **CSRF対策**: SameSite=Strict + CSRFトークン
  - **認証チェック**: Next.js middleware
  - **トークン更新**: リフレッシュトークン方式
- **テスト戦略**
  - **Server Components**: 統合テスト・E2E中心（Playwright）
  - **Client Components**: Vitest + React Testing Library
  - **API Routes/Server Actions**: Vitest統合テスト
  - **E2E**: Playwright（クロスブラウザ対応）

**8.2 プロジェクト構造**
```
frontend/
├── src/
│   ├── app/                  # Next.js App Router
│   │   ├── (auth)/          # 認証関連ページ（Route Group）
│   │   │   ├── login/
│   │   │   └── signup/
│   │   ├── dashboard/       # ダッシュボード（Server Component）
│   │   ├── todos/           # TODO関連ページ（Server Component）
│   │   ├── middleware.ts    # 認証チェック・JWT検証
│   │   └── layout.tsx       # ルートレイアウト
│   ├── actions/             # Server Actions（mutations）
│   │   ├── auth.ts         # 認証関連アクション
│   │   ├── todos.ts        # TODO CRUD
│   │   └── projects.ts     # プロジェクト管理
│   ├── components/          # 再利用可能なコンポーネント
│   │   ├── ui/             # shadcn/uiコンポーネント
│   │   ├── layouts/        # レイアウトコンポーネント
│   │   └── features/       # 機能別コンポーネント（Client Components）
│   ├── lib/                 # ユーティリティ・設定
│   │   ├── api/            # fetch関数（Server Component用）
│   │   ├── hooks/          # カスタムフック（Client Component用）
│   │   ├── auth/           # JWT検証・CSRF対策
│   │   └── utils/          # ヘルパー関数
│   ├── types/               # TypeScript型定義
│   │   ├── api.ts          # APIレスポンス型
│   │   ├── models.ts       # ドメインモデル型
│   │   └── index.ts
│   └── store/               # 状態管理（Zustand、UI状態のみ）
└── public/                  # 静的ファイル
```

**8.3 主要画面**
- [ ] ログイン画面（Client Component: React Hook Form + Zod + Server Action）
- [ ] サインアップ画面（Client Component: React Hook Form + Zod + Server Action）
- [ ] ダッシュボード（Server Component: SSR統計情報表示）
- [ ] TODO一覧（Server Component + Client Component: フィルタリング・検索）
  - リスト表示: Server Component
  - フィルタUI: Client Component
  - リアルタイム更新: TanStack Query（楽観的更新）
- [ ] TODO詳細・編集（Client Component: ドラッグ&ドロップ + Server Action）
- [ ] プロジェクト管理（Server Component + Client Component）
- [ ] 設定画面（Client Component: プロフィール編集 + Server Action）

**8.4 型定義の実装**
```typescript
// types/api.ts
export interface Todo {
  id: number;
  name: string;
  description?: string;
  status: 'todo' | 'in_progress' | 'done';
  priority: 'low' | 'medium' | 'high';
  due_date?: string;
  user_id: number;
  category_id?: number;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: number;
  username: string;
  email: string;
  bio?: string;
  image?: string;
  role: 'user' | 'admin';
  created_at: string;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface ApiError {
  error: string;
  details?: string;
}
```

**8.5 データ取得・API通信の実装**
- [ ] Server Components用fetch関数（JWT自動付与、Next.js middleware経由）
- [ ] Server Actions実装（CSRF保護付き）
- [ ] API型定義の自動生成（openapi-typescript）
- [ ] エラーハンドリング（統一エラー処理）
- [ ] TanStack Query設定（クライアント側の楽観的更新用、補助的）
- [ ] httpOnly cookie認証フロー実装
  - [ ] ログイン時のcookie設定
  - [ ] middlewareでのJWT検証
  - [ ] CSRF対策実装

**実装パターン例:**

```typescript
// app/todos/page.tsx (Server Component)
async function TodosPage() {
  // Server Componentで直接fetch（JWT自動付与）
  const todos = await fetch('http://backend/api/v1/todos', {
    cache: 'no-store', // リアルタイムデータの場合
  }).then(r => r.json())

  return <TodoList todos={todos} />
}

// actions/todos.ts (Server Action)
'use server'
import { revalidatePath } from 'next/cache'

export async function createTodo(formData: FormData) {
  const csrfToken = formData.get('csrfToken')
  // CSRF検証

  const response = await fetch('http://backend/api/v1/todos', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: formData.get('name'),
      description: formData.get('description'),
    }),
  })

  if (!response.ok) throw new Error('Failed to create todo')

  revalidatePath('/todos') // キャッシュ再検証
  return response.json()
}

// components/TodoForm.tsx (Client Component)
'use client'
import { createTodo } from '@/actions/todos'
import { useTransition } from 'react'

export function TodoForm() {
  const [isPending, startTransition] = useTransition()

  async function handleSubmit(formData: FormData) {
    startTransition(async () => {
      await createTodo(formData)
    })
  }

  return (
    <form action={handleSubmit}>
      {/* フォームフィールド */}
      <button disabled={isPending}>Create</button>
    </form>
  )
}

// middleware.ts (JWT検証)
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const token = request.cookies.get('token')

  if (!token && !request.nextUrl.pathname.startsWith('/login')) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  // JWT検証ロジック
  return NextResponse.next()
}
```

**8.6 レスポンシブデザイン**
- [ ] PC対応（1920px以上）
- [ ] タブレット対応（768px-1919px）
- [ ] スマートフォン対応（〜767px）
- [ ] ダークモード対応

**8.7 パフォーマンス最適化**
- [ ] コード分割（React.lazy, dynamic import）
- [ ] 画像最適化（Next.js Image）
- [ ] SSR/SSG活用
- [ ] バンドルサイズ最適化

#### 成果物
- [ ] TypeScriptベースのフロントエンドアプリケーション
- [ ] 型安全なAPI通信層
- [ ] 再利用可能なコンポーネントライブラリ
- [ ] ユニットテスト
- [ ] E2Eテスト（Playwright）
- [ ] Storybookコンポーネントカタログ（オプション）

---

### Phase 9: ドキュメント作成（Week 17-18）

> **⚠️ 注意**: このフェーズの実施前に、各ドキュメントの必要性を検討します。
> プロジェクトの規模、用途、チーム構成に応じて、一部またはすべてを省略する可能性があります。
> 必要性を判断する際の基準：
> - プロジェクトを他者に公開するか
> - チーム開発を行うか
> - 長期的なメンテナンスを想定しているか
> - コントリビューターを募集するか

#### 目標
プロジェクトの理解とメンテナンスを容易にするドキュメントの整備

#### タスク

**9.1 アーキテクチャドキュメント**
- [ ] **ARCHITECTURE.md**: システム設計の説明
  - レイヤードアーキテクチャの構造
  - 各レイヤーの責務と依存関係
  - モジュール間の連携フロー
  - 主要な設計判断とその理由

**9.2 API仕様書**
- [ ] **OpenAPI/Swagger仕様**: APIドキュメントの自動生成
  - Swaggerアノテーションの追加
  - Swagger UI (`/api/docs`) の実装
  - 全エンドポイントの仕様記述
- [ ] **API.md**: API利用ガイド
  - 認証フロー
  - エンドポイント一覧
  - リクエスト/レスポンス例
  - エラーコード一覧
- [ ] **Postmanコレクション**: テスト用コレクション
  - 環境変数設定
  - 全エンドポイントのテストケース

**9.3 運用ドキュメント**
- [ ] **DEPLOYMENT.md**: デプロイ手順
  - 環境構築手順
  - データベースマイグレーション
  - 環境変数の設定
  - デプロイ先の選択肢（Heroku, AWS, GCP等）
  - トラブルシューティング

**9.4 開発者向けドキュメント**
- [ ] **CONTRIBUTING.md**: コントリビューションガイド
  - 開発環境のセットアップ
  - コーディング規約
  - プルリクエストの作成方法
  - コミットメッセージの規約
  - テストの実行方法
- [ ] **README.md拡充**: プロジェクト概要の充実
  - 機能一覧
  - デモ/スクリーンショット
  - クイックスタート
  - 技術スタック
  - ライセンス情報

**9.5 アーキテクチャ図**
- [ ] システム全体図（レイヤー構造）
- [ ] データベースER図
- [ ] APIエンドポイント図
- [ ] デプロイ構成図

#### 成果物（このフェーズ実施が決定した場合）
- [ ] ARCHITECTURE.md
- [ ] API.md
- [ ] DEPLOYMENT.md
- [ ] CONTRIBUTING.md
- [ ] OpenAPI仕様書
- [ ] Postmanコレクション
- [ ] アーキテクチャ図（複数）
- [ ] 充実したREADME.md

#### このフェーズをスキップする場合
最低限の以下のドキュメントのみを保持：
- [ ] README.md（基本的な使い方のみ）
- [ ] コード内のコメント（重要な処理の説明）

---

## 🛠 技術スタックの拡張

### バックエンド追加予定の技術

| カテゴリ | 技術 | 用途 |
|---------|------|------|
| **ORM** | GORM | データベース抽象化 |
| **キャッシュ** | Redis | セッション、キャッシュ |
| **ログ** | zap | 構造化ログ |
| **メトリクス** | Prometheus | 監視 |
| **ドキュメント** | swag | OpenAPI/Swagger |
| **設定管理** | viper | 環境変数 |
| **タスクキュー** | asynq | バックグラウンドジョブ |

### フロントエンド技術スタック

| カテゴリ | 技術 | 用途 |
|---------|------|------|
| **言語** | TypeScript 5+ | 型安全なフロントエンド開発 |
| **フレームワーク** | Next.js 14+ | React フルスタックフレームワーク |
| **ランタイム** | React 18+ | UIライブラリ |
| **状態管理** | Zustand | 軽量な状態管理 |
| **サーバーステート** | TanStack Query | API通信・キャッシング |
| **HTTP Client** | Axios | HTTP通信 |
| **スタイリング** | Tailwind CSS | ユーティリティCSS |
| **コンポーネント** | shadcn/ui | 再利用可能UIコンポーネント |
| **フォーム** | React Hook Form | フォーム管理 |
| **バリデーション** | Zod | TypeScript型安全バリデーション |
| **テスト** | Vitest | 高速テストランナー |
| **コンポーネントテスト** | React Testing Library | コンポーネントテスト |
| **E2Eテスト** | Playwright | エンドツーエンドテスト |
| **型生成** | openapi-typescript | OpenAPIからTypeScript型生成 |
| **リンター** | ESLint | コード品質管理 |
| **フォーマッター** | Prettier | コードフォーマット |

---

## 📊 進捗管理

### マイルストーン

- [ ] **Milestone 1** (Week 1-4): コア機能拡張
  - リファクタリング完了
  - TODO機能拡張完了

- [ ] **Milestone 2** (Week 5-8): 協働機能
  - 検索機能完了
  - 通知機能完了
  - 共有機能完了

- [ ] **Milestone 3** (Week 9-12): 本番対応
  - セキュリティ強化完了
  - パフォーマンス最適化完了
  - CI/CD自動化完了

- [ ] **Milestone 4** (Week 13-16): フロントエンド
  - Webアプリケーション完了

- [ ] **Milestone 5** (Week 17-18): ドキュメント整備（オプション）
  - ドキュメント作成の必要性を検討
  - 必要なドキュメントのみ作成

---

## 🎯 最終ゴール

**プロダクショングレードのTODO管理アプリケーション**

✅ **機能面**
- 個人・チームでの利用に対応
- リアルタイムコラボレーション
- モバイル対応

✅ **技術面**
- スケーラブルなアーキテクチャ
- 高いテストカバレッジ（80%以上）
- 包括的なドキュメント

✅ **運用面**
- 自動化されたCI/CD
- 監視・ロギング
- セキュリティ対策

---

## 📅 タイムライン（概算）

| フェーズ | 期間 | 工数（時間） | 備考 |
|---------|------|------------|------|
| Phase 1-2 | Week 1-4 | 80h | コア機能（リファクタリング、TODO拡張） |
| Phase 3-5 | Week 5-8 | 80h | 協働機能（検索、通知、共有） |
| Phase 6-7 | Week 9-12 | 60h | 本番対応（セキュリティ、CI/CD） |
| Phase 8 | Week 13-16 | 100h | フロントエンド（TypeScript、Next.js、テスト） |
| Phase 9 | Week 17-18 | 30-40h | ドキュメント（オプション） |
| **合計（Phase 9除く）** | **4ヶ月** | **320h** | - |
| **合計（Phase 9含む）** | **4.5ヶ月** | **350-360h** | - |

※ パートタイムで週20時間作業と仮定
※ Phase 8は型定義、テスト実装を含むため工数増
※ Phase 9はプロジェクトの必要性に応じて実施を判断

---

## 📚 参考リソース

### 学習リソース
- [Gin Framework Documentation](https://gin-gonic.com/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Go Best Practices](https://go.dev/doc/effective_go)
- [RealWorld Example Apps](https://github.com/gothinkster/realworld)

### サンプルプロジェクト
- [golang-gin-realworld-example-app](https://github.com/gothinkster/golang-gin-realworld-example-app)
- [go-clean-arch](https://github.com/bxcodec/go-clean-arch)

---

## 📝 補足事項

### ドキュメント作成について
Phase 9（ドキュメント作成）は、以下の状況で実施を検討します：

**実施を推奨するケース:**
- プロジェクトをオープンソースとして公開する
- チーム開発を行う（3人以上）
- 外部のコントリビューターを募集する
- 長期的なメンテナンスを想定している
- ユーザー向けのドキュメントが必要

**スキップまたは簡素化するケース:**
- 個人の学習プロジェクト
- 短期的なプロトタイプ
- 小規模なチーム（1-2人）
- コードの可読性が十分高い

Phase 9実施前（Week 16終了時点）に改めて検討し、必要なドキュメントのみを作成します。

---

**最終更新**: 2026-01-12
**ステータス**: Phase 0 (計画中)
**次のアクション**: Phase 1のリファクタリング開始
