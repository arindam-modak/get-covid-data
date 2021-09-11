package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

func main() {
	fmt.Println("Hello Covid! Please go away now.")
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from web server")
	})

	e.GET("/fetch-data-and-save", fetchDataAndSave)

	e.GET("/get-data-from-location", getDataFromLocation)

	e.Start(":" + os.Getenv("PORT"))
}
