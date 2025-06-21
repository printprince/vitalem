-- Удаляем поле metadata из таблицы notifications
ALTER TABLE notifications DROP COLUMN IF EXISTS metadata; 