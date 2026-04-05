package myfunc

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func CreateGroup(db *sql.DB, name string, listIDs []int) error {
    if db == nil {
        return fmt.Errorf("db is nil")
    }

    // 1. Tạo group mới
    res, err := db.Exec(`INSERT INTO groups (nameG) VALUES (?)`, name)
    if err != nil {
        return fmt.Errorf("failed to insert group: %w", err)
    }

    // 2. Lấy ID của group vừa tạo
    groupID, err := res.LastInsertId()
    if err != nil {
        return fmt.Errorf("failed to get group ID: %w", err)
    }

    // 3. Gán list resource ID vào group_resources
    for _, resID := range listIDs {
        _, err := db.Exec(
            `INSERT INTO group_resources (group_id, resource_id) VALUES (?, ?)`,
            groupID, resID,
        )
        if err != nil {
            return fmt.Errorf("failed to insert resource %d into group: %w", resID, err)
        }
    }

    return nil
}

func RunGroup(db *sql.DB, groupName string) error {
    if db == nil {
        return fmt.Errorf("db is nil")
    }

    // 1. Lấy danh sách resource_id theo tên group
    rows, err := db.Query(`
        SELECT r.id, r.is_web
        FROM resources r
        JOIN group_resources gr ON r.id = gr.resource_id
        JOIN groups g ON g.id = gr.group_id
        WHERE g.nameG = ?`, groupName)
    if err != nil {
        return fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()

    // 2. Duyệt qua tất cả resource trong group
    for rows.Next() {
        var id int
        var isWeb bool

        if err := rows.Scan(&id, &isWeb); err != nil {
            fmt.Printf("⚠️ Lỗi scan row: %v\n", err)
            continue
        }

        if isWeb {
            if err := OpenURL(db, id); err != nil {
                fmt.Printf("⚠️ Lỗi mở web id=%d: %v\n", id, err)
            }
        } else {
            if err := OpenApps(db, id); err != nil {
                fmt.Printf("⚠️ Lỗi mở app id=%d: %v\n", id, err)
            }
        }
    }

    return nil
}

//các hàm update
func UpdateGroup(db *sql.DB, groupID int, listIDs []int) error {
    if db == nil {
        return fmt.Errorf("db is nil")
    }

    tx, err := db.Begin()
    if err != nil {
        return err
    }

    // 1. Xóa các resource cũ trong group
    _, err = tx.Exec(`DELETE FROM group_resources WHERE group_id = ?`, groupID)
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to clear old resources: %w", err)
    }

    // 2. Thêm lại các resource mới
    for _, resID := range listIDs {
        _, err := tx.Exec(
            `INSERT INTO group_resources (group_id, resource_id) VALUES (?, ?)`,
            groupID, resID,
        )
        if err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to insert resource %d: %w", resID, err)
        }
    }

    // 3. Commit transaction
    return tx.Commit()
}

func UpdateGroupByName(db *sql.DB, groupName string, listIDs []int) error {
    if db == nil {
        return fmt.Errorf("db is nil")
    }

    // 1. Lấy groupID từ groupName
    var groupID int
    err := db.QueryRow(`SELECT id FROM groups WHERE nameG = ?`, groupName).Scan(&groupID)
    if err != nil {
        return fmt.Errorf("group %s not found: %w", groupName, err)
    }

    // 2. Gọi UpdateGroup để update theo groupID
    return UpdateGroup(db, groupID, listIDs)
}

// Xóa group theo ID
func DeleteGroup(db *sql.DB, groupID int) error {
    if db == nil {
        return fmt.Errorf("db is nil")
    }

    tx, err := db.Begin()
    if err != nil {
        return err
    }

    // 1. Xóa các resource liên kết
    _, err = tx.Exec(`DELETE FROM group_resources WHERE group_id = ?`, groupID)
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to delete group_resources for group %d: %w", groupID, err)
    }

    // 2. Xóa group
    _, err = tx.Exec(`DELETE FROM groups WHERE id = ?`, groupID)
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to delete group %d: %w", groupID, err)
    }

    return tx.Commit()
}

// Xóa group theo tên
func DeleteGroupByName(db *sql.DB, groupName string) error {
    if db == nil {
        return fmt.Errorf("db is nil")
    }

    var groupID int
    err := db.QueryRow(`SELECT id FROM groups WHERE nameG = ?`, groupName).Scan(&groupID)
    if err != nil {
        return fmt.Errorf("group %s not found: %w", groupName, err)
    }

    return DeleteGroup(db, groupID)
}

func ShowGroups(db *sql.DB) error {
    if db == nil {
        return fmt.Errorf("db is nil")
    }

    rows, err := db.Query(`SELECT id, nameG FROM groups`)
    if err != nil {
        return fmt.Errorf("failed to query groups: %w", err)
    }
    defer rows.Close()

    fmt.Println("📂 Danh sách group:")
    found := false
    for rows.Next() {
        var id int
        var name string
        if err := rows.Scan(&id, &name); err != nil {
            return fmt.Errorf("failed to scan group: %w", err)
        }
        fmt.Printf(" - ID: %d | Name: %s\n", id, name)
        found = true
    }

    if !found {
        fmt.Println("⚠️  Không có group nào trong DB.")
    }

    return nil
}
