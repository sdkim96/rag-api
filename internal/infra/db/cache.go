package db

import (
	"context"
	"time"
)

// 30 Days
const defaultExpiredTTL = 60 * 60 * 24 * 30

func (e *Engine) UpsertCache(ctx context.Context, key string, value []byte, ttlSeconds *int) error {
	ttl := defaultExpiredTTL
	if ttlSeconds != nil {
		ttl = *ttlSeconds
	}
	expiresAt := time.Now().Add(time.Duration(ttl) * time.Second)
	return write(ctx, e.conn, `
INSERT INTO cache (key, value, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE
    SET value      = EXCLUDED.value,
        expires_at = EXCLUDED.expires_at,
        created_at = NOW()
`, key, value, expiresAt)
}

func (e *Engine) GetCache(ctx context.Context, key string) ([]byte, error) {
	return read(ctx, e.conn, `
SELECT value 
  FROM cache
 WHERE key = $1
   AND (expires_at IS NULL OR expires_at > NOW())
`, key)
}

func (e *Engine) CleanupCache(ctx context.Context) error {
	_, err := e.conn.ExecContext(ctx, `
DELETE FROM cache
 WHERE expires_at < NOW()
`)
	return err
}
