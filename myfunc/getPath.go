package myfunc

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shirou/gopsutil/v3/process"
)

// Kết nối DB (tạo file resources.db nếu chưa có)
func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "data/resources.db")
	if err != nil {
		return nil, err
	}

	// Thư mục "data" phải tồn tại; nếu chưa có thì tạo (tùy bạn làm ở ngoài)
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Tạo bảng nếu chưa có
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS resources (
            id      INTEGER PRIMARY KEY AUTOINCREMENT,
            name    TEXT NOT NULL,
            path    TEXT NOT NULL,
            is_web  BOOLEAN NOT NULL
        );

        CREATE TABLE IF NOT EXISTS groups (
            id      INTEGER PRIMARY KEY AUTOINCREMENT,
            nameG   TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS group_resources (
            group_id    INTEGER NOT NULL,
            resource_id INTEGER NOT NULL,
            PRIMARY KEY (group_id, resource_id),
            FOREIGN KEY (group_id) REFERENCES groups(id),
            FOREIGN KEY (resource_id) REFERENCES resources(id)
        );
    `)
	if err != nil {
		return nil, err
	}

	// Đảm bảo name là duy nhất để UPSERT hoạt động
	_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_resources_name ON resources(name);`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Lưu resource vào DB
func SaveResourceToDB(db *sql.DB, name, path string, isWeb bool) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	_, err := db.Exec(`
        INSERT INTO resources (name, path, is_web)
        VALUES (?, ?, ?)
        ON CONFLICT(name) DO UPDATE SET
            path = excluded.path;
    `, name, path, isWeb)
	return err
}

// Lấy tất cả resources (apps + webs)
func GetAllResourcesByName(db *sql.DB) error {
	allResources := make(map[string]string)

	// 1. Lấy danh sách ứng dụng đã cài đặt (platform-specific)
	installedResources, err := GetInstalledResourcesByName()
	if err != nil {
		return err
	}
	for name, path := range installedResources {
		allResources[name] = path
		if err := SaveResourceToDB(db, name, path, false); err != nil {
			log.Printf("Save installed resource failed (%s): %v", name, err)
		}
	}

	// 2. Running processes
	runningResources, err := GetRunningProcessesByName()
	if err != nil {
		return err
	}
	for name, path := range runningResources {
		if _, exists := allResources[name]; !exists {
			allResources[name] = path
		}
		if err := SaveResourceToDB(db, name, path, false); err != nil {
			log.Printf("Save running resource failed (%s): %v", name, err)
		}
	}

	return nil
}

// Resource định nghĩa cấu trúc dữ liệu của một ứng dụng/web
type Resource struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsWeb bool   `json:"is_web"`
}

// Lấy resources từ DB với đầy đủ ID, Name, Path
func GetResourcesFromDB(db *sql.DB) ([]Resource, error) {
	var resources []Resource

	rows, err := db.Query("SELECT id, name, path, is_web FROM resources")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r Resource
		if err := rows.Scan(&r.ID, &r.Name, &r.Path, &r.IsWeb); err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resources, nil
}

// in ra dữ liệu lấy từ GetResourcesFromDB
func ShowDB(db *sql.DB) {
	res, err := GetResourcesFromDB(db)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range res {
		fmt.Printf("ID: %d | Name: %s | Path: %s\n", r.ID, r.Name, r.Path)
	}
}

// Lấy process đang chạy (cross-platform nhờ gopsutil)
func GetRunningProcessesByName() (map[string]string, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	processMap := make(map[string]string)
	for _, p := range procs {
		path, err := p.Exe()
		if err != nil || path == "" {
			continue
		}

		path = filepath.Clean(strings.ToLower(path))
		if name, err := p.Name(); err == nil {
			processMap[strings.ToLower(name)] = path
		}
	}
	return processMap, nil
}
