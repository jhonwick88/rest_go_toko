package routes

import (
	"database/sql"

	"rest_go_toko/handlers"
	"rest_go_toko/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the Gin engine, attaches middleware, and registers the endpoints.
func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Activation Route (Unprotected)
	r.POST("/api/license/activate", handlers.ActivateLicense)

	// API Group (Protected by License Middleware)
	api := r.Group("/api")
	api.Use(middleware.RequireLicense())
	{
		// Category routes
		api.GET("/categories", handlers.GetCategories(db))
		api.GET("/categories/:id/items", handlers.GetCategoryItems(db))

		// Item routes
		api.GET("/items", handlers.GetItems(db))
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
