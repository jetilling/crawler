CREATE TABLE sites (
  id SERIAL PRIMARY KEY,
  link TEXT UNIQUE,
  status_code INTEGER,
  title TEXT,
  last_updated TIMESTAMP DEFAULT now()
);