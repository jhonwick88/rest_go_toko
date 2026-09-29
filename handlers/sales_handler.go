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
	"rest_go_toko/services"
)

// GetSales returns all sales headers
func GetSales(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := GetPaginationParams(c)
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")
		searchQuery := strings.TrimSpace(c.Query("q"))
		if searchQuery == "" {
			searchQuery = strings.TrimSpace(c.Query("search"))
		}

		var query string
		var args []interface{}

		var conditions []string
		if startDate != "" && endDate != "" {
			conditions = append(conditions, "DATE >= ? AND DATE <= ?")
			args = append(args, startDate, endDate)
		}
		if searchQuery != "" {
			conditions = append(conditions, "(INVOICE_NO LIKE ? OR CASHIER LIKE ? OR INVOICE_NO IN (SELECT DISTINCT INVOICE_NO FROM SALES_ITEMS WHERE ITEMNAME LIKE ? OR ITEMNO LIKE ? OR ITEMUPC LIKE ?))")
			wildcard := "%" + searchQuery + "%"
			args = append(args, wildcard, wildcard, wildcard, wildcard, wildcard)
		}

		whereClause := ""
		if len(conditions) > 0 {
			whereClause = "WHERE " + strings.Join(conditions, " AND ") + " "
		}

		query = fmt.Sprintf("SELECT ID, INVOICE_NO, DATE, CASHIER, SUBTOTAL, DISCOUNT, GRAND_TOTAL, TAX, PAYMENT_METHOD, PAID_AMOUNT, CHANGE_AMOUNT, STATUS, VOID_REASON, VOIDED_AT, VOIDED_BY, REFUND_REASON, REFUNDED_AT, REFUNDED_BY, CUSTOMER_ID, IS_VOID FROM SALES %sORDER BY DATE DESC LIMIT ? OFFSET ?", whereClause)
		args = append(args, limit, offset)
		
		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("[Error] Query GetSales failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengambil data penjualan")
			return
		}
		defer rows.Close()

		var sales []models.Sales
		var invoiceNos []string
		for rows.Next() {
			var s models.Sales
			var customerID sql.NullInt64
			var cashierID, paymentMethod, saleDate, status, voidReason, voidedAt, voidedBy, refundReason, refundedAt, refundedBy sql.NullString

			if err := rows.Scan(&s.ID, &s.ReceiptNo, &saleDate, &cashierID, &s.Subtotal, &s.Discount, &s.Total, &s.Tax, &paymentMethod, &s.CashTendered, &s.ChangeDue, &status, &voidReason, &voidedAt, &voidedBy, &refundReason, &refundedAt, &refundedBy, &customerID, &s.IsVoid); err != nil {
				log.Printf("[Error] Scan GetSales failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal membaca data penjualan")
				return
			}
			
			s.SaleDate = saleDate.String
			s.CashierID = cashierID.String
			s.PaymentMethod = paymentMethod.String
			s.Status = status.String
			s.VoidReason = voidReason.String
			s.VoidedAt = voidedAt.String
			s.VoidedBy = voidedBy.String
			s.RefundReason = refundReason.String
			s.RefundedAt = refundedAt.String
			s.RefundedBy = refundedBy.String
			s.Items = []models.SalesItem{}
			
			if customerID.Valid {
				cid := int(customerID.Int64)
				s.CustomerID = &cid
			}

			invoiceNos = append(invoiceNos, s.ReceiptNo)
			sales = append(sales, s)
		}

		// Batch fetch items for all invoices in a single query (resolving N+1 query problem)
		if len(invoiceNos) > 0 {
			placeholders := make([]string, len(invoiceNos))
			itemArgs := make([]interface{}, len(invoiceNos))
			for i, inv := range invoiceNos {
				placeholders[i] = "?"
				itemArgs[i] = inv
			}

			itemsQuery := fmt.Sprintf(`SELECT 
				si.ID, si.INVOICE_NO, si.ITEMNO, si.ITEMNAME, si.ITEMUPC, si.QTY, si.PRICE, si.DISCOUNT, si.SUBTOTAL, si.NOTE,
				COALESCE(si.CATEGORY_ID, i.CATEGORY_ID, 0) AS CATEGORY_ID,
				COALESCE(NULLIF(si.CATEGORY_NAME, ''), c.NAME, 'Tanpa Kategori') AS CATEGORY_NAME
			FROM SALES_ITEMS si
			LEFT JOIN ITEM i ON (LOWER(si.ITEMNO) = LOWER(i.ITEMNO) OR (si.ITEMUPC != '' AND LOWER(si.ITEMUPC) = LOWER(i.ITEMUPC)))
			LEFT JOIN CATEGORY c ON (COALESCE(si.CATEGORY_ID, i.CATEGORY_ID, 0) = c.ID)
			WHERE si.INVOICE_NO IN (%s)
			ORDER BY si.ID ASC`, strings.Join(placeholders, ","))

			itemRows, err := db.Query(itemsQuery, itemArgs...)
			if err != nil {
				log.Printf("[Error] Query batch GetSales items failed: %v", err)
			} else {
				defer itemRows.Close()
				itemsByInvoice := make(map[string][]models.SalesItem)
				for itemRows.Next() {
					var it models.SalesItem
					var id sql.NullInt64
					var saleId, itemNo, itemName, itemUpc, note, catName sql.NullString
					var qty, price, discount, subtotal sql.NullFloat64
					var catId sql.NullInt64

					if err := itemRows.Scan(&id, &saleId, &itemNo, &itemName, &itemUpc, &qty, &price, &discount, &subtotal, &note, &catId, &catName); err == nil {
						it.ID = int(id.Int64)
						it.SaleID = saleId.String
						it.ItemNo = itemNo.String
						it.ItemName = itemName.String
						it.ItemUPC = itemUpc.String
						it.Qty = qty.Float64
						it.Price = price.Float64
						it.Discount = discount.Float64
						it.Total = subtotal.Float64
						it.Note = note.String
						it.CategoryID = int(catId.Int64)
						it.CategoryName = catName.String
						itemsByInvoice[it.SaleID] = append(itemsByInvoice[it.SaleID], it)
					}
				}

				for i := range sales {
					if its, ok := itemsByInvoice[sales[i].ReceiptNo]; ok {
						sales[i].Items = its
					}
				}
			}
		}

		if sales == nil {
			sales = []models.Sales{}
		}
		SendSuccess(c, "Data ditemukan", sales)
	}
}

// CreateSale creates a new sale with its items in a transaction
func CreateSale(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateSalesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid: " + err.Error())
			return
		}

		if len(req.Items) == 0 {
			SendError(c, http.StatusBadRequest, "Keranjang belanja tidak boleh kosong")
			return
		}

		// Generate IDs
		saleID := uuid.New().String()
		if req.ReceiptNo == "" {
			req.ReceiptNo = "INV-" + time.Now().Format("20060102-150405")
		}

		// Start a database transaction
		tx, err := db.Begin()
		if err != nil {
			log.Printf("[Error] Failed to start transaction: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal memulai transaksi database")
			return
		}

		// Insert Header
		headerQuery := "INSERT INTO SALES (ID, INVOICE_NO, DATE, CASHIER, SUBTOTAL, DISCOUNT, GRAND_TOTAL, TAX, PAYMENT_METHOD, PAID_AMOUNT, CHANGE_AMOUNT, STATUS, VOID_REASON, VOIDED_AT, VOIDED_BY, REFUND_REASON, REFUNDED_AT, REFUNDED_BY, CUSTOMER_ID, IS_VOID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		var custID interface{} = nil
		if req.CustomerID != nil {
			custID = *req.CustomerID
		}

		if req.Status == "" {
			req.Status = "completed"
		}

		nowDate := time.Now().Format(time.RFC3339)
		if req.SaleDate == "" {
			req.SaleDate = nowDate
		}

		_, err = tx.Exec(headerQuery, saleID, req.ReceiptNo, req.SaleDate, req.CashierID, req.Subtotal, req.Discount, req.Total, req.Tax, req.PaymentMethod, req.CashTendered, req.ChangeDue, req.Status, req.VoidReason, req.VoidedAt, req.VoidedBy, req.RefundReason, req.RefundedAt, req.RefundedBy, custID, req.IsVoid)
		if err != nil {
			tx.Rollback()
			log.Printf("[Error] Insert Sales Header failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menyimpan data transaksi (Header)")
			return
		}

		// Check specific Pro Features
		hasStockMovement := false
		hasAdvancedPos := false
		if claims, err := services.GetLicenseClaims(); err == nil && claims != nil {
			if smFlag, ok := claims.Features["stock_movement"].(bool); ok {
				hasStockMovement = smFlag
			}
			if apFlag, ok := claims.Features["advanced_pos"].(bool); ok {
				hasAdvancedPos = apFlag
			}
		}

		// Insert Items
		var maxID int
		_ = tx.QueryRow("SELECT COALESCE(MAX(ID), 0) FROM SALES_ITEMS").Scan(&maxID)

		itemQuery := "INSERT INTO SALES_ITEMS (ID, INVOICE_NO, ITEMNO, ITEMNAME, ITEMUPC, QTY, PRICE, DISCOUNT, SUBTOTAL, NOTE, TAX, CATEGORY_ID, CATEGORY_NAME) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		updateStockQuery := "UPDATE ITEM SET OBQUANTITY = OBQUANTITY - ? WHERE ITEMNO = ?"
		stockLedgerQuery := "INSERT INTO STOCK_LEDGER (ITEMNO, TX_TYPE, QTY, NOTES, REFERENCE_ID, TIMESTAMP) VALUES (?, 'OUT', ?, 'Sale', ?, ?)"
		
		for i, item := range req.Items {
			maxID++
			
			noteToSave := item.Note
			if item.UnitName != "" {
				if noteToSave != "" {
					noteToSave += " | "
				}
				noteToSave += "Satuan: " + item.UnitName
			}
			
			// Only allow saving Tax if advanced_pos is enabled
			itemTax := 0.0
			if hasAdvancedPos {
				itemTax = item.Tax
			}

			catID := item.CategoryID
			catName := item.CategoryName

			// Fallback resolution from master ITEM and CATEGORY if not provided
			if catID == 0 || catName == "" {
				var dbCatID sql.NullInt64
				var dbCatName sql.NullString
				_ = tx.QueryRow(`
					SELECT COALESCE(i.CATEGORY_ID, 0), COALESCE(c.NAME, 'Tanpa Kategori')
					FROM ITEM i
					LEFT JOIN CATEGORY c ON i.CATEGORY_ID = c.ID
					WHERE LOWER(i.ITEMNO) = LOWER(?) OR (i.ITEMUPC != '' AND LOWER(i.ITEMUPC) = LOWER(?))
					LIMIT 1
				`, item.ItemNo, item.ItemUPC).Scan(&dbCatID, &dbCatName)

				if catID == 0 && dbCatID.Valid {
					catID = int(dbCatID.Int64)
				}
				if catName == "" && dbCatName.Valid {
					catName = dbCatName.String
				}
			}
			if catName == "" {
				catName = "Tanpa Kategori"
			}

			_, err = tx.Exec(itemQuery, maxID, req.ReceiptNo, item.ItemNo, item.ItemName, item.ItemUPC, item.Qty, item.Price, item.Discount, item.Total, noteToSave, itemTax, catID, catName)
			if err != nil {
				tx.Rollback()
				log.Printf("[Error] Insert Sales Item failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal menyimpan data barang transaksi")
				return
			}

			// Deduct stock based on ratio
			qtyToDeduct := item.Qty
			if item.Ratio > 0 {
				qtyToDeduct = item.Qty * item.Ratio
			}
			
			_, err = tx.Exec(updateStockQuery, qtyToDeduct, item.ItemNo)
			if err != nil {
				tx.Rollback()
				log.Printf("[Error] Update Stock failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal mengurangi stok produk")
				return
			}

			if hasStockMovement {
				_, err = tx.Exec(stockLedgerQuery, item.ItemNo, -qtyToDeduct, req.ReceiptNo, nowDate)
				if err != nil {
					tx.Rollback()
					log.Printf("[Error] Insert Stock Ledger failed: %v", err)
					SendError(c, http.StatusInternalServerError, "Gagal mencatat mutasi stok")
					return
				}
			}

			req.Items[i].ID = maxID
			req.Items[i].SaleID = req.ReceiptNo
		}

		// Insert Payments if advanced_pos is enabled and payments exist
		if hasAdvancedPos && len(req.Payments) > 0 {
			paymentQuery := "INSERT INTO SALES_PAYMENTS (INVOICENO, PAYMENT_METHOD, AMOUNT, PAYMENT_DATE) VALUES (?, ?, ?, ?)"
			for _, payment := range req.Payments {
				_, err = tx.Exec(paymentQuery, req.ReceiptNo, payment.PaymentMethod, payment.Amount, nowDate)
				if err != nil {
					tx.Rollback()
					log.Printf("[Error] Insert Sales Payment failed: %v", err)
					SendError(c, http.StatusInternalServerError, "Gagal menyimpan detail pembayaran")
					return
				}
			}
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			log.Printf("[Error] Transaction commit failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal memfinalisasi transaksi")
			return
		}

		req.ID = saleID
		SendSuccess(c, "Transaksi berhasil disimpan", req)
	}
}

// UpdateSaleStatus updates a sale's status (voided, refunded).
func UpdateSaleStatus(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		invoiceNo := c.Param("invoiceno")
		if invoiceNo == "" {
			SendError(c, http.StatusBadRequest, "InvoiceNo wajib diisi")
			return
		}

		type UpdateStatusInput struct {
			Status string `json:"status" binding:"required"`
			Reason string `json:"reason"`
			User   string `json:"user"`
		}

		var input UpdateStatusInput
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid")
			return
		}

		now := time.Now().Format(time.RFC3339)
		
		var query string
		var args []interface{}
		
		if input.Status == "voided" {
			query = "UPDATE SALES SET STATUS = ?, IS_VOID = 1, VOID_REASON = ?, VOIDED_AT = ?, VOIDED_BY = ? WHERE INVOICE_NO = ?"
			args = []interface{}{input.Status, input.Reason, now, input.User, invoiceNo}
		} else if input.Status == "refunded" {
			query = "UPDATE SALES SET STATUS = ?, REFUND_REASON = ?, REFUNDED_AT = ?, REFUNDED_BY = ? WHERE INVOICE_NO = ?"
			args = []interface{}{input.Status, input.Reason, now, input.User, invoiceNo}
		} else {
			query = "UPDATE SALES SET STATUS = ? WHERE INVOICE_NO = ?"
			args = []interface{}{input.Status, invoiceNo}
		}

		res, err := db.Exec(query, args...)
		if err != nil {
			log.Printf("[Error] UpdateSaleStatus failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengupdate status transaksi")
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			SendError(c, http.StatusNotFound, "Transaksi tidak ditemukan")
			return
		}

		SendSuccess(c, "Status transaksi berhasil diupdate", nil)
	}
}
