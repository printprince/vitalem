-- Добавляем поле metadata для хранения дополнительных данных уведомлений
ALTER TABLE notifications ADD COLUMN metadata JSONB; 