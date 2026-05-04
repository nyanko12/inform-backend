CREATE TABLE IF NOT EXISTS products (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_code VARCHAR,
  name VARCHAR NOT NULL,
  genre VARCHAR,
  image_url VARCHAR,
  content_volume FLOAT,
  content_unit VARCHAR,
  days_to_consume INT NOT NULL,
  registered_date DATE NOT NULL,
  next_due_date DATE NOT NULL,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_user_id ON products(user_id);
CREATE INDEX IF NOT EXISTS idx_products_next_due_date ON products(next_due_date);
