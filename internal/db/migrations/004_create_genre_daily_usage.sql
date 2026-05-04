CREATE TABLE IF NOT EXISTS genre_daily_usage (
  id SERIAL PRIMARY KEY,
  genre VARCHAR NOT NULL,
  unit VARCHAR NOT NULL,
  daily_usage_per_person FLOAT NOT NULL
);

INSERT INTO genre_daily_usage (genre, unit, daily_usage_per_person) VALUES
  ('シャンプー', 'ml', 10),
  ('ボディソープ', 'ml', 15),
  ('歯磨き粉', 'g', 2),
  ('トイレットペーパー', 'ロール', 0.3),
  ('洗濯洗剤（液体）', 'ml', 30),
  ('食器用洗剤', 'ml', 5),
  ('ティッシュ', '枚', 10)
ON CONFLICT DO NOTHING;
