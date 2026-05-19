package query

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
)

type GetProductQuery struct {
	TenantSchema string
	ID           uuid.UUID
}

type RecipeItemResult struct {
	ComponentID uuid.UUID
	Quantity    int
}

type GetProductResult struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       float64
	Tags        []string
	Recipe      []RecipeItemResult
	CreatedAt   time.Time
}

type GetProductHandler struct {
	repo catalog.ProductReadRepository
}

func NewGetProductHandler(repo catalog.ProductReadRepository) *GetProductHandler {
	return &GetProductHandler{repo: repo}
}

func (h *GetProductHandler) Handle(ctx context.Context, q GetProductQuery) (*GetProductResult, error) {
	p, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	recipe := make([]RecipeItemResult, len(p.Recipe))
	for i, item := range p.Recipe {
		recipe[i] = RecipeItemResult{ComponentID: item.ComponentID, Quantity: item.Quantity}
	}
	return &GetProductResult{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Tags:        p.Tags,
		Recipe:      recipe,
		CreatedAt:   p.CreatedAt,
	}, nil
}
