package api

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/wI2L/fizz"
	"go.uber.org/zap"
	"net/http"
)

func OpenAPISpecHandler(f *fizz.Fizz, logger *zap.Logger) func(*gin.Context) {
	return func(c *gin.Context) {
		spec := f.Generator().API()
		specBytes, _ := json.Marshal(spec)
		loader := openapi3.NewLoader()
		doc, _ := loader.LoadFromData(specBytes)
		if err := doc.Validate(loader.Context); err != nil {
			logger.Error("Fizz generated an invalid openapi v3 spec", zap.Error(err))
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		patchConditions(doc)
		patchExpression(doc)

		// point rule model to the newly added expression schema
		doc.Components.Schemas["VersionedRule"].Value.Properties["expression"] = &openapi3.SchemaRef{Ref: "#/components/schemas/Expression"}
		doc.Components.Schemas["Rule"].Value.Properties["expression"] = &openapi3.SchemaRef{Ref: "#/components/schemas/Expression"}

		// Define the security scheme with additional headers
		doc.Components.SecuritySchemes = openapi3.SecuritySchemes{
			"JWTAuth": &openapi3.SecuritySchemeRef{
				Value: openapi3.NewJWTSecurityScheme().WithBearerFormat("JWT"),
			},
			"ApiKeyAuth": &openapi3.SecuritySchemeRef{
				Value: &openapi3.SecurityScheme{
					Type:        "apiKey",
					In:          "header",
					Name:        "X-API-KEY",
					Description: "Make sure to include the X-ORG-ID header when using this API key.",
				},
			},
		}

		// FIXME: should we validate the override result?

		c.JSON(200, doc)
	}
}

// patchExpression is required because we use protobuf conversion for pass-through of expression value
func patchExpression(doc *openapi3.T) {
	exprArray := &openapi3.Schema{
		Properties: map[string]*openapi3.SchemaRef{
			"expression": {Value: &openapi3.Schema{
				Type:  "array",
				Items: &openapi3.SchemaRef{Ref: "#/components/schemas/Expression"},
			}},
		},
	}
	doc.Components.Schemas["And"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "object",
			Properties: map[string]*openapi3.SchemaRef{
				"and": {
					Value: exprArray,
				},
			},
		},
	}
	doc.Components.Schemas["Or"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "object",
			Properties: map[string]*openapi3.SchemaRef{
				"or": {
					Value: exprArray,
				},
			},
		},
	}
	doc.Components.Schemas["Expression"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "object",
			OneOf: []*openapi3.SchemaRef{
				{Ref: "#/components/schemas/And"},
				{Ref: "#/components/schemas/Or"},
				{Ref: "#/components/schemas/Condition"},
			},
		},
	}
}

// patchConditions is required, because fizz cannot translate a string constant type to an enum
func patchConditions(doc *openapi3.T) {
	doc.Components.Schemas["ConditionType"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "string",
			Enum: []interface{}{
				"e",
				"hv",
				"eq",
				"neq",
				"px",
				"npx",
				"sx",
				"nsx",
				"in",
				"nin",
				"some",
				"all",
				"none",
				"rgx",
				"nrgx",
			},
		},
	}
	doc.Components.Schemas["DRef"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "object",
			Properties: map[string]*openapi3.SchemaRef{
				"src": {Value: openapi3.NewStringSchema()},
				"dst": {Value: openapi3.NewStringSchema()},
			},
		},
	}
	doc.Components.Schemas["ConditionRef"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "object",
			Properties: map[string]*openapi3.SchemaRef{
				"ref":   {Ref: "#/components/schemas/DRef"},
				"kind":  {Ref: "#/components/schemas/ConditionType"},
				"value": {Value: openapi3.NewBytesSchema()},
			},
		},
	}
	doc.Components.Schemas["ConditionKey"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "object",
			Properties: map[string]*openapi3.SchemaRef{
				"key":   {Value: openapi3.NewStringSchema()},
				"kind":  {Ref: "#/components/schemas/ConditionType"},
				"value": {Value: openapi3.NewBytesSchema()},
			},
		},
	}
	doc.Components.Schemas["Condition"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "object",
			Properties: map[string]*openapi3.SchemaRef{
				"condition": {
					Value: &openapi3.Schema{
						OneOf: []*openapi3.SchemaRef{
							{Ref: "#/components/schemas/ConditionKey"},
							{Ref: "#/components/schemas/ConditionRef"},
						},
					},
				},
			},
		},
	}
}
