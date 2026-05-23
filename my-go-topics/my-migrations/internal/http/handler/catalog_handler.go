package handler

import (
	"context"
	"errors"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
	httpapi "github.com/krzysztofkolcz/mymigrations/internal/http/api"
)

func (s *Server) CreateComponent(
	ctx context.Context,
	req httpapi.CreateComponentRequestObject,
) (httpapi.CreateComponentResponseObject, error) {

	id, err := s.createComponent.Handle(ctx, appcommand.CreateComponentCommand{
		TenantSchema:    tenantSchema(req.Params.XTenantID),
		Name:            req.Body.Name,
		Code:            req.Body.Code,
		Description:     strVal(req.Body.Description),
		Manufacturer:    strVal(req.Body.Manufacturer),
		ManufacturerURL: strVal(req.Body.ManufacturerUrl),
		Price:           float64Val(req.Body.Price),
		WeightG:         intVal(req.Body.WeightG),
		Quantity:        intVal(req.Body.Quantity),
		ImageURL:        strVal(req.Body.ImageUrl),
	})
	if err != nil {
		if errors.Is(err, catalog.ErrInvalidName) || errors.Is(err, catalog.ErrInvalidCode) {
			return httpapi.CreateComponent400JSONResponse{N400JSONResponse: badRequestError(err.Error())}, nil
		}
		return httpapi.CreateComponent500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.CreateComponent201JSONResponse{Id: id}, nil
}

func (s *Server) GetComponent(
	ctx context.Context,
	req httpapi.GetComponentRequestObject,
) (httpapi.GetComponentResponseObject, error) {

	result, err := s.getComponent.Handle(ctx, appquery.GetComponentQuery{
		TenantSchema: tenantSchema(req.Params.XTenantID),
		ID:           req.Id,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return httpapi.GetComponent404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		return httpapi.GetComponent500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.GetComponent200JSONResponse{
		Id:              result.ID,
		Name:            result.Name,
		Code:            result.Code,
		Description:     &result.Description,
		Manufacturer:    &result.Manufacturer,
		ManufacturerUrl: &result.ManufacturerURL,
		Price:           &result.Price,
		WeightG:         &result.WeightG,
		Quantity:        &result.Quantity,
		ImageUrl:        &result.ImageURL,
		CreatedAt:       &result.CreatedAt,
	}, nil
}

func (s *Server) CreateProduct(
	ctx context.Context,
	req httpapi.CreateProductRequestObject,
) (httpapi.CreateProductResponseObject, error) {

	var tags []string
	if req.Body.Tags != nil {
		tags = *req.Body.Tags
	}

	id, err := s.createProduct.Handle(ctx, appcommand.CreateProductCommand{
		TenantSchema: tenantSchema(req.Params.XTenantID),
		Name:         req.Body.Name,
		Description:  strVal(req.Body.Description),
		Price:        float64Val(req.Body.Price),
		Tags:         tags,
	})
	if err != nil {
		if errors.Is(err, catalog.ErrInvalidName) {
			return httpapi.CreateProduct400JSONResponse{N400JSONResponse: badRequestError(err.Error())}, nil
		}
		return httpapi.CreateProduct500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.CreateProduct201JSONResponse{Id: id}, nil
}

func (s *Server) GetProduct(
	ctx context.Context,
	req httpapi.GetProductRequestObject,
) (httpapi.GetProductResponseObject, error) {

	result, err := s.getProduct.Handle(ctx, appquery.GetProductQuery{
		TenantSchema: tenantSchema(req.Params.XTenantID),
		ID:           req.Id,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return httpapi.GetProduct404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		return httpapi.GetProduct500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	recipe := make([]httpapi.RecipeItem, len(result.Recipe))
	for i, item := range result.Recipe {
		recipe[i] = httpapi.RecipeItem{
			ComponentId: item.ComponentID,
			Quantity:    item.Quantity,
		}
	}

	return httpapi.GetProduct200JSONResponse{
		Id:          result.ID,
		Name:        result.Name,
		Description: &result.Description,
		Price:       &result.Price,
		Tags:        &result.Tags,
		Recipe:      &recipe,
		CreatedAt:   &result.CreatedAt,
	}, nil
}

func (s *Server) SetProductComponent(
	ctx context.Context,
	req httpapi.SetProductComponentRequestObject,
) (httpapi.SetProductComponentResponseObject, error) {
	productID := req.Id
	componentID := req.ComponentId

	err := s.setProductComponent.Handle(ctx,
		appcommand.SetProductComponentCommand{
			TenantSchema: tenantSchema(req.Params.XTenantID),
			ProductID:    productID,
			ComponentID:  componentID,
			Quantity:     req.Body.Quantity,
		})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return httpapi.SetProductComponent404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		if errors.Is(err, catalog.ErrInvalidQuantity) {
			return httpapi.SetProductComponent400JSONResponse{N400JSONResponse: badRequestError(err.Error())}, nil
		}
		return httpapi.SetProductComponent500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.SetProductComponent204Response{}, nil
}

func (s *Server) RemoveProductComponent(
	ctx context.Context,
	req httpapi.RemoveProductComponentRequestObject,
) (httpapi.RemoveProductComponentResponseObject, error) {
	err := s.removeProductComponent.Handle(ctx,
		appcommand.RemoveProductComponentCommand{
			TenantSchema: tenantSchema(req.Params.XTenantID),
			ProductID:    req.Id,
			ComponentID:  req.ComponentId,
		},
	)
	if err != nil {
		if errors.Is (err, domain.ErrNotFound){
			return httpapi.RemoveProductComponent404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		if errors.Is(err, catalog.ErrComponentNotFound) {
			return httpapi.RemoveProductComponent404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		return httpapi.RemoveProductComponent500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.RemoveProductComponent204Response{}, nil
}

// helpers for optional pointer fields from generated request types

func strVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func float64Val(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func intVal(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
