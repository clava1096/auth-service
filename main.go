package main

import (
	"auth-service/config"
	"auth-service/connections"
	_ "auth-service/docs"
	"auth-service/models"
	"auth-service/repositories"
	"auth-service/routers"
	"auth-service/services"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"os"
	"os/signal"
	"syscall"
)

// @title Тестовое задание на позицию Junior Backend Developer
// @version 1.0
// @description JWT-авторизация с refresh-токенами.
// @description Все ошибки возвращаются в формате:
// @description ```json
// @description {"error": "описание ошибки"}
// @description ```
// @termsOfService http://swagger.io/terms/
// @contact.name Vyacheslav
// @contact.email obvintseff.vyacheslav@yandex.ru
// @license.name UNLICENSED
// @license.url https://medods.yonote.ru/share/1982193d-43fc-4075-a608-cc0687c5eac2/doc/testovoe-zadanie-na-poziciyu-junior-backend-developer-6iFFklIyMI
// @host 127.0.0.1:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	c := config.GetConfig()
	app := Setup(c)

	if !fiber.IsChild() {

		models.Migrate()
	}

	go func() {
		if err := app.Listen("0.0.0.0:" + c.Application.Port); err != nil {
			log.Fatal(err)
		}
	}()

	// graceful shutdown (не дописанный) :D
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	_ = <-ch
	log.Info("Graceful shutdown")
	_ = app.Shutdown()
	fmt.Println("Running cleanup tasks...")
	// db.Close()
	fmt.Println("Fiber was successful shutdown.")
}

func CheckConnections(err error) {
	if err != nil {
		log.Fatalf("Fatal connection error: %v", err)
	}
}

func Setup(c *config.Config) *fiber.App {
	_, err := config.Load("config/config.yml")
	CheckConnections(err)
	CheckConnections(connections.ConnectPostgres())

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:8080, http://127.0.0.1:8080",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	logg := logger.New(logger.Config{
		Format: "${ip}\t- -\t[${time}]\t\"${method} ${path} ${protocol}\" ${status} ${bytesSent} ${referer} ${ua} ${latency} ${error}\n",
	})
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	app.Use(logg)
	api := app.Group("/api")

	TokenRepository := repositories.NewTokenRepository()
	TokenService := services.NewTokenService(TokenRepository, *c)
	userRepo := repositories.NewUserRepository()
	userService := services.NewUserService(userRepo)
	handler := routers.NewTokenHandler(TokenService, userService)
	Route(api, handler)

	return app
}

func Route(api fiber.Router, h *routers.TokenH) {
	api.Get("/tokens", h.TokenHandler)
	api.Post("/refresh", h.RefreshTokenHandler)
	api.Post("/me", h.GetUser)
	api.Post("/logout", h.Logout)
	api.Get("/get-users-GUID", h.GetAllUsers) // этот маршрут сделан для проверяющего!
}
