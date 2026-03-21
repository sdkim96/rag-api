package db

import (
	"context"
	"fmt"
)

func (e *Engine) InsertSource(ctx context.Context, args ...any) error {
	return write(ctx, e.conn, `
INSERT INTO sources (id, owner_id, uri, mime_type, name, size, origin)
VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, args...)
}

func (e *Engine) SelectSource(ctx context.Context, id string) ([]byte, error) {
	return read(ctx, e.conn, `
SELECT row_to_json(t)
  FROM (
	SELECT s.id
	     , s.owner_id
	     , s.uri
	     , s.mime_type
	     , s.name
	     , s.size
	     , s.origin
	     , s.created_at
	     , i.status
	  FROM sources s
	  LEFT JOIN (
	    SELECT DISTINCT ON (source_id)
	           source_id, status
	      FROM indexing
	     ORDER BY source_id, created_at DESC
	  ) i ON s.id = i.source_id
	 WHERE s.id = $1
	   AND s.deleted_at IS NULL
  ) t
	`, id)
}

func (e *Engine) SelectSearchableSourceIDs(
	ctx context.Context,
	ownerID string,
) ([]string, error) {
	rows, err := search(ctx, e.conn, `
SELECT s.id
  FROM sources s
 INNER JOIN search sr ON s.id = sr.source_id
 INNER JOIN (
    SELECT DISTINCT ON (source_id)
           source_id, status
      FROM indexing
     ORDER BY source_id, created_at DESC
 ) i ON s.id = i.source_id
 WHERE s.owner_id  = $1
   AND s.deleted_at IS NULL
   AND i.status    = 'completed'
    `, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (e *Engine) SelectSources(
	ctx context.Context,
	id []string,
	offset, limit int,
	keyword string,
) ([][]byte, error) {
	query := `
SELECT row_to_json(t)
  FROM (
    SELECT s.id
         , s.owner_id
         , s.uri
         , s.mime_type
         , s.name
         , s.size
         , s.origin
         , s.created_at
         , i.status
      FROM sources s
      LEFT JOIN (
        SELECT DISTINCT ON (source_id)
               source_id, status
          FROM indexing
         ORDER BY source_id, created_at DESC
      ) i ON s.id = i.source_id
     WHERE s.deleted_at IS NULL`

	args := []any{}
	idx := 1

	if len(id) > 0 {
		query += fmt.Sprintf(" AND s.id = ANY($%d)", idx)
		args = append(args, id)
		idx++
	}

	if keyword != "" {
		query += fmt.Sprintf(" AND s.name ILIKE $%d", idx)
		args = append(args, "%"+keyword+"%")
		idx++
	}

	query += fmt.Sprintf(`
     ORDER BY s.created_at DESC
     OFFSET $%d LIMIT $%d
  ) t`, idx, idx+1)

	args = append(args, offset, limit)

	return readAll(ctx, e.conn, query, args...)
}

func (e *Engine) DeleteSource(ctx context.Context, id string) error {
	return write(ctx, e.conn, `
UPDATE sources
   SET deleted_at = NOW()
 WHERE id = $1
	`, id)
}
