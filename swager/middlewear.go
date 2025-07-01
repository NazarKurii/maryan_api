package openapi

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gin-gonic/gin"
)

func loadSpec(path string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true // support for $ref
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	if err := doc.Validate(loader.Context); err != nil {
		return nil, err
	}

	return doc, nil
}

func ValidateRequestMiddleware(pathToOpenApiYamlFile string) (gin.HandlerFunc, error) {
	// Load and validate OpenAPI doc
	doc, err := loadSpec(pathToOpenApiYamlFile)
	if err != nil {
		return nil, err
	}

	// Create a router from the spec
	router, err := legacy.NewRouter(doc) // Or routers.NewRouter(doc) if you're using the new router
	if err != nil {
		return nil, err
	}

	// Return Gin middleware
	return func(c *gin.Context) {
		// Match request to OpenAPI route
		route, pathParams, err := router.FindRoute(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "route not found in OpenAPI spec"})
			return
		}

		// Validate request against the OpenAPI schema
		input := &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
		}
		if err := openapi3filter.ValidateRequest(c.Request.Context(), input); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Pass to next middleware
		c.Next()
	}, nil
}
