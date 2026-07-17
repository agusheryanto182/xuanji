package v1

import (
	"errors"
	"time"

	"github.com/agusheryanto182/redis-playground/internal/controller/restapi/v1/request"
	"github.com/agusheryanto182/redis-playground/internal/controller/restapi/v1/response"
	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/internal/usecase/product"
	"github.com/gofiber/fiber/v2"
)

func (r *V1) Store(ctx *fiber.Ctx) error {
	var body request.Store

	if err := r.bindAndValidate(ctx, &body); err != nil {
		r.l.Error(err, "restapi - v1 - store")
		return errorResponse(ctx, fiber.StatusBadRequest, response.ErrInvalidRequestBody, entity.ErrorInvalidRequestBody.Error())
	}

	product := &entity.Product{
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Stock:       body.Stock,
	}

	result, err := r.p.Store(ctx.UserContext(), product)
	if err != nil {
		return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
	}

	return successResponse(ctx, fiber.StatusCreated, result, nil)
}

func (r *V1) GetAll(ctx *fiber.Ctx) error {
	limit, offset := parseLimitAndOffset(ctx, 15, 0)

	products, total, err := r.p.GetAll(ctx.UserContext(), limit, offset)
	if err != nil {
		return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
	}

	meta := &response.Meta{
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	return successResponse(ctx, fiber.StatusOK, products, meta)
}

func (r *V1) GetByID(ctx *fiber.Ctx) error {
	productID, err := parseUUIDParam(ctx, "id")
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrIdIsRequired):
			return errorResponse(ctx, fiber.StatusBadRequest, response.ErrIdIsRequired, entity.ErrIdIsRequired.Error())

		default:
			r.l.Error(err, "restapi - v1- getByID")
			return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
		}
	}

	product, err := r.p.GetByID(ctx.UserContext(), productID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return errorResponse(ctx, fiber.StatusNotFound, response.ErrNotFound, entity.ErrProductNotFound.Error())

		default:
			return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
		}
	}

	return successResponse(ctx, fiber.StatusOK, product, nil)
}

func (r *V1) Update(ctx *fiber.Ctx) error {
	productID, err := parseUUIDParam(ctx, "id")
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrIdIsRequired):
			return errorResponse(ctx, fiber.StatusBadRequest, response.ErrIdIsRequired, entity.ErrIdIsRequired.Error())

		default:
			r.l.Error(err, "restapi - v1- getByID")
			return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
		}
	}

	var body request.Update

	if err := r.bindAndValidate(ctx, &body); err != nil {
		r.l.Error(err, "restapi - v1 - update")
		return errorResponse(ctx, fiber.StatusBadRequest, response.ErrInvalidRequestBody, entity.ErrorInvalidRequestBody.Error())
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
			return errorResponse(ctx, fiber.StatusNotFound, response.ErrNotFound, entity.ErrProductNotFound.Error())

		default:
			return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
		}
	}

	return successResponse(ctx, fiber.StatusOK, result, nil)
}

func (r *V1) Patch(ctx *fiber.Ctx) error {
	productID, err := parseUUIDParam(ctx, "id")
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrIdIsRequired):
			return errorResponse(ctx, fiber.StatusBadRequest, response.ErrIdIsRequired, entity.ErrIdIsRequired.Error())

		default:
			r.l.Error(err, "restapi - v1- getByID")
			return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
		}
	}

	var body request.Patch

	if err := r.bindAndValidate(ctx, &body); err != nil {
		r.l.Error(err, "restapi - v1 - patch")
		return errorResponse(ctx, fiber.StatusBadRequest, response.ErrInvalidRequestBody, entity.ErrorInvalidRequestBody.Error())
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
			return errorResponse(ctx, fiber.StatusNotFound, response.ErrNotFound, entity.ErrProductNotFound.Error())

		default:
			return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
		}
	}

	return ctx.Status(200).JSON(result)
}

func (r *V1) Delete(ctx *fiber.Ctx) error {
	productID, err := parseUUIDParam(ctx, "id")
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrIdIsRequired):
			return errorResponse(ctx, fiber.StatusBadRequest, response.ErrIdIsRequired, entity.ErrIdIsRequired.Error())

		default:
			r.l.Error(err, "restapi - v1- getByID")
			return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
		}
	}

	if err := r.p.Delete(ctx.UserContext(), productID.String()); err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return errorResponse(ctx, fiber.StatusNotFound, response.ErrNotFound, entity.ErrProductNotFound.Error())

		default:
			return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
		}
	}

	return successResponse[any](
		ctx,
		fiber.StatusOK,
		nil,
		nil,
	)
}
