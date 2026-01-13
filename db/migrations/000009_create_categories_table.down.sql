-- カテゴリーテーブルのロールバック

-- TODOテーブルからカテゴリー外部キーを削除
DROP INDEX IF EXISTS idx_todos_category;
ALTER TABLE todos DROP COLUMN IF EXISTS category_id;

-- カテゴリーテーブルとインデックスを削除
DROP INDEX IF EXISTS idx_categories_user_name;
DROP TABLE IF EXISTS categories;
