package routes

import (
	"database/sql"

	"rest_go_toko/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the Gin engine, attaches middleware, and registers the endpoints.
func SetupRouter(db *sql.DB) *gin.Engine {
	// Set Gin mode (use release mode or debug based on needs; we can keep standard defaults)
	// For production readiness, you can use gin.New() and customize middleware.
	r := gin.New()

	// Logger middleware writes logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	r.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// API Group
	api := r.Group("/api")
	{
		// Category routes
		api.GET("/categories", handlers.GetCategories(db))
		api.GET("/categories/:id/items", handlers.GetCategoryItems(db))

		// Item routes
		api.GET("/items", handlers.GetItems(db))
		// Search route must be defined before the wildcard route to avoid matching conflicts in standard routers,
		// though Gin handles it natively, this is clean development practice.
		api.GET("/items/search", handlers.SearchItems(db))
		api.GET("/items/:itemno", handlers.GetItemByNo(db))
		api.PUT("/items/:itemno", handlers.UpdateItem(db))
		api.POST("/items", handlers.CreateItem(db))

		// User routes
		api.GET("/users", handlers.GetUsers(db))
		api.POST("/users", handlers.CreateUser(db))

		// Quick Item routes
		api.GET("/quick-items", handlers.GetQuickItems(db))
		api.POST("/quick-items", handlers.CreateQuickItem(db))

		// Sales routes
		api.GET("/sales", handlers.GetSales(db))
		api.POST("/sales", handlers.CreateSale(db))
		api.PATCH("/sales/:invoiceno/status", handlers.UpdateSaleStatus(db))

		// Cash Reconciliation routes
		api.GET("/cash-reconciliations", handlers.GetCashReconciliations(db))
		api.POST("/cash-reconciliations", handlers.CreateCashReconciliation(db))

		// Audit Log routes
		api.GET("/audit-logs", handlers.GetAuditLogs(db))
		api.POST("/audit-logs", handlers.CreateAuditLog(db))

		// Company routes
		api.GET("/company", handlers.GetCompany(db))
		api.PUT("/company", handlers.UpdateCompany(db))
	}

	return r
}
