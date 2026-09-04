package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"rest_go_toko/models"
	"github.com/gin-gonic/gin"
)

// GetUnits returns a list of all item units
func GetUnits(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT ID, NAME FROM ITEM_UNIT ORDER BY ID ASC")
		if err != nil {
			log.Printf("[Error] Query GetUnits failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units", "details": err.Error()})
			return
		}
		defer rows.Close()

		var units []models.ItemUnit
		for rows.Next() {
			var unit models.ItemUnit
			if err := rows.Scan(&unit.ID, &unit.Name); err != nil {
				log.Printf("[Error] Scan GetUnits failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse units data"})
				return
			}
			units = append(units, unit)
		}

		if units == nil {
			units = []models.ItemUnit{}
		}

		c.JSON(http.StatusOK, gin.H{"data": units})
	}
}

// CreateUnit creates a new item unit
func CreateUnit(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name cannot be empty"})
			return
		}

		result, err := db.Exec("INSERT INTO ITEM_UNIT (NAME) VALUES (?)", req.Name)
		if err != nil {
			log.Printf("[Error] Query CreateUnit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create unit", "details": err.Error()})
			return
		}

		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"message": "Unit created successfully", "data": models.ItemUnit{ID: int(id), Name: req.Name}})
	}
}

// UpdateUnit updates an existing unit
func UpdateUnit(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
			return
		}

		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name cannot be empty"})
			return
		}

		res, err := db.Exec("UPDATE ITEM_UNIT SET NAME = ? WHERE ID = ?", req.Name, id)
		if err != nil {
			log.Printf("[Error] Query UpdateUnit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update unit", "details": err.Error()})
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Unit updated successfully", "data": models.ItemUnit{ID: id, Name: req.Name}})
	}
}

// DeleteUnit deletes a unit
func DeleteUnit(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
			return
		}

		res, err := db.Exec("DELETE FROM ITEM_UNIT WHERE ID = ?", id)
		if err != nil {
			log.Printf("[Error] Query DeleteUnit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete unit", "details": err.Error()})
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Unit deleted successfully"})
	}
}
