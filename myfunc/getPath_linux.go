//go:build linux

package myfunc

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// GetInstalledResourcesByName quét các file .desktop trong thư mục applications
// để tìm các ứng dụng đã cài đặt trên Linux
func GetInstalledResourcesByName() (map[string]string, error) {
	resources := make(map[string]string)

	// Các thư mục chứa .desktop files trên Linux
	desktopDirs := []string{
		"/usr/share/applications",
		"/usr/local/share/applications",
	}

	// Thêm ~/.local/share/applications nếu có
	homeDir, err := os.UserHomeDir()
	if err == nil {
		desktopDirs = append(desktopDirs, filepath.Join(homeDir, ".local", "share", "applications"))
	}

	// Thêm thư mục từ $XDG_DATA_DIRS nếu có
	xdgDataDirs := os.Getenv("XDG_DATA_DIRS")
	if xdgDataDirs != "" {
		for _, dir := range strings.Split(xdgDataDirs, ":") {
			appDir := filepath.Join(dir, "applications")
			desktopDirs = append(desktopDirs, appDir)
		}
	}

	// Loại bỏ trùng lặp
	seen := make(map[string]bool)
	uniqueDirs := []string{}
	for _, dir := range desktopDirs {
		if !seen[dir] {
			seen[dir] = true
			uniqueDirs = append(uniqueDirs, dir)
		}
	}

	for _, dir := range uniqueDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}

			filePath := filepath.Join(dir, entry.Name())
			name, execPath := parseDesktopFile(filePath)
			if name != "" && execPath != "" {
				resources[strings.ToLower(name)] = execPath
			}
		}
	}

	// Quét thêm các binary trong /usr/bin, /usr/local/bin (các app phổ biến)
	commonBinDirs := []string{"/usr/bin", "/usr/local/bin"}
	for _, binDir := range commonBinDirs {
		entries, err := os.ReadDir(binDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := strings.ToLower(entry.Name())
			if _, exists := resources[name]; !exists {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				// Chỉ thêm file có quyền thực thi
				if info.Mode()&0111 != 0 {
					resources[name] = filepath.Join(binDir, entry.Name())
				}
			}
		}
	}

	return resources, nil
}

// parseDesktopFile đọc file .desktop và trả về (Name, Exec path)
func parseDesktopFile(path string) (string, string) {
	file, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer file.Close()

	var name, execPath string
	inDesktopEntry := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Chỉ đọc section [Desktop Entry]
		if line == "[Desktop Entry]" {
			inDesktopEntry = true
			continue
		}
		if strings.HasPrefix(line, "[") && line != "[Desktop Entry]" {
			if inDesktopEntry {
				break // Đã qua section khác
			}
			continue
		}

		if !inDesktopEntry {
			continue
		}

		if strings.HasPrefix(line, "Name=") && name == "" {
			name = strings.TrimPrefix(line, "Name=")
		}
		if strings.HasPrefix(line, "Exec=") && execPath == "" {
			execPath = strings.TrimPrefix(line, "Exec=")
			// Loại bỏ các tham số %u, %U, %f, %F, etc.
			execPath = cleanExecPath(execPath)
		}
		if strings.HasPrefix(line, "NoDisplay=true") {
			return "", "" // Bỏ qua app ẩn
		}
	}

	return name, execPath
}

// cleanExecPath loại bỏ các placeholder và tham số từ trường Exec
func cleanExecPath(exec string) string {
	parts := strings.Fields(exec)
	if len(parts) == 0 {
		return ""
	}

	// Lấy phần đầu tiên (path tới executable)
	execPath := parts[0]

	// Loại bỏ env prefix nếu có (ví dụ: "env VAR=val /usr/bin/app")
	if execPath == "env" {
		for i := 1; i < len(parts); i++ {
			if !strings.Contains(parts[i], "=") {
				return parts[i]
			}
		}
	}

	return execPath
}
