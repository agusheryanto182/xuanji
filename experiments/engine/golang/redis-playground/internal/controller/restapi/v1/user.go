package v1

import (
	"errors"

	"github.com/agusheryanto182/redis-playground/internal/controller/restapi/v1/request"
	"github.com/agusheryanto182/redis-playground/internal/controller/restapi/v1/response"
	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary     Register
// @Description Register a new user
// @ID          register
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body     request.Register true "Registration data"
// @Success     201     {object} entity.User
// @Failure     400     {object} response.Error
// @Failure     409     {object} response.Error
// @Failure     500     {object} response.Error
// @Router      /auth/register [post]
func (r *V1) register(ctx *fiber.Ctx) error {
	var body request.Register

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - register body parser")

		return errorResponse(ctx, fiber.StatusBadRequest, response.ErrInvalidRequestBody, entity.ErrorInvalidRequestBody.Error())
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - register validation")

		return errorResponse(ctx, fiber.StatusBadRequest, response.ErrInvalidValidation, entity.ErrInvalidValidation.Error())
	}

	user, err := r.u.Register(ctx.UserContext(), body.Username, body.Email, body.Password)
	if err != nil {
		r.l.Error(err, "restapi - v1 - uc register")

		if errors.Is(err, entity.ErrUserAlreadyExists) {
			return errorResponse(ctx, fiber.StatusConflict, response.ErrUserAlreadyExists, entity.ErrUserAlreadyExists.Error())
		}

		return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
	}

	return successResponse(
		ctx,
		fiber.StatusCreated,
		user,
		nil,
	)
}

// @Summary     Login
// @Description Authenticate user and get JWT token
// @ID          login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body     request.Login true "Login credentials"
// @Success     200     {object} response.Token
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     500     {object} response.Error
// @Router      /auth/login [post]
func (r *V1) login(ctx *fiber.Ctx) error {
	var body request.Login

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - login - body parser")

		return errorResponse(ctx, fiber.StatusBadRequest, response.ErrInvalidRequestBody, entity.ErrorInvalidRequestBody.Error())
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - register validation")

		return errorResponse(ctx, fiber.StatusBadRequest, response.ErrInvalidValidation, entity.ErrInvalidValidation.Error())
	}

	token, err := r.u.Login(ctx.UserContext(), body.Email, body.Password)
	if err != nil {
		r.l.Error(err, "restapi - v1 - uc login")

		if errors.Is(err, entity.ErrInvalidCredentials) {
			return errorResponse(ctx, fiber.StatusUnauthorized, response.ErrInvalidCredentials, entity.ErrInvalidCredentials.Error())
		}

		return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
	}

	return successResponse(
		ctx,
		fiber.StatusOK,
		token,
		nil,
	)
}

// @Summary     Get profile
// @Description Get current user profile
// @ID          profile
// @Tags        user
// @Produce     json
// @Success     200 {object} entity.User
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /user/profile [get]
func (r *V1) profile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, fiber.StatusUnauthorized, response.ErrUnauthorized, entity.ErrUnauthorized.Error())
	}

	user, err := r.u.GetUser(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - uc profile")

		if errors.Is(err, entity.ErrUserNotFound) {
			return errorResponse(ctx, fiber.StatusNotFound, response.ErrNotFound, entity.ErrUserNotFound.Error())
		}

		return errorResponse(ctx, fiber.StatusInternalServerError, response.ErrInternal, entity.ErrInternalServerError.Error())
	}

	return successResponse(
		ctx,
		fiber.StatusOK,
		user,
		nil,
	)
}
