package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go-temporal-workflow/forms"
	"go-temporal-workflow/security/tokens"
	"go-temporal-workflow/services/users"
	"net/http"
	"strings"
)

type UsersHandlers interface {
	SignUp(ctx *fiber.Ctx) error
	SignIn(ctx *fiber.Ctx) error
}

type usersHandlers struct {
	usersService users.Service
}

func NewUsersHandlers(usersService users.Service, app *fiber.App) {
	handler := &usersHandlers{usersService: usersService}
	app.Post("/signup", handler.PostSignUp)
	app.Post("/signin", handler.PostSignIn)
	app.Get("/access", handler.GetAccess)
	app.Post("/deposit", handler.PostDeposit)
}

func (h *usersHandlers) PostSignUp(ctx *fiber.Ctx) error {
	form := new(forms.SignUpInput)
	err := ctx.BodyParser(form)
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(fiber.Map{"error": err.Error()})
	}
	form.Email = strings.ToLower(strings.TrimSpace(form.Email))
	err = h.usersService.SignUp(ctx.Context(), form)
	if err != nil {
		return ctx.
			Status(http.StatusUnprocessableEntity).
			JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.
		Status(http.StatusCreated).
		JSON(fiber.Map{"status": "created"})
}

func (h *usersHandlers) PostSignIn(ctx *fiber.Ctx) error {
	form := new(forms.SignInInput)
	err := ctx.BodyParser(form)
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(fiber.Map{"error": err.Error()})
	}
	form.Email = strings.ToLower(strings.TrimSpace(form.Email))
	out, err := h.usersService.SignIn(ctx.Context(), form)
	if err != nil {
		return ctx.
			Status(http.StatusUnprocessableEntity).
			JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.
		Status(http.StatusOK).
		JSON(out)
}

func (h *usersHandlers) GetAccess(ctx *fiber.Ctx) error {
	token := ctx.Request().Header.Peek("Authorization")
	payload, err := tokens.Parse(string(token))
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(fiber.Map{"error": err.Error()})
	}
	user, err := h.usersService.GetUser(ctx.Context(), payload.UserID)
	if err != nil {
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.
		Status(http.StatusOK).
		JSON(user)
}

func (h *usersHandlers) PostDeposit(ctx *fiber.Ctx) error {
	token := ctx.Request().Header.Peek("Authorization")
	_, err := tokens.Parse(string(token))
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(fiber.Map{"error": err.Error()})
	}
	var form forms.DepositInput
	err = ctx.BodyParser(&form)
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(fiber.Map{"error": err.Error()})
	}
	user, err := h.usersService.Deposit(ctx.Context(), &form)
	if err != nil {
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.
		Status(http.StatusOK).
		JSON(user)
}
