CREATE TABLE IF NOT EXISTS movies (
    id bigint PRIMARY KEY,
    title text NOT NULL,
    year integer NOT NULL,
    director text NOT NULL
);
