# OpenAPI 運用ガイド

## 目的
- OpenAPI を単一の契約ソースとし、フロント/バックの型と検証を同期させる
- 変更時に「何を更新すべきか」「どこに反映されるか」を明確にする

## 主要ファイル
- 仕様: `openapi/openapi.yaml`
- Go 生成物: `backend/internal/openapi/gen/openapi.gen.go`
- TS 生成物: `frontend/src/openapi/types.ts`
- Lint 設定: `openapi/redocly.yaml`

## 変更手順（やりたいこと別）

### エンドポイントを追加/変更したい
1. `openapi/openapi.yaml` の `paths` を更新
2. Go/TS 生成を実行
3. 生成物の差分を確認

### スキーマ（型）を変更したい
1. `openapi/openapi.yaml` の `components.schemas` を更新
2. Go/TS 生成を実行
3. 影響範囲（handler/FE）を更新

### Nullable 方針を変更したい
1. `openapi/openapi.yaml` の該当フィールドを更新
2. Go/TS 生成を実行
3. レスポンス検証で落ちないことを確認

## 生成コマンド
```bash
# Go 型生成
$(go env GOPATH)/bin/oapi-codegen --config openapi/oapi-codegen.yaml openapi/openapi.yaml

# TS 型生成
./frontend/node_modules/.bin/openapi-typescript openapi/openapi.yaml -o frontend/src/openapi/types.ts
```

## Lint コマンド
```bash
npx @redocly/cli@latest lint --config openapi/redocly.yaml openapi/openapi.yaml
```

## 内部的な動作の要点

### ルーティング
- `openapi/openapi.yaml` の `paths` が `RegisterHandlersWithOptions` のルート定義に変換される
- `/api/v1/todos` は OpenAPI 生成のルート登録のみで管理

### リクエスト検証
- `openapi` 仕様に基づいたミドルウェアで **body/params を検証**
- 検証エラーは `400` で返す
- `OPENAPI_DEBUG=true` のときのみ `error_source: "openapi"` を付与

### レスポンス検証
- `/api/v1/todos` 系のみ対象
- レスポンス全体をバッファして検証
- 契約違反は `500` を返し、ログに詳細を残す

## 運用ルール
- 生成物配下は手編集禁止
- OpenAPI の変更時は必ず生成物を更新
- CI で lint と差分検知を行う
