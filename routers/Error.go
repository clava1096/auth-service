package routers

import (
	"auth-service/models"
	"github.com/gofiber/fiber/v2"
)

func ErrorResponse(ctx *fiber.Ctx, err string, code int) error {
	return ctx.Status(code).JSON(models.ErrorResponse{
		Error: err,
	})
}
