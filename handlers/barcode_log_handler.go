package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RecordBarcodeLog records an entry into BARCODE_LOGS and AUDIT_LOGS.
func RecordBarcodeLog(db *sql.DB, itemNo, itemName, newBarcode, prevBarcode, userID, username string) error {
	trimmedBarcode := strings.TrimSpace(newBarcode)
	if trimmedBarcode == "" {
		return nil
	}

	if username == "" {
		username = "Admin"
	}
	if userID == "" {
		userID = "admin"
	}

	logID := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	// Insert into BARCODE_LOGS
	query := `INSERT INTO BARCODE_LOGS (ID, ITEM_NO, ITEM_NAME, BARCODE, PREVIOUS_BARCODE, USER_ID, USERNAME, RECORDED_AT)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, logID, itemNo, itemName, trimmedBarcode, prevBarcode, userID, username, now)
	if err != nil {
		log.Printf("[Error] RecordBarcodeLog failed: %v", err)
		return err
	}

	// Insert into AUDIT_LOGS
	auditID := uuid.New().String()
	auditDetails := fmt.Sprintf("Merekam barcode '%s' untuk produk '%s' (%s)", trimmedBarcode, itemName, itemNo)
	if prevBarcode != "" {
		auditDetails += fmt.Sprintf(" [sebelumnya: %s]", prevBarcode)
	}
	_, _ = db.Exec("INSERT INTO AUDIT_LOGS (ID, TIMESTAMP, USER_ID, USERNAME, ACTION, DETAILS) VALUES (?, ?, ?, ?, ?, ?)",
		auditID, now, userID, username, "record_barcode", auditDetails)

	return nil
}

// GetBarcodeLogs retrieves barcode recording history and analytics.
func GetBarcodeLogs(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := GetPaginationParams(c)
		startDate := strings.TrimSpace(c.Query("start_date"))
		endDate := strings.TrimSpace(c.Query("end_date"))
		userFilter := strings.TrimSpace(c.Query("user"))
		if userFilter == "" {
			userFilter = strings.TrimSpace(c.Query("username"))
		}
		searchQuery := strings.TrimSpace(c.Query("q"))
		if searchQuery == "" {
			searchQuery = strings.TrimSpace(c.Query("search"))
		}

		var conditions []string
		var args []interface{}

		if startDate != "" {
			startDay := startDate
			if len(startDay) >= 10 {
				startDay = startDay[:10]
			}
			conditions = append(conditions, "substr(RECORDED_AT, 1, 10) >= ?")
			args = append(args, startDay)
		}
		if endDate != "" {
			endDay := endDate
			if len(endDay) >= 10 {
				endDay = endDay[:10]
			}
			conditions = append(conditions, "substr(RECORDED_AT, 1, 10) <= ?")
			args = append(args, endDay)
		}
		if userFilter != "" && userFilter != "all" {
			conditions = append(conditions, "LOWER(USERNAME) = LOWER(?)")
			args = append(args, userFilter)
		}
		if searchQuery != "" {
			wildcard := "%" + searchQuery + "%"
			conditions = append(conditions, "(ITEM_NO LIKE ? OR ITEM_NAME LIKE ? OR BARCODE LIKE ? OR USERNAME LIKE ?)")
			args = append(args, wildcard, wildcard, wildcard, wildcard)
		}

		whereClause := ""
		if len(conditions) > 0 {
			whereClause = "WHERE " + strings.Join(conditions, " AND ") + " "
		}

		// 1. Fetch Logs
		query := fmt.Sprintf(`SELECT ID, ITEM_NO, ITEM_NAME, BARCODE, COALESCE(PREVIOUS_BARCODE, ''), COALESCE(USER_ID, ''), USERNAME, RECORDED_AT
			FROM BARCODE_LOGS %sORDER BY RECORDED_AT DESC LIMIT ? OFFSET ?`, whereClause)
		queryArgs := append(args, limit, offset)

		rows, err := db.Query(query, queryArgs...)
		if err != nil {
			log.Printf("[Error] Query GetBarcodeLogs failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengambil riwayat rekam barcode")
			return
		}
		defer rows.Close()

		var logs []models.BarcodeLog
		for rows.Next() {
			var l models.BarcodeLog
			if err := rows.Scan(&l.ID, &l.ItemNo, &l.ItemName, &l.Barcode, &l.PreviousBarcode, &l.UserID, &l.Username, &l.RecordedAt); err != nil {
				log.Printf("[Error] Scan GetBarcodeLogs failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal membaca riwayat rekam barcode")
				return
			}
			logs = append(logs, l)
		}
		if logs == nil {
			logs = []models.BarcodeLog{}
		}

		// 2. Aggregate Staff Performance in current filter
		// Use same where conditions except pagination
		staffStatsQuery := fmt.Sprintf(`SELECT USERNAME, COUNT(*) as cnt
			FROM BARCODE_LOGS %sGROUP BY USERNAME ORDER BY cnt DESC`, whereClause)
		statRows, err := db.Query(staffStatsQuery, args...)
		var staffStats []models.StaffBarcodeStat
		totalPeriod := 0
		topRecorder := "-"

		if err == nil {
			defer statRows.Close()
			for statRows.Next() {
				var s models.StaffBarcodeStat
				if err := statRows.Scan(&s.Username, &s.Count); err == nil {
					totalPeriod += s.Count
					staffStats = append(staffStats, s)
				}
			}
		}

		if len(staffStats) > 0 {
			topRecorder = fmt.Sprintf("%s (%d item)", staffStats[0].Username, staffStats[0].Count)
			for i := range staffStats {
				if totalPeriod > 0 {
					staffStats[i].Percentage = (float64(staffStats[i].Count) / float64(totalPeriod)) * 100.0
				}
			}
		} else {
			staffStats = []models.StaffBarcodeStat{}
		}

		// 3. Store-wide barcode coverage stats
		var totalStoreItems, storeWithBarcode int
		_ = db.QueryRow(`SELECT count(*), 
			count(CASE WHEN ITEMUPC IS NOT NULL AND trim(ITEMUPC) != '' THEN 1 END) 
			FROM ITEM`).Scan(&totalStoreItems, &storeWithBarcode)

		storeProgressPct := 0.0
		if totalStoreItems > 0 {
			storeProgressPct = (float64(storeWithBarcode) / float64(totalStoreItems)) * 100.0
		}

		resp := models.BarcodeLogResponse{
			Logs:             logs,
			TotalPeriod:      totalPeriod,
			TopRecorder:      topRecorder,
			StaffStats:       staffStats,
			TotalStoreItems:  totalStoreItems,
			StoreWithBarcode: storeWithBarcode,
			StoreProgressPct: storeProgressPct,
		}

		SendSuccess(c, "Data log barcode berhasil diambil", resp)
	}
}
