package myfunc

import (
	"database/sql"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// OpenApps mở ứng dụng theo đường dẫn/tên, trả về error nếu có lỗi
func OpenApps(db *sql.DB, AppID int) error {
	var name, path string
	// Lấy ra name và path từ DB
	err := db.QueryRow("SELECT name, path FROM resources WHERE id = ?", AppID).Scan(&name, &path)
	if err != nil {
		return fmt.Errorf("Không tìm thấy ứng dụng với id %d: %v", AppID, err)
	}

	// Tạo lệnh mở app tùy theo hệ điều hành
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command(path)
	case "darwin":
		// Nếu là .app bundle, dùng "open -a"
		if strings.HasSuffix(path, ".app") {
			cmd = exec.Command("open", "-a", path)
		} else {
			cmd = exec.Command("open", path)
		}
	default: // Linux và các OS khác
		cmd = exec.Command(path)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Không thể mở ứng dụng %s: %v", name, err)
	}
	fmt.Printf("Đang mở ứng dụng %s...\n", name)
	return nil
}

