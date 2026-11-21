package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir := "migrations"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}

	files := map[string]string{
		"000001_create_movies.up.sql": `CREATE TABLE IF NOT EXISTS movies (
    id bigint PRIMARY KEY,
    title text NOT NULL,
    year integer NOT NULL,
    director text NOT NULL
);
`,
		"000001_create_movies.down.sql": `DROP TABLE IF EXISTS movies;
`,
	}

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			panic(err)
		}
		fmt.Println("created:", path)
	}
}
