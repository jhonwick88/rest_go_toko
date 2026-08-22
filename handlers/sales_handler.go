package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetSales returns all sales headers
func GetSales(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := GetPaginationParams(c)

		query := "SELECT ID, INVOICE_NO, DATE, CASHIER, SUBTOTAL, DISCOUNT, GRAND_TOTAL, TAX, PAYMENT_METHOD, PAID_AMOUNT, CHANGE_AMOUNT, STATUS, VOID_REASON, VOIDED_AT, VOIDED_BY, REFUND_REASON, REFUNDED_AT, REFUNDED_BY, CUSTOMER_ID, IS_VOID FROM SALES ORDER BY DATE DESC LIMIT ? OFFSET ?"
		rows, err := db.Query(query, limit, offset)
		if err != nil {
			log.Printf("[Error] Query GetSales failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengambil data penjualan")
			return
		}
		defer rows.Close()

		var sales []models.Sales
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
			
			if customerID.Valid {
				cid := int(customerID.Int64)
				s.CustomerID = &cid
			}

			// Fetch items for this sale
			itemsQuery := "SELECT ID, INVOICE_NO, ITEMNO, ITEMNAME, ITEMUPC, QTY, PRICE, DISCOUNT, SUBTOTAL, NOTE FROM SALES_ITEMS WHERE INVOICE_NO = ?"
			itemRows, err := db.Query(itemsQuery, s.ReceiptNo)
			if err != nil {
				log.Printf("[Error] Query GetSales items failed: %v", err)
			} else {
				var items []models.SalesItem
				for itemRows.Next() {
					var it models.SalesItem
					var id sql.NullInt64
					var saleId, itemNo, itemName, itemUpc, note sql.NullString
					var qty, price, discount, subtotal sql.NullFloat64

					if err := itemRows.Scan(&id, &saleId, &itemNo, &itemName, &itemUpc, &qty, &price, &discount, &subtotal, &note); err == nil {
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
						items = append(items, it)
					}
				}
				itemRows.Close()
				s.Items = items
			}
			if s.Items == nil {
				s.Items = []models.SalesItem{}
			}

			sales = append(sales, s)
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

		// Insert Items
		var maxID int
		_ = tx.QueryRow("SELECT COALESCE(MAX(ID), 0) FROM SALES_ITEMS").Scan(&maxID)

		itemQuery := "INSERT INTO SALES_ITEMS (ID, INVOICE_NO, ITEMNO, ITEMNAME, ITEMUPC, QTY, PRICE, DISCOUNT, SUBTOTAL, NOTE) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		for i, item := range req.Items {
			maxID++
			_, err = tx.Exec(itemQuery, maxID, req.ReceiptNo, item.ItemNo, item.ItemName, item.ItemUPC, item.Qty, item.Price, item.Discount, item.Total, item.Note)
			if err != nil {
				tx.Rollback()
				log.Printf("[Error] Insert Sales Item failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal menyimpan data barang transaksi")
				return
			}
			req.Items[i].ID = maxID
			req.Items[i].SaleID = req.ReceiptNo
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
