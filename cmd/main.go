package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/go-playground/validator/v10"

	"spgo/generated"
	"spgo/handler"
	"spgo/repository"
	"spgo/service"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	var server generated.ServerInterface = newServer()

	generated.RegisterHandlers(e, server)
	e.Use(middleware.Logger())
	e.Logger.Fatal(e.Start(":1323"))
}

func newServer() *handler.Server {
	LoadConfig()
	var repo repository.RepositoryInterface = repository.NewRepository(repository.NewRepositoryOptions{
		Db: postgesDB,
	})

	var serv service.ServiceInterface = service.NewService(service.NewServiceOptions{
		Repository: repo,
		Db:         postgesDB,
	})

	return handler.NewServer(
		handler.NewServerOptions{
			Service: serv,
		},
	)
}
