package myfunc

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Lưu web vào DB (is_web = true)
func SaveWebToDB(db *sql.DB, name string, path string) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	_, err := db.Exec(`
        INSERT INTO resources (name, path, is_web)
        VALUES (?, ?, 1)
        ON CONFLICT(name) DO UPDATE SET
            path = excluded.path,
            is_web = 1;
    `, name, path)
	return err
}

// Lưu nhiều URL vào DB
func SaveAllWebsByInput(db *sql.DB, urls map[string]string) error {
	for name, path := range urls {
		if err := SaveWebToDB(db, name, path); err != nil {
			log.Printf("Save web failed (%s): %v", name, err)
		}
	}
	return nil
}

// Lấy tất cả webs từ DB (chỉ chọn is_web = 1)
func GetWebsFromDB(db *sql.DB) (map[string]string, error) {
	webs := make(map[string]string)

	rows, err := db.Query("SELECT name, path FROM resources WHERE is_web = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, path string
		if err := rows.Scan(&name, &path); err != nil {
			return nil, err
		}
		webs[name] = path
	}

	// kiểm tra lỗi khi duyệt rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return webs, nil
}
