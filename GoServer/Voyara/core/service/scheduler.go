package service

import (
	"fmt"
	"log"
	"time"
)

// StartPaymentTimeoutScheduler runs every 5 minutes to cancel unpaid orders older than 30 minutes.
func StartPaymentTimeoutScheduler() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			if err := cancelExpiredOrders(); err != nil {
				log.Printf("[VOYARA] Payment timeout scheduler error: %v", err)
			}
		}
	}()
	log.Println("[VOYARA] Payment timeout scheduler started (interval: 5min)")
}

func cancelExpiredOrders() error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %v", err)
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT id FROM voyara_orders
		WHERE payment_status = 'pending' AND created_at < NOW() - INTERVAL 30 MINUTE
		FOR UPDATE`)
	if err != nil {
		return fmt.Errorf("query expired orders: %v", err)
	}
	defer rows.Close()

	var cancelledCount int
	for rows.Next() {
		var orderID int
		if err := rows.Scan(&orderID); err != nil {
			continue
		}
		_, err := tx.Exec(`UPDATE voyara_orders SET payment_status = 'cancelled', cancelled_at = NOW() WHERE id = ? AND payment_status = 'pending'`, orderID)
		if err != nil {
			log.Printf("[VOYARA] Failed to cancel order %d: %v", orderID, err)
			continue
		}
		cancelledCount++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %v", err)
	}

	if cancelledCount > 0 {
		log.Printf("[VOYARA] Cancelled %d expired unpaid orders", cancelledCount)
	}
	return nil
}
