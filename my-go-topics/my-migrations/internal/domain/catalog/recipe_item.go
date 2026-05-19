package catalog

import "github.com/google/uuid"

type RecipeItem struct {
	ComponentID uuid.UUID
	Quantity    int
}
