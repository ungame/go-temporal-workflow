package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go-temporal-workflow/forms"
	"go-temporal-workflow/rmq"
	"go-temporal-workflow/security/tokens"
	"go-temporal-workflow/services/subscriptions"
	"net/http"
)

type SubscriptionsHandlers interface {
}

type subscriptionsHandlers struct {
	publisher            rmq.Publisher
	subscriptionsService subscriptions.Service
}

func NewSubscriptionsHandler(publisher rmq.Publisher, subscriptionsService subscriptions.Service, app *fiber.App) {
	h := subscriptionsHandlers{publisher: publisher, subscriptionsService: subscriptionsService}

	app.Post("/subscriptions", h.PostSubscribe)
	app.Get("/subscriptions/workflows/:id", h.GetWorkflow)
	app.Put("/subscriptions/workflows/:id/cancel", h.PutCancelWorkflow)
}

func (h *subscriptionsHandlers) PostSubscribe(ctx *fiber.Ctx) error {
	token := ctx.Request().Header.Peek("Authorization")
	payload, err := tokens.Parse(string(token))
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(fiber.Map{"error": err.Error()})
	}

	msg := forms.SubscribeInput{
		UserID: payload.UserID,
	}

	err = h.publisher.Send(&rmq.PublisherOptions{
		ExchangeName: "subscription",
		Persistent:   true,
		Message:      rmq.NewMessage(msg),
	})
	if err != nil {
		return ctx.
			Status(http.StatusInternalServerError).
			JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.
		Status(http.StatusOK).
		JSON(fiber.Map{"status": "sent"})
}

func (h *subscriptionsHandlers) GetWorkflow(ctx *fiber.Ctx) error {
	token := ctx.Request().Header.Peek("Authorization")
	_, err := tokens.Parse(string(token))
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(fiber.Map{"error": err.Error()})
	}

	out, err := h.subscriptionsService.GetWorkflow(ctx.Context(), ctx.Params("id"))
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.
		Status(http.StatusOK).
		JSON(out)
}

func (h *subscriptionsHandlers) PutCancelWorkflow(ctx *fiber.Ctx) error {
	token := ctx.Request().Header.Peek("Authorization")
	payload, err := tokens.Parse(string(token))
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(fiber.Map{"error": err.Error()})
	}

	msg := forms.CancelSubscriptionInput{
		UserID: payload.UserID,
		SubID:  ctx.Params("id"),
	}

	err = h.publisher.Send(&rmq.PublisherOptions{
		ExchangeName: "subscription",
		Persistent:   true,
		Message:      rmq.NewMessage(msg),
	})
	if err != nil {
		return ctx.
			Status(http.StatusInternalServerError).
			JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.
		Status(http.StatusOK).
		JSON(fiber.Map{"status": "sent"})
}
