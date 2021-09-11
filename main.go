package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	// echoSwagger "github.com/swaggo/echo-swagger"
	// _ "github.com/arindam-modak/get-covid-data/docs"
)

func main() {
	fmt.Println("Hello Covid! Please go away now.")
	e := echo.New()
	e.Use(middleware.CORS())
	// e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from web server")
	})

	e.GET("/fetch-data-and-save", fetchDataAndSave)

	e.GET("/get-data-from-location", getDataFromLocation)

	e.Start(":" + os.Getenv("PORT"))
}
