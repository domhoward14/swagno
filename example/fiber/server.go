package main

import (
	"fmt"

	"github.com/go-swagno/swagno"
	"github.com/go-swagno/swagno-fiber/swagger"
	"github.com/go-swagno/swagno/components/endpoint"
	"github.com/go-swagno/swagno/components/mime"
	"github.com/go-swagno/swagno/example/models"
	"github.com/go-swagno/swagno/http/response"
	"github.com/gofiber/fiber/v2"
)

func main() {
	sw := swagno.New(swagno.Config{Title: "Testing API", Version: "v1.0.0"})
	desc := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed id malesuada lorem, et fermentum sapien. Vivamus non pharetra risus, in efficitur leo. Suspendisse sed metus sit amet mi laoreet imperdiet. Donec aliquam eros eu blandit feugiat. Quisque scelerisque justo ac vehicula bibendum. Fusce suscipit arcu nisl, eu maximus odio consequat quis. Curabitur fermentum eleifend tellus, lobortis hendrerit velit varius vitae."

	endpoints := []*endpoint.EndPoint{
		endpoint.New(
			endpoint.WithMethod(endpoint.GET),
			endpoint.WithPath("/product"),
			endpoint.WithTags("product"),
			endpoint.WithSuccessfulReturns([]response.Info{
				response.New("string", "200", "Product List"),                                                        // OK
				response.New(models.StructResponse1{}, "201", "Product List"),                                        // OK
				response.New(models.StructResponse1{InterfaceType: "string"}, "202", "Product List"),                 // OK
				response.New(models.StructResponse1{InterfaceType: models.StructResponse2{}}, "203", "Product List"), // OK
				response.New([]int{}, "204", "Product List"),                                                         // OK
				response.New([]*int{}, "205", "Product List"),                                                        // OK
				response.New([][]int{}, "206", "Product List"),                                                       // OK
				response.New([]models.StructResponse1{}, "207", "Product List"),                                      // OK
				response.New([]*models.StructResponse1{}, "208", "Product List"),                                     // OK
				response.New([]map[string]string{}, "209", "Product List"),                                           // OK
				response.New([]interface{}{}, "210", "Product List"),                                                 // OK
				response.New([]interface{}{models.StructResponse1{}}, "211", "Product List"),                         // not desired result
				response.New([][]models.StructResponse1{}, "212", "Product List"),                                    // OK
				response.New([][]*models.StructResponse1{}, "213", "Product List"),                                   // OK
				response.New(map[string]int{"field": 123}, "214", "Product List"),                                    // OK
				response.New(map[string]*int{}, "215", "Product List"),                                               // OK
				response.New(map[string]models.StructResponse1{}, "216", "Product List"),                             // OK
				response.New(map[string]*models.StructResponse1{}, "217", "Product List"),                            // OK
				response.New(map[string][]models.StructResponse1{}, "218", "Product List"),                           // OK
				response.New(map[string][]*models.StructResponse1{}, "219", "Product List"),                          // OK
				response.New(map[string]interface{}{}, "220", "Product List"),                                        // OK
				response.New(map[string]interface{}{"field": 123}, "221", "Product List"),                            // OK
			}),
			endpoint.WithDescription(desc),
			endpoint.WithProduce([]mime.MIME{mime.JSON, mime.XML}),
			endpoint.WithConsume([]mime.MIME{mime.JSON}),
			endpoint.WithSummary("this is a test summary"),
		),
	}

	sw.AddEndpoints(endpoints)

	app := fiber.New()
	swagger.SwaggerHandler(app, sw.GenerateDocs(), swagger.WithPrefix("/swagger"))

	fmt.Println(app.Listen(":8080"))
}
