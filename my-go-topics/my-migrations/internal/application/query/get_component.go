package query

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
)

type GetComponentQuery struct {
	TenantSchema string
	ID           uuid.UUID
}

type GetComponentResult struct {
	ID              uuid.UUID
	Name            string
	Code            string
	Description     string
	Manufacturer    string
	ManufacturerURL string
	Price           float64
	WeightG         int
	Quantity        int
	ImageURL        string
	CreatedAt       time.Time
}

type GetComponentHandler struct {
	repo catalog.ComponentReadRepository
}

func NewGetComponentHandler(repo catalog.ComponentReadRepository) *GetComponentHandler {
	return &GetComponentHandler{repo: repo}
}

func (h *GetComponentHandler) Handle(ctx context.Context, q GetComponentQuery) (*GetComponentResult, error) {
	c, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	return &GetComponentResult{
		ID:              c.ID,
		Name:            c.Name,
		Code:            c.Code,
		Description:     c.Description,
		Manufacturer:    c.Manufacturer,
		ManufacturerURL: c.ManufacturerURL,
		Price:           c.Price,
		WeightG:         c.WeightG,
		Quantity:        c.Quantity,
		ImageURL:        c.ImageURL,
		CreatedAt:       c.CreatedAt,
	}, nil
}
