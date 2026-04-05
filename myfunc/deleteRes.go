package myfunc

import (
	"database/sql"
	"fmt"
)

// Xóa resource theo ID
func DeleteResourceByID(db *sql.DB, id int) error {
	query := `DELETE FROM resources WHERE id = ?`
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("lỗi khi xóa resource ID %d: %w", id, err)
	}
	fmt.Printf("✅ Resource ID %d đã được xóa.\n", id)
	return nil
}

// Xóa resource theo tên
func DeleteResourceByName(db *sql.DB, name string) error {
	query := `DELETE FROM resources WHERE name = ?`
	res, err := db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("lỗi khi xóa resource '%s': %w", name, err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		fmt.Printf("⚠️ Không tìm thấy resource có tên '%s'.\n", name)
	} else {
		fmt.Printf("✅ Resource '%s' đã được xóa.\n", name)
	}
	return nil
}
