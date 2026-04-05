package myfunc

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func FindRes(db *sql.DB, keywords []string) error {
	if db == nil {
        return fmt.Errorf("kết nối database không hợp lệ")
    }
    if len(keywords) == 0 {
        return fmt.Errorf("danh sách từ khóa trống")
    }
    for i := range keywords {
        keywords[i] = strings.ToLower(keywords[i])
    }
    query := `SELECT id,name, path FROM resources WHERE name LIKE ? OR path LIKE ?`

    for _, kw := range keywords {
        // Thêm wildcard % để tìm gần giống
        pattern := "%" + kw + "%"

        rows, err := db.Query(query, pattern, pattern)
        if err != nil {
            return fmt.Errorf("lỗi truy vấn: %v", err)
        }
        defer rows.Close()

        for rows.Next() {
            var name, path string
            var id int
            if err := rows.Scan(&id,&name, &path); err != nil {
                return fmt.Errorf("lỗi đọc dữ liệu: %v", err)
            }
            fmt.Printf("id:%d |Tên: %s | Đường dẫn: %s\n",id, name, path)
        }
    }
    return nil
}
