package myfunc

import (
	"database/sql"
	"fmt"
	"os/exec"
	"runtime"
)

func OpenURL(db *sql.DB, webID int) error {
	var name, url string
	// Lấy ra name và url từ DB
	err := db.QueryRow("SELECT name, path FROM resources WHERE id = ?", webID).Scan(&name, &url)
	if err != nil {
		return fmt.Errorf("Không tìm thấy website với id %d: %v", webID, err)
	}
	// Tạo lệnh mở URL tùy hệ điều hành
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	default: // Linux
		cmd = exec.Command("xdg-open", url)
	}
	// Thực thi
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Không thể mở website %s: %v", name, err)
	}
	fmt.Printf("Đang mở website %s...\n", name)
	return nil
}
