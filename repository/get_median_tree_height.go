package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"spgo/util"
)

func (r *Repository) GetMedianTreeHeight(ctx context.Context, estateID uuid.UUID) (int, error) {
	var medianTreeHeight sql.NullFloat64

	tx := util.GetTxFromContext(ctx, r.Db)

	query := `
        WITH RankedHeights AS (
            SELECT
                tree_height,
                ROW_NUMBER() OVER (ORDER BY tree_height) AS row_num,
                COUNT(*) OVER () AS total_count
            FROM plots
            WHERE estate_id = ?
        ),
        Median AS (
            SELECT
                CASE
                    WHEN total_count % 2 = 1 THEN tree_height
                    ELSE (SELECT AVG(tree_height) FROM RankedHeights WHERE row_num IN (total_count / 2, total_count / 2 + 1))
                END AS median_tree_height
            FROM RankedHeights
            WHERE row_num = (total_count + 1) / 2
        )
        SELECT median_tree_height FROM Median
    `

	// Execute raw SQL query
	if err := tx.Raw(query, estateID).Scan(&medianTreeHeight).Error; err != nil {
		return 0, err
	}

	// Check if medianTreeHeight is null
	if !medianTreeHeight.Valid {
		return 0, nil // Return 0 or some default value as per your requirement
	}

	return int(medianTreeHeight.Float64), nil
}
