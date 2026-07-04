package v1

import (
	"fmt"

	"github.com/agusheryanto182/ecommerce-monolith/internal/controller/restapi/v1/request"
	"github.com/agusheryanto182/ecommerce-monolith/internal/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (r *V1) Store(ctx *fiber.Ctx) error {
	var body request.Store

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - Store Body Parser")

		return errorResponse(ctx, 400, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - Store Validation")

		return errorResponse(ctx, 400, "invalid request body")
	}

	product := &entity.Product{
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Stock:       body.Stock,
	}

	fmt.Printf("%#v\n", r.p)

	result, err := r.p.Store(ctx.UserContext(), product)
	if err != nil {
		r.l.Error(err, "restapi - v1 - Store")

		return errorResponse(ctx, 500, "internal server error")
	}

	return ctx.Status(201).JSON(result)
}

func (r *V1) GetByID(ctx *fiber.Ctx) error {
	var id = ctx.Params("id")

	if id == "" {
		return errorResponse(ctx, 400, "invalid request body")
	}

	productID, err := uuid.Parse(id)
	if err != nil {
		return errorResponse(ctx, 400, "invalid request body")
	}

	product, err := r.p.GetByID(ctx.UserContext(), productID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - Get")

		return errorResponse(ctx, 500, "internal server error")
	}

	return ctx.Status(200).JSON(product)
}

func (r *V1) Update(ctx *fiber.Ctx) error {
	return nil
}

func (r *V1) UpdatePartial(ctx *fiber.Ctx) error {
	return nil
}

func (r *V1) Delete(ctx *fiber.Ctx) error {
	return nil
}
