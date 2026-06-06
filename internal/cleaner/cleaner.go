package cleaner

import (
	"context"
	"log"
	"time"

	"seasons/backend/gen/dbstore"
)

func StartSessionCleaner(ctx context.Context, queries dbstore.Querier, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := queries.DeleteExpiredSessions(ctx); err != nil {
					log.Printf("session cleaner: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
