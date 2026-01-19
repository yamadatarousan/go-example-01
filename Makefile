# ============================================================================
# Makefile - ビルド・テスト自動化
# ============================================================================
# このMakefileは、よく使うコマンドをまとめたものです。
# 「make コマンド名」で実行できます。
#
# 使い方:
#   make test         - テストを実行
#   make build        - バイナリをビルド
#   make run          - アプリケーションを起動
#   make lint         - 静的解析を実行
#   make clean        - ビルド成果物を削除
#   make migrate-up   - マイグレーションを適用
#   make migrate-down - マイグレーションをロールバック
#   make docker-build - Dockerイメージをビルド
#   make docker-run   - Dockerコンテナを起動
#
# 注意: Makefileではインデントに「タブ」を使う必要があります（スペース不可）
# ============================================================================

# .PHONY宣言
# これらのターゲットは「ファイル名」ではなく「コマンド名」であることを宣言
# 同名のファイルが存在しても、常にコマンドとして実行される
.PHONY: test build run lint clean migrate-up migrate-down docker-build docker-run

# ============================================================================
# テスト関連
# ============================================================================

# テストの実行
# 1. docker-composeでテスト用PostgreSQLを起動（--wait: 起動完了まで待機）
# 2. go testでテストを実行
#    -v: 詳細出力（各テスト名と結果を表示）
#    -race: 競合状態の検出（並行処理のバグを検出）
# 3. テスト終了後、コンテナとボリュームを削除（-v: ボリュームも削除）
test:
	docker-compose -f docker-compose.test.yml up -d --wait
	go test -v -race ./cmd/api/...
	docker-compose -f docker-compose.test.yml down -v

# ============================================================================
# ビルド関連
# ============================================================================

# バイナリのビルド
# -o bin/api: 出力先を bin/api に指定
# ./cmd/api:  ビルド対象のパッケージ（main関数があるディレクトリ）
build:
	go build -o bin/api ./cmd/api

# 開発用: アプリケーションの起動
# go run はビルドと実行を一度に行う（開発時に便利）
# 本番では build で作成したバイナリを使う
run:
	go run ./cmd/api

# ============================================================================
# コード品質
# ============================================================================

# 静的解析（リンター）の実行
# golangci-lint: 複数のリンターを統合したツール
# 事前にインストールが必要: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
# ./... は現在のディレクトリ以下の全パッケージを対象にする
lint:
	golangci-lint run ./...

# ビルド成果物の削除
# rm -rf: 再帰的に強制削除
# bin/ ディレクトリごと削除
clean:
	rm -rf bin/

# ============================================================================
# データベースマイグレーション
# ============================================================================
# 事前にmigrateツールのインストールが必要:
#   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
#
# 使用例:
#   DATABASE_URL="postgres://user:password@localhost:5432/todo_db?sslmode=disable" make migrate-up

# マイグレーションの適用（データベーススキーマを最新にする）
# $(DATABASE_URL): 環境変数 DATABASE_URL の値を使用
# -path db/migrations: マイグレーションファイルのディレクトリ
# up: 未適用のマイグレーションを全て適用
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

# マイグレーションのロールバック（1つ前の状態に戻す）
# down: 最新のマイグレーションを1つ取り消す
# 全て戻す場合は: migrate -path db/migrations -database "$(DATABASE_URL)" down -all
migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down

# ============================================================================
# Docker関連
# ============================================================================

# Dockerイメージのビルド
# -t gin-quickstart:latest: イメージ名とタグを指定
# . : 現在のディレクトリをビルドコンテキストとして使用（Dockerfileの場所）
docker-build:
	docker build -t gin-quickstart:latest .

# Dockerコンテナの起動
# -p 8080:8080: ホストのポート8080をコンテナのポート8080にマッピング
# --env-file .env: .envファイルから環境変数を読み込む
#                  （DB接続情報などを設定）
# gin-quickstart:latest: 起動するイメージ名
docker-run:
	docker run -p 8080:8080 --env-file .env gin-quickstart:latest
