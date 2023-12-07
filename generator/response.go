package generator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-swagno/swagno/components/parameter"
	"github.com/go-swagno/swagno/utils"
)

type ResponseGenerator struct {
}

func NewResponseGenerator() *ResponseGenerator {
	return &ResponseGenerator{}
}

func (g ResponseGenerator) GenerateJsonResponseScheme(model any) *parameter.JsonResponseSchema {
	hash := utils.GenerateHash(utils.ConvertToString(model))

	switch reflect.TypeOf(model).Kind() {
	case reflect.Slice:
		sliceElement := reflect.TypeOf(model).Elem()
		if sliceElement.Kind() == reflect.Pointer {
			sliceElement = sliceElement.Elem()
		}

		sliceElementKind := sliceElement.Kind()
		switch sliceElementKind {
		case reflect.Struct:
			return &parameter.JsonResponseSchema{
				Type: "array",
				Items: &parameter.JsonResponseSchemeItems{
					Ref: strings.ReplaceAll(fmt.Sprintf("#/definitions/%s_%s", reflect.TypeOf(model).Elem().String(), hash), "[]", ""),
				},
			}

		case reflect.Slice:
			// TODO: make this recursive, now support only [][]type
			innerElement := reflect.New(sliceElement.Elem()).Elem()
			if innerElement.Kind() == reflect.Pointer {
				innerElement = reflect.New(sliceElement.Elem().Elem()).Elem()
			}

			innerElementScheme := g.GenerateJsonResponseScheme(innerElement.Interface())
			items := &parameter.JsonResponseSchemeItems{}

			if innerElementScheme.Type != "" {
				items.Type = innerElementScheme.Type
			} else {
				items.Ref = strings.ReplaceAll(fmt.Sprintf("#/definitions/%s_%s", sliceElement.Elem().String(), hash), "[]", "")
			}

			return &parameter.JsonResponseSchema{
				Type: "array",
				Items: &parameter.JsonResponseSchemeItems{
					Type:  "array",
					Items: items,
				},
			}

		case reflect.Map:
			return &parameter.JsonResponseSchema{
				Type: "array",
				Items: &parameter.JsonResponseSchemeItems{
					Ref: strings.ReplaceAll(fmt.Sprintf("#/definitions/%s_%s", sliceElement.String(), hash), "[]", ""),
				},
			}

		case reflect.Interface:
			return &parameter.JsonResponseSchema{
				Type: "array",
				Items: &parameter.JsonResponseSchemeItems{
					Type: "object", // TODO: get real type of interface
				},
			}

		default:
			return &parameter.JsonResponseSchema{
				Type: "array",
				Items: &parameter.JsonResponseSchemeItems{
					Type: getType(sliceElementKind.String()),
				},
			}
		}

	case reflect.Map:
		ref := strings.ReplaceAll(fmt.Sprintf("#/definitions/%T_%s", model, hash), "[]", "")
		// preventing override map definitions, generate random name for map if its not typed
		if reflect.TypeOf(model).Name() == "" {
			ref = strings.ReplaceAll(fmt.Sprintf("#/definitions/%T_%s_%s", model, utils.GetHashOfMap(utils.ConvertInterfaceToMap(model)), hash), "[]", "")
		}
		return &parameter.JsonResponseSchema{
			Ref: ref,
		}

	default:
		if g.hasStructFields(model) {
			return &parameter.JsonResponseSchema{
				Ref: strings.ReplaceAll(fmt.Sprintf("#/definitions/%T_%s", model, hash), "[]", ""),
			}
		}
	}

	return &parameter.JsonResponseSchema{
		Type: getType(reflect.TypeOf(model).String()),
	}
}

func (g ResponseGenerator) hasStructFields(s interface{}) bool {
	rv := reflect.ValueOf(s)

	if rv.Kind() != reflect.Struct {
		return false
	}

	numFields := rv.NumField()
	return numFields > 0
}
