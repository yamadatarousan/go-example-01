# go-example-01

### データベースへの直接接続

`psql`を使い、コンテナ内のデータベースに直接接続してデータを参照・操作できます。

```bash
docker-compose exec db psql -U user -d todo_db
```
- `\dt`: テーブル一覧を表示
- `SELECT * FROM todos;`: `todos`テーブルの内容を表示
- `\q`: `psql`を終了

---