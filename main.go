package main

import (
	"log"
	"myapp/myfunc"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := myfunc.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initial load of resources
	err = myfunc.GetAllResourcesByName(db)
	if err != nil {
		log.Printf("Warning: Failed to load some resources: %v", err)
	}

	r := gin.Default()

	api := r.Group("/api")
	{
		// 1. Tải lại dữ liệu (reload)
		api.POST("/reload", func(c *gin.Context) {
			err := myfunc.GetAllResourcesByName(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Resources reloaded successfully"})
		})

		// 2. Lấy toàn bộ resource (showdb)
		api.GET("/resources", func(c *gin.Context) {
			res, err := myfunc.GetResourcesFromDB(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": res})
		})

		// 3. Xóa resource (delres)
		api.DELETE("/resources", func(c *gin.Context) {
			var req struct {
				ID   *int    `json:"id"`
				Name *string `json:"name"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
				return
			}

			if req.ID != nil {
				err = myfunc.DeleteResourceByID(db, *req.ID)
			} else if req.Name != nil {
				err = myfunc.DeleteResourceByName(db, *req.Name)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Must provide 'id' or 'name'"})
				return
			}

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Resource deleted successfully"})
		})
        
		// 4. Mở App theo ID
		api.POST("/resources/run", func(c *gin.Context) {
			var req struct {
				ID *int `json:"id"`
			}
			if err := c.ShouldBindJSON(&req); err != nil || req.ID == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Must provide 'id' as an integer"})
				return
			}

			var isWeb bool
			err := db.QueryRow("SELECT is_web FROM resources WHERE id = ?", *req.ID).Scan(&isWeb)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
				return
			}

			if isWeb {
				err = myfunc.OpenURL(db, *req.ID) 
			} else {
				err = myfunc.OpenApps(db, *req.ID)
			}

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Launched successfully"})
		})

		// 5. Thêm Web (addweb)
		api.POST("/webs", func(c *gin.Context) {
			var req struct {
				Name string `json:"name" binding:"required"`
				URL  string `json:"url" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err := myfunc.SaveWebToDB(db, req.Name, req.URL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Web resource added successfully"})
		})

		// 6. Tạo Group (makegroup)
		api.POST("/groups", func(c *gin.Context) {
			var req struct {
				Name string `json:"name" binding:"required"`
				IDs  []int  `json:"ids" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err = myfunc.CreateGroup(db, req.Name, req.IDs)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Group created successfully"})
		})

		// 7. Chạy Group (rungroup)
		api.POST("/groups/run", func(c *gin.Context) {
			var req struct {
				Name string `json:"name" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err = myfunc.RunGroup(db, req.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Group launched successfully"})
		})

		// 8. Xóa Group (del)
		api.DELETE("/groups", func(c *gin.Context) {
			var req struct {
				Name string `json:"name" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err = myfunc.DeleteGroupByName(db, req.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
		})
	}

	port := "8080"
	log.Printf("Host OS Application Launcher Service running on http://127.0.0.1:%s", port)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
