# go-example-01

## データベース運用手順

このプロジェクトでは、データベースとして **PostgreSQL** を、スキーマ管理ツールとして **golang-migrate/migrate** を使用します。

### 1. 初回セットアップ

1.  **ツールのインストール (macOSの場合):**
    ```bash
    brew install golang-migrate
    ```

2.  **データベースの起動:**
    プロジェクトのルートディレクトリでDockerコンテナを起動します。
    ```bash
    docker-compose up -d
    ```

### 2. マイグレーション操作

マイグレーションコマンドは、プロジェクトのルートディレクトリで実行します。

-   **最新バージョンまで適用する (`up`)**
    未適用のマイグレーションをすべて実行し、データベースを最新の状態にします。
    ```bash
    migrate -database "postgres://user:password@localhost:5433/todo_db?sslmode=disable" -path go/db/migrations up
    ```

-   **1つ前のバージョンに戻す (`down`)**
    最後に適用したマイグレーションを1つだけ取り消します。
    ```bash
    migrate -database "postgres://user:password@localhost:5433/todo_db?sslmode=disable" -path go/db/migrations down 1
    ```

-   **現在のバージョンを確認する (`version`)**
    現在データベースに適用されているマイグレーションのバージョンを確認します。
    ```bash
    migrate -database "postgres://user:password@localhost:5433/todo_db?sslmode=disable" -path go/db/migrations version
    ```

-   **新しいマイグレーションファイルを作成する**
    スキーマを変更する際は、以下のコマンドで新しい`up`/`down`ファイルを作成します。
    ```bash
    migrate create -ext sql -dir go/db/migrations -seq [変更内容の短い説明]
    ```
    例: `migrate create -ext sql -dir go/db/migrations -seq add_user_table`

-   **`dirty`状態の修正（緊急時）**
    マイグレーションが途中で失敗すると、データベースが`dirty`状態になることがあります。その際は、`force`コマンドで特定のクリーンなバージョンに強制的に設定し直します。
    ```bash
    # 例: バージョン3の状態に強制的に戻す
    migrate -database "postgres://user:password@localhost:5433/todo_db?sslmode=disable" -path go/db/migrations force 3
    ```

### 3. データベースへの直接接続

`psql`を使い、コンテナ内のデータベースに直接接続してデータを参照・操作できます。

```bash
docker-compose exec db psql -U user -d todo_db
```
- `\dt`: テーブル一覧を表示
- `SELECT * FROM todos;`: `todos`テーブルの内容を表示
- `\q`: `psql`を終了

---