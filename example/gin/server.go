package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-swagno/swagno"
	"github.com/go-swagno/swagno-gin/swagger"
	"github.com/go-swagno/swagno/components/endpoint"
	"github.com/go-swagno/swagno/components/http/response"
	"github.com/go-swagno/swagno/components/mime"
	"github.com/go-swagno/swagno/components/parameter"
	"github.com/go-swagno/swagno/example/models"
)

func main() {
	sw := swagno.New(swagno.Config{Title: "Testing API", Version: "v1.0.0"})
	desc := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed id malesuada lorem, et fermentum sapien. Vivamus non pharetra risus, in efficitur leo. Suspendisse sed metus sit amet mi laoreet imperdiet. Donec aliquam eros eu blandit feugiat. Quisque scelerisque justo ac vehicula bibendum. Fusce suscipit arcu nisl, eu maximus odio consequat quis. Curabitur fermentum eleifend tellus, lobortis hendrerit velit varius vitae."

	endpoints := []*endpoint.EndPoint{
		endpoint.New(
			endpoint.GET,
			"/product",
			endpoint.WithTags("product"),
			endpoint.WithSuccessfulReturns([]response.Response{models.Product{}}),
			endpoint.WithErrors([]response.Response{models.UnsuccessfulResponse{}}),
			endpoint.WithDescription(desc),
			endpoint.WithProduce([]mime.MIME{mime.JSON, mime.XML}),
			endpoint.WithConsume([]mime.MIME{mime.JSON}),
			endpoint.WithSummary("this is a test summary"),
		),
		endpoint.New(
			endpoint.GET,
			"/product/{id}",
			endpoint.WithTags("product"),
			endpoint.WithParams(parameter.IntParam("id", parameter.Path, parameter.WithRequired())),
			endpoint.WithSuccessfulReturns([]response.Response{models.SuccessfulResponse{}}),
			endpoint.WithErrors([]response.Response{models.UnsuccessfulResponse{}}),
			endpoint.WithProduce([]mime.MIME{mime.JSON, mime.XML}),
		),
	}

	sw.AddEndpoints(endpoints)

	app := gin.Default()
	app.GET("/swagger/*any", swagger.SwaggerHandler(sw.MustToJson()))

	app.Run()
}
