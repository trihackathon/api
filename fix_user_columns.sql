-- 既存のユーザーテーブルのカラムを修正
-- NOT NULL制約を削除し、デフォルト値を設定

-- genderカラムが存在する場合、制約を変更
ALTER TABLE users 
ALTER COLUMN gender DROP NOT NULL,
ALTER COLUMN gender SET DEFAULT 'other';

-- weightカラムが存在する場合、制約を変更
ALTER TABLE users 
ALTER COLUMN weight DROP NOT NULL,
ALTER COLUMN weight SET DEFAULT 60;

-- chronotypeカラムが存在する場合、制約を変更
ALTER TABLE users 
ALTER COLUMN chronotype DROP NOT NULL,
ALTER COLUMN chronotype SET DEFAULT 'both';

-- 既存のNULLレコードにデフォルト値を設定
UPDATE users SET gender = 'other' WHERE gender IS NULL;
UPDATE users SET weight = 60 WHERE weight IS NULL;
UPDATE users SET chronotype = 'both' WHERE chronotype IS NULL;
