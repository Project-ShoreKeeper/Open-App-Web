//go:build darwin

package myfunc

import (
	"os"
	"path/filepath"
	"strings"
)

// GetInstalledResourcesByName quét thư mục /Applications và ~/Applications
// để tìm các ứng dụng .app trên macOS
func GetInstalledResourcesByName() (map[string]string, error) {
	resources := make(map[string]string)

	// Các thư mục chứa ứng dụng trên macOS
	appDirs := []string{
		"/Applications",
	}

	// Thêm ~/Applications nếu có
	homeDir, err := os.UserHomeDir()
	if err == nil {
		appDirs = append(appDirs, filepath.Join(homeDir, "Applications"))
	}

	for _, dir := range appDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".app") {
				continue
			}

			appName := strings.TrimSuffix(entry.Name(), ".app")
			appPath := filepath.Join(dir, entry.Name())

			// Tìm executable bên trong .app bundle
			macOSDir := filepath.Join(appPath, "Contents", "MacOS")
			exeEntries, err := os.ReadDir(macOSDir)
			if err != nil || len(exeEntries) == 0 {
				// Nếu không tìm thấy executable, dùng đường dẫn .app
				resources[strings.ToLower(appName)] = appPath
				continue
			}

			// Lấy executable đầu tiên (thường trùng tên app)
			exePath := filepath.Join(macOSDir, exeEntries[0].Name())
			// Ưu tiên executable cùng tên với app
			for _, exe := range exeEntries {
				if strings.EqualFold(exe.Name(), appName) {
					exePath = filepath.Join(macOSDir, exe.Name())
					break
				}
			}

			resources[strings.ToLower(appName)] = exePath
		}
	}

	// Thêm các ứng dụng từ Homebrew Cask (nếu có)
	brewCaskDir := "/opt/homebrew/Caskroom"
	if _, err := os.Stat(brewCaskDir); err == nil {
		entries, err := os.ReadDir(brewCaskDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				name := strings.ToLower(entry.Name())
				if _, exists := resources[name]; !exists {
					// Tìm .app trong Caskroom
					versionDirs, _ := os.ReadDir(filepath.Join(brewCaskDir, entry.Name()))
					for _, vd := range versionDirs {
						appGlob := filepath.Join(brewCaskDir, entry.Name(), vd.Name(), "*.app")
						matches, _ := filepath.Glob(appGlob)
						if len(matches) > 0 {
							resources[name] = matches[0]
							break
						}
					}
				}
			}
		}
	}

	return resources, nil
}
