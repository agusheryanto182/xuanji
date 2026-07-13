package v1

import (
	"errors"
	"time"

	"github.com/agusheryanto182/redis-playground/internal/controller/restapi/v1/request"
	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/internal/usecase/product"
	"github.com/gofiber/fiber/v2"
)

func (r *V1) Store(ctx *fiber.Ctx) error {
	var body request.Store

	if err := r.bindAndValidate(ctx, &body); err != nil {
		return errorResponse(ctx, 400, err.Error())
	}

	product := &entity.Product{
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Stock:       body.Stock,
	}

	result, err := r.p.Store(ctx.UserContext(), product)
	if err != nil {
		return errorResponse(ctx, 500, entity.ErrInternalServerError.Error())
	}

	return ctx.Status(201).JSON(result)
}

func (r *V1) GetByID(ctx *fiber.Ctx) error {
	productID, err := parseUUIDParam(ctx, "id")
	if err != nil {
		return errorResponse(ctx, 400, err.Error())
	}

	product, err := r.p.GetByID(ctx.UserContext(), productID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return errorResponse(ctx, 404, err.Error())

		default:
			return errorResponse(ctx, 500, entity.ErrInternalServerError.Error())
		}
	}

	return ctx.Status(200).JSON(product)
}

func (r *V1) Update(ctx *fiber.Ctx) error {
	productID, err := parseUUIDParam(ctx, "id")
	if err != nil {
		return errorResponse(ctx, 400, err.Error())
	}

	var body request.Update

	if err := r.bindAndValidate(ctx, &body); err != nil {
		return errorResponse(ctx, 400, err.Error())
	}

	product := &entity.Product{
		ID:          productID,
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Stock:       body.Stock,
	}

	result, err := r.p.Update(ctx.UserContext(), product)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return errorResponse(ctx, 404, err.Error())

		default:
			return errorResponse(ctx, 500, entity.ErrInternalServerError.Error())
		}
	}

	return ctx.Status(200).JSON(result)
}

func (r *V1) Patch(ctx *fiber.Ctx) error {
	productID, err := parseUUIDParam(ctx, "id")
	if err != nil {
		return errorResponse(ctx, 400, err.Error())
	}

	var body request.Patch

	if err := r.bindAndValidate(ctx, &body); err != nil {
		return errorResponse(ctx, 400, err.Error())
	}

	product := product.PatchInput{
		ID:          productID,
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Stock:       body.Stock,
		UpdatedAt:   time.Now().UTC(),
	}

	result, err := r.p.Patch(ctx.UserContext(), product)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return errorResponse(ctx, 404, err.Error())

		default:
			return errorResponse(ctx, 500, entity.ErrInternalServerError.Error())
		}
	}

	return ctx.Status(200).JSON(result)
}

func (r *V1) Delete(ctx *fiber.Ctx) error {
	productID, err := parseUUIDParam(ctx, "id")
	if err != nil {
		return errorResponse(ctx, 400, err.Error())
	}

	if err := r.p.Delete(ctx.UserContext(), productID.String()); err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return errorResponse(ctx, 404, err.Error())

		default:
			return errorResponse(ctx, 500, entity.ErrInternalServerError.Error())
		}
	}

	return ctx.Status(204).Send(nil)
}
