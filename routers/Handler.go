package routers

import (
	"auth-service/config"
	"auth-service/models"
	"auth-service/services"
	"auth-service/webhook"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"net/http"
	"time"
)

type TokenH struct {
	tokenService *services.TokenService
	userService  *services.UserService
}

func NewTokenHandler(tokenService *services.TokenService, userService *services.UserService) *TokenH {
	return &TokenH{
		tokenService: tokenService,
		userService:  userService,
	}
}

// GetAllUsers godoc
// @Summary Получить список всех пользователей. Этот маршрут добавлен для удобство проверяющего!
// @Description Возвращает список GUID всех пользователей в системе
// @Tags Пользователь
// @Accept json
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} models.ErrorResponse
// @Router /api/get-users-GUID [get]
//
// @ExampleRequest
// curl -X GET "http://localhost:8080/api/get-users-GUID"
//
// @ExampleResponse 200
// [
//
//	{"guid": "a1b2c3d4-e5f6-7890"},
//	{"guid": "b2c3d4e5-f6a7-8901"}
//
// ]
func (h *TokenH) GetAllUsers(ctx *fiber.Ctx) error {
	u, _ := h.userService.GetUsers()
	if len(u) == 0 {
		_, err := h.userService.NewUsers(config.GetConfig().Usr.Count)
		if err != nil {
			log.Fatalf("Error while create simple users: %s", err)
		}
	}
	users, err := h.userService.GetUsers()
	if err != nil {
		return ErrorResponse(ctx, "Internal Server Error", 500)
	}
	return ctx.Status(http.StatusOK).JSON(users)
}

// TokenHandler godoc
// @Summary Получить токены
// @Description Генерирует пару access/refresh токенов для пользователя
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param guid query string true "GUID пользователя"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/tokens [get]
//
// @ExampleRequest
// curl -X GET "http://localhost:8080/api/tokens?guid=a1b2c3d4-e5f6-7890"
//
// @ExampleResponse 200
//
//	{
//	  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
//	  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
//	}
//
// @ExampleResponse 400
// {"error": "Bad Request"}
//
// @ExampleResponse 403
// {"error": "Your refresh token is valid and time not expired"}
func (h *TokenH) TokenHandler(ctx *fiber.Ctx) error {
	guid := ctx.Query("guid")
	userAgent := ctx.Get("User-Agent")
	ip := ctx.IP()

	if guid == "" {
		return ErrorResponse(ctx, "Bad Request", 400)
	}
	if !h.userService.IsExist(guid) {
		return ErrorResponse(ctx, "User not found", 404)
	}

	refreshToken, err := h.tokenService.FindTokenByUserGUID(guid)
	if err != nil || refreshToken.ExpiresAt.Before(time.Now()) {
		if err == nil {
			_ = h.tokenService.DeleteTokenByID(refreshToken.ID)
		}
		access, refresh, err := h.tokenService.GenerateTokens(guid, userAgent, ip)
		if err != nil {
			return ErrorResponse(ctx, "Internal Server Error", 500)
		}
		return ctx.Status(http.StatusOK).JSON(models.NewTokenResponse(access, refresh))
	}

	return ErrorResponse(ctx, "Your refresh token is valid and time not expired", 403)
}

// RefreshTokenHandler godoc
// @Summary Обновить токены
// @Description Обновляет пару access/refresh токенов по валидному refresh токену
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param request body models.TokenRequest true "Запрос с токенами"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/refresh [post]
//
// @ExampleRequest
// curl -X POST "http://localhost:8080/api/refresh" \
// -H "Content-Type: application/json" \
// -d '{"access_token":"old_token", "refresh_token":"old_refresh_token"}'
//
// @ExampleResponse 200
//
//	{
//	  "access_token": "new_token",
//	  "refresh_token": "new_refresh_token"
//	}
//
// @ExampleResponse 403
// {"error": "Your User-Agent is edited, logout"}
func (h *TokenH) RefreshTokenHandler(ctx *fiber.Ctx) error {
	req, err := h.parseTokenRequest(ctx)
	if err != nil {
		return ErrorResponse(ctx, err.Error(), 400)
	}

	claims, err := h.parseAccessToken(req.AccessToken)
	if err != nil {
		return ErrorResponse(ctx, "Invalid access token", 400)
	}

	stored, err := h.getStoredRefreshToken(claims.Sub)
	if err != nil {
		return ErrorResponse(ctx, "Not Found!", 404)
	}

	if !h.isRefreshTokenValid(stored, req.RefreshToken, claims) {
		return ErrorResponse(ctx, "Invalid refresh/access token pair", 400)
	}

	userAgent := ctx.Get("User-Agent")
	ip := ctx.IP()

	if stored.UserAgent != userAgent {
		_ = h.tokenService.DeleteTokenByID(stored.ID)
		return ErrorResponse(ctx, "Your User-Agent is edited, logout", 403)
	}

	if stored.IpAddress != ip {
		attempt := webhook.LoginAttempt{
			UserGUID: userAgent,
			IP:       ip,
			Event:    "new_ip",
		}
		go func() {
			err = webhook.EditIpWebhook(config.GetConfig().Webhook.Url, attempt)
			if err != nil {
				log.Errorf("Failed to send webhook: %s", err.Error())
			}
		}()
	}

	_ = h.tokenService.DeleteTokenByID(stored.ID)
	access, refresh, err := h.tokenService.GenerateTokens(stored.UserGuid, userAgent, ip)
	if err != nil {
		return ErrorResponse(ctx, "Internal Server Error", 500)
	}

	return ctx.Status(http.StatusOK).JSON(models.NewTokenResponse(access, refresh))
}

// GetUser godoc
// @Summary Получить информацию о пользователе
// @Description Возвращает GUID пользователя по валидному access токену
// @Tags Пользователь
// @Accept json
// @Produce json
// @Param request body models.TokenRequest true "Запрос с access токеном"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/me [post]
//
// @ExampleRequest
// curl -X POST "http://localhost:8080/api/me" \
// -H "Content-Type: application/json" \
// -d '{"access_token":"your_token"}'
//
// @ExampleResponse 200
// {"user_guid": "a1b2c3d4-e5f6-7890"}
//
// @ExampleResponse 404
// {"error": "Not Found!"}
func (h *TokenH) GetUser(ctx *fiber.Ctx) error {
	req, err := h.parseTokenRequest(ctx)
	if err != nil {
		return ErrorResponse(ctx, err.Error(), 400)
	}

	claims, err := h.parseAccessToken(req.AccessToken)
	if err != nil {
		return ErrorResponse(ctx, "Invalid token", 400)
	}

	stored, err := h.getStoredRefreshToken(claims.Sub)
	if err != nil {
		return ErrorResponse(ctx, "Not Found!", 404)
	}

	return ctx.Status(http.StatusOK).JSON(models.NewUserResponse(stored.UserGuid))
}

// Logout godoc
// @Summary Выход из системы
// @Description Удаляет refresh токен пользователя
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param request body models.TokenRequest true "Запрос с токенами"
// @Success 200 {object} models.Logout
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/logout [post]
//
// @ExampleRequest
// curl -X POST "http://localhost:8080/api/logout" \
// -H "Content-Type: application/json" \
// -d '{"access_token":"your_token"}'
//
// @ExampleResponse 200
// {"msg": "Ok."}
//
// @ExampleResponse 500
// {"error": "Internal Server Error"}
func (h *TokenH) Logout(ctx *fiber.Ctx) error {
	req, err := h.parseTokenRequest(ctx)
	if err != nil {
		return ErrorResponse(ctx, err.Error(), 400)
	}

	claims, err := h.parseAccessToken(req.AccessToken)
	if err != nil {
		return ErrorResponse(ctx, "Invalid token", 400)
	}

	stored, err := h.getStoredRefreshToken(claims.Sub)
	if err != nil {
		return ErrorResponse(ctx, "Not Found!", 404)
	}

	if err := h.tokenService.DeleteTokenByID(stored.ID); err != nil {
		return ErrorResponse(ctx, "Internal Server Error", 500)
	}

	return ctx.Status(http.StatusOK).JSON(models.Logout{Msg: "Ok."})
}

func (h *TokenH) parseTokenRequest(ctx *fiber.Ctx) (*models.TokenRequest, error) {
	var req models.TokenRequest
	if err := ctx.BodyParser(&req); err != nil {
		return nil, errors.New("invalid request body")
	}
	return &req, nil
}

func (h *TokenH) parseAccessToken(token string) (*models.TokenClaims, error) {
	claims := models.GetClaims(token, config.GetConfig().Jwt.SecretKey)
	if claims == nil {
		return nil, errors.New("invalid access token")
	}
	return claims, nil
}

func (h *TokenH) getStoredRefreshToken(guid string) (*models.Token, error) {
	return h.tokenService.FindTokenByUserGUID(guid)
}

func (h *TokenH) isRefreshTokenValid(stored *models.Token, input string, claims *models.TokenClaims) bool {
	return h.tokenService.ValidateRefreshToken(stored.RefreshToken, input) &&
		h.tokenService.ValidateTokenPair(*claims, input)
}
