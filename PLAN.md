# go-example-01 プロジェクト拡張プラン

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

## 📈 拡張ロードマップ

### Phase 1: コード構造のリファクタリング（Week 1-2）

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

**1.2 レイヤー分離の実装**
- [ ] Domain層: エンティティとビジネスルール
- [ ] Repository層: データベース操作の抽象化
- [ ] Service層: ビジネスロジックの集約
- [ ] Handler層: HTTPリクエスト/レスポンス処理
- [ ] インターフェースによる依存性注入

**1.3 設定管理の改善**
- [ ] 環境変数の一元管理
- [ ] 開発/本番環境の切り替え
- [ ] `.env`ファイルのサポート（`godotenv`）

#### 成果物
- [ ] リファクタリング後のコード
- [ ] アーキテクチャ図
- [ ] 移行ガイドドキュメント

---

### Phase 2: TODO機能の拡張（Week 3-4）

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

#### 成果物
- [ ] 拡張されたTODOモデル
- [ ] カテゴリー・タグ機能の実装
- [ ] 新しいエンドポイントのテスト

---

### Phase 3: 検索・フィルタリング機能（Week 5）

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
- [ ] 検索・フィルタリング実装
- [ ] 全文検索機能
- [ ] 統計情報API

---

### Phase 4: 通知・リマインダー機能（Week 6）

#### 目標
期限管理を支援する通知機能

#### タスク

**4.1 通知モデル**
```sql
-- 000011_create_notifications_table.up.sql
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
```

**4.2 リマインダー設定**
```sql
-- 000012_create_reminders_table.up.sql
CREATE TABLE IF NOT EXISTS reminders (
    id SERIAL PRIMARY KEY,
    todo_id INT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    remind_at TIMESTAMPTZ NOT NULL,
    is_sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**4.3 通知エンドポイント**
```
GET    /api/v1/notifications              # 通知一覧
GET    /api/v1/notifications/unread       # 未読通知
PUT    /api/v1/notifications/:id/read     # 既読にする
PUT    /api/v1/notifications/read-all     # 全て既読
DELETE /api/v1/notifications/:id          # 通知削除
```

**4.4 バックグラウンドワーカー**
- [ ] 定期的に期限をチェックするジョブ
- [ ] リマインダー送信処理
- [ ] 通知の生成

#### 成果物
- [ ] 通知システム
- [ ] リマインダー機能
- [ ] バックグラウンドジョブ

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
- [ ] プロジェクト機能
- [ ] メンバー管理
- [ ] TODO担当機能
- [ ] コメント機能

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

### Phase 7: CI/CD・ドキュメント（Week 11-12）

#### 目標
自動化と開発者体験の向上

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

**7.2 OpenAPI仕様書**
- [ ] Swaggerアノテーションの追加
- [ ] Swagger UI (`/api/docs`)
- [ ] Postmanコレクション

**7.3 Docker最適化**
- [ ] マルチステージビルド
- [ ] docker-compose.yml更新
- [ ] 本番用Dockerfile

**7.4 ドキュメント整備**
- [ ] ARCHITECTURE.md
- [ ] API.md
- [ ] DEPLOYMENT.md
- [ ] CONTRIBUTING.md

#### 成果物
- [ ] CI/CDパイプライン
- [ ] API仕様書
- [ ] 包括的なドキュメント

---

### Phase 8: フロントエンド開発（Week 13-16）

#### 目標
Webフロントエンドの実装

#### タスク

**8.1 技術選定**
- フレームワーク: React / Next.js
- 状態管理: Zustand / Redux
- UIライブラリ: Tailwind CSS
- HTTP Client: Axios

**8.2 主要画面**
- [ ] ログイン・サインアップ
- [ ] ダッシュボード
- [ ] TODO一覧・詳細
- [ ] プロジェクト管理
- [ ] 設定画面

**8.3 レスポンシブデザイン**
- [ ] PC対応
- [ ] タブレット対応
- [ ] スマートフォン対応

#### 成果物
- [ ] フロントエンドアプリケーション
- [ ] E2Eテスト

---

## 🛠 技術スタックの拡張

### 追加予定の技術

| カテゴリ | 技術 | 用途 |
|---------|------|------|
| **ORM** | GORM | データベース抽象化 |
| **キャッシュ** | Redis | セッション、キャッシュ |
| **ログ** | zap | 構造化ログ |
| **メトリクス** | Prometheus | 監視 |
| **ドキュメント** | swag | OpenAPI/Swagger |
| **設定管理** | viper | 環境変数 |
| **タスクキュー** | asynq | バックグラウンドジョブ |

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
  - CI/CD・ドキュメント完了

- [ ] **Milestone 4** (Week 13-16): フロントエンド
  - Webアプリケーション完了

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

| フェーズ | 期間 | 工数（時間） |
|---------|------|------------|
| Phase 1-2 | Week 1-4 | 80h |
| Phase 3-5 | Week 5-8 | 80h |
| Phase 6-7 | Week 9-12 | 60h |
| Phase 8 | Week 13-16 | 80h |
| **合計** | **4ヶ月** | **300h** |

※ パートタイムで週20時間作業と仮定

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

**最終更新**: 2026-01-12
**ステータス**: Phase 0 (計画中)
**次のアクション**: Phase 1のリファクタリング開始
