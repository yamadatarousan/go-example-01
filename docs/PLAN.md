# go-example-01 プロジェクト拡張プラン

## 前提
- AIはexamplesディレクトリ配下しか手を付けない
- AIはプロジェクトディレクトリの直下を原則として編集しない(README.mdやPLAN.mdなどのドキュメントを除く)
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
- **📁 空ファイル作成ルール（写経用ディレクトリ）**:
  - `examples/` 配下にファイルやディレクトリを作成したら、**同じパスで空ファイルを写経用ディレクトリにも作成する**
  - 例: `examples/frontend/src/app/page.tsx` を作成 → `frontend/src/app/page.tsx`（空）も作成
  - 例: `examples/internal/handler/foo.go` を作成 → `internal/handler/foo.go`（空）も作成
  - 空ファイルにはコメントを入れず、完全に空の状態にする
  - 目的: ユーザーが写経すべきファイル構造を明確にするため
  - **⚠️ 例外: ツール生成ファイルはプレースホルダー不要**:
    - `create-next-app`が生成するファイル（`layout.tsx`, `page.tsx`等の初期ファイル）
    - `npx shadcn@latest add`が生成するファイル（`components/ui/*`）
    - これらはユーザーが同じコマンドを実行すれば生成されるため、空ファイルは不要
    - 代わりにPLAN.mdに実行すべきコマンドを記載する
  - **プレースホルダーを作成するもの**: AIが手書きするファイル（`types/index.ts`, `lib/server-api.ts`等）

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
- [x] **優先度**: `priority` (high/medium/low)
- [x] **期限**: `due_date` (TIMESTAMPTZ)
- [x] **ステータス**: `status` (todo/in_progress/done)
- [x] **カテゴリー**: `category_id` (FK to categories)
- [x] **タグ機能**: `tags` (many-to-many)
- [x] **サブタスク**: `parent_todo_id` (自己参照)

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
- [x] **ユーザー認証**
  - POST /signup - ユーザー登録
  - POST /login - ログイン
- [x] **TODO基本操作**
  - GET /api/v1/todos - TODO一覧取得
  - GET /api/v1/todos/:id - TODO詳細取得
  - POST /api/v1/todos - TODO作成（優先度、期限、ステータス、説明等を含む）
  - PUT /api/v1/todos/:id - TODO更新
  - DELETE /api/v1/todos/:id - TODO削除
- [x] **TODO拡張機能**（Phase 2で追加）
  - POST /api/v1/todos/:id/complete - TODO完了
  - POST /api/v1/todos/:id/reopen - TODO再開
  - GET /api/v1/todos/overdue - 期限切れTODO一覧
  - GET /api/v1/todos/today - 今日のTODO一覧
  - GET /api/v1/todos/week - 今週のTODO一覧
- [x] **カテゴリー機能**（Phase 2で追加）
  - POST /api/v1/categories - カテゴリー作成
  - GET /api/v1/categories - カテゴリー一覧取得
  - GET /api/v1/categories/:id - カテゴリー詳細取得
  - PUT /api/v1/categories/:id - カテゴリー更新
  - DELETE /api/v1/categories/:id - カテゴリー削除
- [x] **管理者機能**
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
- [x] リフレッシュトークンの実装
- [x] レート制限（100req/min）
- [x] セキュリティヘッダーの追加
- [x] 監査ログの拡充（全CRUD操作）
- [x] HTTPS強制

**6.2 パフォーマンス最適化**
- [x] データベースインデックスの追加
```sql
CREATE INDEX idx_todos_user_status ON todos(user_id, status);
CREATE INDEX idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL;
```
- [x] N+1クエリの排除
- [x] Redisキャッシュの導入
- [x] ページネーション改善

**6.3 監視・ロギング**
- [x] 構造化ロギング（zap）
- [x] Prometheusメトリクス
- [x] ヘルスチェックエンドポイント

#### 成果物
- [x] セキュリティ強化実装
- [x] パフォーマンス最適化
- [x] 監視基盤

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
- [x] マルチステージビルド
- [x] docker-compose.yml更新
- [x] 本番用Dockerfile

**7.3 自動デプロイ**
- [x] ステージング環境へのデプロイ設定
- [x] 本番環境へのデプロイ設定（手動承認）
- [x] ロールバック機能

#### 成果物
- [x] CI/CDパイプライン
- [x] 自動デプロイ環境

---

### Phase 8: フロントエンド開発（Week 13-22）

#### 概要
フロントエンド開発を10のサブフェーズに分割し、段階的に実装する。
各サブフェーズは約3日〜1週間で完了する小さな単位とし、毎回「動くもの」を確認できる形で進める。

#### 技術スタック
- **言語**: TypeScript
- **フレームワーク**: Next.js 14 (App Router)
- **UIライブラリ**: Tailwind CSS + shadcn/ui
- **フォーム**: React Hook Form + Zod
- **認証**: httpOnly Cookie + JWT
- **テスト**: Vitest + Playwright

#### フロントエンドアーキテクチャ

##### ディレクトリ構造の方針

Next.jsは「非固定的」（unopinionated）なフレームワークであり、`app/`内のルーティング規約以外は公式の「正解」がない。本プロジェクトでは**戦略A（app外配置）**を採用する。

**Next.js公式が紹介する3つの戦略:**

| 戦略 | 構造 | 特徴 |
|------|------|------|
| **A** | `app/`外に`components/`, `lib/`, `types/` | ルーティングとロジックの分離 |
| **B** | `app/`内に`_components/`, `_lib/` | 関連ファイルが近い |
| **C** | 機能ごとに分割 + ルートグループ`()` | 大規模向け、スケーラブル |

**採用理由:**
- TODOアプリの規模では戦略Aで十分
- ルーティング（`app/`）とビジネスロジック（`lib/`, `components/`）が明確に分離される
- shadcn/uiのデフォルト構成と一致

**戦略Aの背景:**
- Next.js 13以前は`pages/`がルーティング専用で、`pages/`外に`components/`等を置くのが自然だった
- App Router導入後もその慣習が継続している
- `app/`内には特殊ファイル規約があり、ビジネスロジックを混在させると分かりにくくなる

**設計思想の違い:**

| アプローチ | 戦略 | 考え方 |
|-----------|------|--------|
| **レイヤー分離** | A | 「種類」で分ける（全コンポーネント、全型定義をまとめる） |
| **コロケーション** | B, C | 「機能」で分ける（関連ファイルを近くに置く） |

##### プロジェクト構造

```
frontend/src/
├── app/                    # ルーティング専用（Next.js規約）
│   ├── layout.tsx          # ルートレイアウト
│   ├── page.tsx            # トップページ
│   ├── globals.css         # グローバルスタイル
│   ├── loading.tsx         # ローディングUI
│   ├── error.tsx           # エラーUI
│   ├── not-found.tsx       # 404 UI
│   ├── login/              # /login
│   ├── signup/             # /signup
│   ├── todos/              # /todos
│   └── ...
├── components/             # 共有コンポーネント
│   └── ui/                 # shadcn/uiコンポーネント
├── lib/                    # ユーティリティ関数
│   └── utils.ts            # cn関数等
└── types/                  # TypeScript型定義
    └── index.ts            # 共通型
```

##### Next.js App Router の規約

| ファイル | 用途 |
|---------|------|
| `layout.tsx` | 共有レイアウト（ヘッダー、ナビ等） |
| `page.tsx` | ページ本体（これがあるとルートが公開される） |
| `loading.tsx` | ローディング状態（React Suspense） |
| `error.tsx` | エラーバウンダリ |
| `not-found.tsx` | 404ページ |
| `route.ts` | APIエンドポイント（本プロジェクトでは使用しない） |

##### 特殊なフォルダ命名規則

| 記法 | 用途 | 例 |
|------|------|-----|
| `_folder` | プライベートフォルダ（ルーティング対象外） | `_components/` |
| `(folder)` | ルートグループ（URLに影響しない） | `(auth)/login` → `/login` |
| `[folder]` | 動的ルート | `[id]` → `/todos/123` |
| `[[...folder]]` | オプショナルキャッチオール | `[[...slug]]` |

---

### Phase 8.1: プロジェクト初期化（3日）

#### 目標
Next.js プロジェクトを作成し、開発環境を整える

#### タスク
- [x] Next.js プロジェクト作成（create-next-app）
- [x] TypeScript 設定確認
- [x] ESLint 設定
- [x] Tailwind CSS 設定
- [x] ディレクトリ構造の作成

#### 成果物
- [x] `npm run dev` で起動できるプロジェクト

#### 写経用ディレクトリ（frontend/）の初期化手順

AIは`examples/frontend/`のみを編集する。写経用の`frontend/`は以下の手順で初期化する。

**1. プロジェクト作成**
```bash
# プロジェクトルート（go-example-01/）で実行
npx create-next-app@latest frontend
```

create-next-appの質問への回答:
```
✔ Would you like to use the recommended Next.js defaults? › No, customize settings
✔ Would you like to use TypeScript? … Yes
✔ Which linter would you like to use? › ESLint
✔ Would you like to use React Compiler? … No
✔ Would you like to use Tailwind CSS? … Yes
✔ Would you like your code inside a `src/` directory? … Yes
✔ Would you like to use App Router? (recommended) … Yes
✔ Would you like to customize the import alias (`@/*` by default)? … No
```

**2. ディレクトリ構造の作成**
```bash
mkdir -p frontend/src/{components,lib,types}
```

**3. 動作確認**
```bash
cd frontend && npm run dev
```

---

### Phase 8.2: UIコンポーネント基盤（3日）

#### 目標
共通UIコンポーネントと型定義を整備する

#### タスク
- [x] shadcn/ui セットアップ
- [x] 基本コンポーネント追加（Button, Input, Label, Card）
- [x] lib/utils.ts（cn関数）
- [x] types/index.ts（Todo, User, ApiError等）
- [x] 基本UIコンポーネントのUIテスト（Button/Input/Label/Card）

#### 成果物
- [x] 共通コンポーネントが使える状態

#### 写経用ディレクトリ（frontend/）のセットアップ手順

**1. shadcn/ui 初期化**
```bash
cd frontend
npx shadcn@latest init
```

質問への回答:
```
✔ Which color would you like to use as the base color? › Neutral
✔ Would you like to use CSS variables for theming? › Yes
```

**2. 基本コンポーネント追加**
```bash
npx shadcn@latest add button input label card
```

---

### Phase 8.3: API通信基盤（3日）

#### 目標
バックエンドとの通信基盤を整備する

#### タスク
- [x] lib/server-api.ts（認証付きfetch関数）
- [x] fetchWithAuth / fetchWithoutAuth
- [x] ActionResult型の定義
- [x] モックデータ切り替え機能
- [x] lib/server-api.ts のユニットテスト（認証あり/なし、エラー分岐）

#### 成果物
- [x] Server ActionsからAPI呼び出しができる状態

#### 写経用ディレクトリ（frontend/）のセットアップ手順

**環境変数ファイルの作成:**
```bash
# frontend/.env.local を作成
echo 'API_BASE_URL=http://localhost:8080' > frontend/.env.local
echo 'USE_MOCK=false' >> frontend/.env.local
```

---

### Phase 8.4: ログイン画面（3日）

#### 目標
ログイン画面を実装する

#### タスク
- [x] app/login/page.tsx
- [x] ログインフォーム（React Hook Form + Zod）
- [x] app/login/actions.ts（Server Action）
- [x] Cookie へのJWT保存（actions.ts内で実装）
- [x] login-form.tsxとServer Actionの連携
- [x] ログインフォームの表示/挙動テスト（成功・失敗含む）
- [x] ログイン単体のE2Eテスト（ログイン→/todos遷移）

#### 成果物
- [x] ログインフォームが表示され、送信できる

---

### Phase 8.4.2: 起動時のテーブル存在チェック/自動マイグレーション（1日）

#### 目標
本番/開発でDBテーブル不足による500を防ぐ

#### タスク
- [x] 起動時に必須テーブルの存在チェックを行う
- [x] AUTO_MIGRATE=true のときだけ起動時にマイグレーション実行
- [x] 不足時は起動エラーにして早期検知
- [x] READMEに運用方針（本番は存在チェック、開発は自動マイグレーション可）を追記

#### 成果物
- [x] 起動時テーブル存在チェックが有効
- [x] 開発環境で自動マイグレーションが動作

---

### Phase 8.4.1: フロントエンドテスト基盤（1日）

#### 目標
各フェーズの実装に合わせてすぐにテストを書ける状態にする

#### タスク
- [x] jsdom + @testing-library/react + @testing-library/jest-dom のセットアップ
- [x] Vitest設定（jsdom環境）と共通テストユーティリティの追加
- [x] E2Eテスト基盤（Playwright）のセットアップ

#### 成果物
- [x] UIテスト（jsdom + jest-dom）が動作する
- [x] E2Eテスト（Playwright）が動作する

#### 写経用ディレクトリ（frontend/）のセットアップ手順

**パッケージのインストール:**
```bash
cd frontend
npm install react-hook-form zod @hookform/resolvers
```

---

### Phase 8.5: サインアップ・認証フロー（3日）

#### 目標
サインアップとルート保護を実装する

#### タスク
- [x] app/signup/page.tsx（ページ表示）
- [ ] サインアップフォーム（UI + バリデーション）
- [ ] app/signup/actions.ts（Server Action）
- [ ] 認証後のリダイレクト処理（成功時の導線）
- [ ] ルート保護 middleware.ts
  - ログイン必須: /todos, /dashboard, /projects, /settings
  - 未ログイン専用（ログイン済みは /todos へ）: /login, /signup
  - 公開: /
- [ ] サインアップフォームの表示/挙動テスト（成功・失敗：重複ユーザー含む）
- [ ] 認証フローE2E（サインアップ→ログイン→/todos）

#### 成果物
- [ ] サインアップ → ログイン → TODO一覧へのフローが動く

---

### Phase 8.6: TODO一覧表示（3日）

#### 目標
TODO一覧を表示する

#### タスク
- [ ] app/todos/page.tsx（Server Component）
- [ ] TODOアイテムコンポーネント
- [ ] ヘッダーコンポーネント
- [ ] app/todos/actions.ts（取得用）
- [ ] TODO一覧のUIテスト

#### 成果物
- [ ] バックエンドからTODOを取得して一覧表示できる

---

### Phase 8.7: TODO CRUD（3日）

#### 目標
TODOの追加・更新・削除を実装する

#### タスク
- [ ] TODO追加フォーム
- [ ] TODO更新（完了チェック、タイトル編集）
- [ ] TODO削除
- [ ] Server Actions（create, update, delete）
- [ ] TODO CRUDのUIテスト
- [ ] TODO CRUDのE2Eテスト

#### 成果物
- [ ] TODOのCRUD操作が全て動く

---

### Phase 8.8: エラー・ローディング状態（3日）

#### 目標
エラーハンドリングとローディング状態を統一する

#### タスク
- [ ] app/error.tsx（グローバルエラー）
- [ ] app/not-found.tsx（404）
- [ ] app/loading.tsx
- [ ] components/ui/skeleton.tsx
- [ ] sonner セットアップ（トースト通知）
- [ ] layout.tsx に Toaster 追加
- [ ] エラー/ローディング表示のUIテスト

#### 成果物
- [ ] エラー・ローディング状態が統一パターンで表示される

---

### Phase 8.9: ダッシュボード（3日）

#### 目標
統計情報を表示するダッシュボードを実装する

#### タスク
- [ ] app/dashboard/page.tsx
- [ ] 統計カードコンポーネント
- [ ] 今日のTODO表示
- [ ] 期限切れTODO表示
- [ ] app/dashboard/actions.ts
- [ ] ダッシュボードのUIテスト

#### 成果物
- [ ] ダッシュボードに統計情報が表示される

---

### Phase 8.10: プロジェクト・設定・仕上げ（1週間）

#### 目標
残りの画面を実装し、全体を仕上げる

#### タスク
- [ ] プロジェクト一覧・詳細画面
- [ ] プロジェクト作成・編集・削除
- [ ] メンバー管理
- [ ] 設定画面（プロフィール表示）
- [ ] ログアウト機能
- [ ] 共通ナビゲーション整備
- [ ] レスポンシブ対応（PC、タブレット、スマートフォン）
- [ ] 主要画面のUIテスト追加
- [ ] 主要フローのE2Eテスト追加

#### 成果物
- [ ] 全画面が揃い、レスポンシブ対応が完了

---

#### Phase 8 全体の成果物
- [ ] TypeScriptベースのフロントエンドアプリケーション
- [ ] 型安全なAPI通信層
- [ ] 再利用可能なコンポーネント
- [ ] 統一されたエラーハンドリング・ローディング状態

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

## ⏸ 後回しタスク（未着手）

Phase 6で言及したが、現状は未着手のため後回しにする。

- [ ] Redisキャッシュの導入
- [ ] 構造化ロギング（zap）
- [ ] Prometheusメトリクス
- [ ] HTTPS強制
- [ ] /admin のフロントエンド実装・ルート保護（管理者ロール対応）

---

**最終更新**: 2026-01-27
**完了ステータス**: Phase 8.3まで完了
**次のアクション**: Phase 8.4のタスクを進める
