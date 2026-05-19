package catalog

import (
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/tenant"
)

func nullableText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func nullableInt4(i int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

func float64ToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(strconv.FormatFloat(f, 'f', 2, 64))
	return n
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, err := n.Float64Value()
	if err != nil {
		return 0
	}
	return f.Float64
}

func timestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func componentFromQueryRow(row querydb.Component) *catalog.Component {
	return &catalog.Component{
		ID:              row.ID,
		Name:            row.Name,
		Code:            row.Code,
		Description:     row.Description.String,
		Manufacturer:    row.Manufacturer.String,
		ManufacturerURL: row.ManufacturerUrl.String,
		Price:           numericToFloat64(row.Price),
		WeightG:         int(row.WeightG.Int32),
		Quantity:        int(row.Quantity),
		ImageURL:        row.ImageUrl.String,
		CreatedAt:       timestamptzToTime(row.CreatedAt),
	}
}

func productFromQueryRow(row querydb.Product, components []querydb.ProductComponent) *catalog.Product {
	recipe := make([]catalog.RecipeItem, len(components))
	for i, c := range components {
		recipe[i] = catalog.RecipeItem{
			ComponentID: c.ComponentID,
			Quantity:    int(c.Quantity),
		}
	}
	return &catalog.Product{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description.String,
		Price:       numericToFloat64(row.Price),
		Tags:        row.Tags,
		Recipe:      recipe,
		CreatedAt:   timestamptzToTime(row.CreatedAt),
	}
}
