package generator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-swagno/swagno/components/definition"
	"github.com/go-swagno/swagno/http/response"
	"github.com/go-swagno/swagno/utils"
)

type DefinitionGenerator struct {
	Definitions map[string]definition.Definition
}

func NewDefinitionGenerator(definitionMap map[string]definition.Definition) *DefinitionGenerator {
	return &DefinitionGenerator{
		Definitions: definitionMap,
	}
}
func (g DefinitionGenerator) getDefiniationName(t interface{}) string {
	hash := utils.GenerateHash(utils.ConvertToString(t))
	// create random names for maps
	if reflect.TypeOf(t).Kind() == reflect.Map && reflect.TypeOf(t).Name() == "" {
		return strings.ReplaceAll(fmt.Sprintf("%T_%s_%s", t, utils.GetHashOfMap(utils.ConvertInterfaceToMap(t)), hash), "[]", "")
	}
	return strings.ReplaceAll(fmt.Sprintf("%T_%s", t, hash), "[]", "")
}

func (g DefinitionGenerator) CreateDefinition(t interface{}, prefix string) {
	properties := make(map[string]definition.DefinitionProperties)
	definitionName := prefix + g.getDefiniationName(t)

	reflectReturn := reflect.TypeOf(t)
	switch reflectReturn.Kind() {
	case reflect.Slice:
		reflectReturn = reflectReturn.Elem()

		if reflectReturn.Kind() == reflect.Pointer {
			reflectReturn = reflectReturn.Elem()
		}

		if reflectReturn.Kind() == reflect.Struct {
			properties = g.createStructDefinitions(reflectReturn, t, definitionName+"_")
		} else if reflectReturn.Kind() == reflect.Slice {
			if reflectReturn.Elem().Kind() == reflect.Pointer {
				reflectReturn = reflectReturn.Elem()
			}
			sliceElement := reflect.New(reflectReturn.Elem()).Elem()

			if sliceElement.Kind() == reflect.Struct {
				properties = g.createStructDefinitions(reflect.TypeOf(sliceElement.Interface()), sliceElement.Interface(), definitionName+"_")
			}
		} else if reflectReturn.Kind() == reflect.Map {
			// TODO: this is not working well, need to get map values
			properties = g.createMapDefinitions(reflect.ValueOf(reflect.New(reflectReturn).Elem().Interface()), definitionName+"_")
		}
	case reflect.Struct:
		if reflectReturn == reflect.TypeOf(response.CustomResponseType{}) {
			// if CustomResponseType, use Model struct in it
			g.CreateDefinition(t.(response.CustomResponseType).Model, "")
			return
		}
		properties = g.createStructDefinitions(reflectReturn, t, definitionName+"_")
	case reflect.Map:
		properties = g.createMapDefinitions(reflect.ValueOf(t), definitionName+"_")

	default:
		g.Definitions[definitionName] = definition.Definition{
			Type:       getType(reflectReturn.String()),
			Properties: properties,
		}
		return
	}

	g.Definitions[definitionName] = definition.Definition{
		Type:       "object",
		Properties: properties,
	}
}

func (g DefinitionGenerator) createStructDefinitions(_struct reflect.Type, data any, prefix string) map[string]definition.DefinitionProperties {
	properties := make(map[string]definition.DefinitionProperties)
	for i := 0; i < _struct.NumField(); i++ {
		field := _struct.Field(i)
		fieldType := getType(field.Type.Kind().String())
		fieldJsonTag := getJsonTag(field)
		exampleTag := getExampleTag(field)

		// check swagno tag for custom type
		if field.Tag.Get("swagno") != "" {
			properties[fieldJsonTag] = g.getSwagnoTag(field)
			continue
		}

		// skip ignored tags
		if fieldJsonTag == "-" {
			continue
		}

		// skip for function and channel types
		if fieldType == "func" || fieldType == "chan" {
			continue
		}

		// if item type is array, create Definition for array element type
		switch fieldType {
		case "array":
			if field.Type.Elem().Kind() == reflect.Struct {
				// TODO make a constructor function for swaggerdefinition.DefinitionProperties and create tests for all types to ensure it's extracting the tags correctly
				hash := utils.GenerateHash(utils.ConvertToString(reflect.New(field.Type.Elem()).Elem().Interface()))
				properties[fieldJsonTag] = definition.DefinitionProperties{
					Example: exampleTag,
					Type:    fieldType,
					Items: &definition.DefinitionPropertiesItems{
						Ref: g.createRef("%s%s_%s", prefix, field.Type.Elem().String(), hash),
					},
				}
				if _struct == field.Type.Elem() {
					continue // prevent recursion
				}
				g.CreateDefinition(reflect.New(field.Type.Elem()).Elem().Interface(), prefix)
			} else {
				properties[fieldJsonTag] = definition.DefinitionProperties{
					Example: exampleTag,
					Type:    fieldType,
					Items: &definition.DefinitionPropertiesItems{
						Type: getType(field.Type.Elem().Kind().String()),
					},
				}
			}

		case "struct":
			if field.Type.String() == "time.Time" {
				properties[fieldJsonTag] = g.timeProperty(field)
			} else if field.Type.String() == "time.Duration" {
				properties[fieldJsonTag] = g.durationProperty(field)
			} else {
				properties[fieldJsonTag] = g.refProperty(field)
				g.CreateDefinition(reflect.New(field.Type).Elem().Interface(), prefix)
			}

		case "ptr":
			if field.Type.Elem() == _struct { // prevent recursion
				properties[fieldJsonTag] = definition.DefinitionProperties{
					Example: fmt.Sprintf("Recursive Type: %s", field.Type.Elem().String()),
				}
				continue
			}
			if field.Type.Elem().Kind() == reflect.Struct {
				if field.Type.Elem().String() == "time.Time" {
					properties[fieldJsonTag] = g.timeProperty(field)
				} else if field.Type.String() == "time.Duration" {
					properties[fieldJsonTag] = g.durationProperty(field)
				} else {
					properties[fieldJsonTag] = g.refProperty(field)
					g.CreateDefinition(reflect.New(field.Type.Elem()).Elem().Interface(), prefix)
				}
			} else {
				properties[fieldJsonTag] = definition.DefinitionProperties{
					Example: exampleTag,
					Type:    getType(field.Type.Elem().Kind().String()),
				}
			}

		case "map":
			name := fmt.Sprintf("%s%s.%s", prefix, _struct.String(), fieldJsonTag)
			mapKeyType := field.Type.Key()
			mapValueType := field.Type.Elem()
			if mapValueType.Kind() == reflect.Ptr {
				mapValueType = mapValueType.Elem()
			}
			properties[fieldJsonTag] = definition.DefinitionProperties{
				Ref: g.createRef("%s", name),
			}
			if mapValueType.Kind() == reflect.Struct {
				hash := utils.GenerateHash(utils.ConvertToString(reflect.New(mapValueType).Elem().Interface()))
				g.Definitions[name] = definition.Definition{
					Type: "object",
					Properties: map[string]definition.DefinitionProperties{
						getType(mapKeyType.String()): {
							Ref: g.createRef("%s%s_%s", prefix, mapValueType.String(), hash),
						},
					},
				}
				g.CreateDefinition(reflect.New(mapValueType).Elem().Interface(), prefix)
			} else {
				g.Definitions[name] = definition.Definition{
					Type: "object",
					Properties: map[string]definition.DefinitionProperties{
						getType(mapKeyType.String()): {
							Example: exampleTag,
							Type:    getType(mapValueType.String()),
						},
					},
				}
			}

		case "interface":
			if reflect.TypeOf(data) == reflect.TypeOf(response.CustomResponseType{}) {
				return properties
			}

			if reflect.ValueOf(data).Kind() == reflect.Struct {
				val := reflect.ValueOf(data).FieldByName(field.Name).Interface()
				hash := utils.GenerateHash(utils.ConvertToString(val))

				if val != nil {
					properties[fieldJsonTag] = definition.DefinitionProperties{
						Example: exampleTag,
						Ref:     g.createRef("%s%s_%s", prefix, reflect.TypeOf(val), hash),
					}
					g.CreateDefinition(val, prefix)
				} else {
					properties[fieldJsonTag] = definition.DefinitionProperties{
						Example: exampleTag,
						Type:    "object",
					}
				}
			} else {
				properties[fieldJsonTag] = definition.DefinitionProperties{
					Example: exampleTag,
					Type:    "object",
				}
			}

		default:
			properties[fieldJsonTag] = g.defaultProperty(field)
		}
	}

	return properties
}

func (g DefinitionGenerator) createMapDefinitions(v reflect.Value, prefix string) map[string]definition.DefinitionProperties {
	properties := make(map[string]definition.DefinitionProperties)

	g.walkMap(v, properties, prefix)

	return properties
}

func (g DefinitionGenerator) timeProperty(field reflect.StructField) definition.DefinitionProperties {
	return definition.DefinitionProperties{
		Example: getExampleTag(field),
		Type:    "string",
		Format:  "date-time",
	}
}

func (g DefinitionGenerator) durationProperty(field reflect.StructField) definition.DefinitionProperties {
	return definition.DefinitionProperties{
		Example: getExampleTag(field),
		Type:    "integer",
	}
}

func (g DefinitionGenerator) createRef(pattern string, a ...any) string {
	return fmt.Sprintf("#/definitions/%s", fmt.Sprintf(pattern, a...))
}

func (g DefinitionGenerator) refProperty(field reflect.StructField) definition.DefinitionProperties {
	return definition.DefinitionProperties{
		Example: getExampleTag(field),
		Ref:     g.createRef("%s", field.Type.Elem().String()),
	}
}

func (g DefinitionGenerator) defaultProperty(field reflect.StructField) definition.DefinitionProperties {
	return definition.DefinitionProperties{
		Example: getExampleTag(field),
		Type:    getType(field.Type.Kind().String()),
	}
}

func (g DefinitionGenerator) getSwagnoTag(field reflect.StructField) definition.DefinitionProperties {
	return definition.DefinitionProperties{
		Example: getExampleTag(field),
		Type:    field.Tag.Get("swagno"),
	}
}

func (g DefinitionGenerator) walkMap(v reflect.Value, m map[string]definition.DefinitionProperties, prefix string) reflect.Value {
	if v.Kind() != reflect.Map {
		return v
	}

	if len(v.MapKeys()) == 0 {
		v = createExampleMap(v.Interface())
		if len(v.MapKeys()) == 0 {
			v = reflect.ValueOf(map[string]interface{}{"string": nil})

		}
	}

	for _, k := range v.MapKeys() {
		val := g.walkMap(v.MapIndex(k), m, prefix)
		if val.Type().Kind() == reflect.Pointer {
			val = reflect.New(val.Type().Elem()).Elem()
		}

		switch val.Type().Kind() {
		case reflect.Struct:
			hash := utils.GenerateHash(utils.ConvertToString(val.Interface()))
			m[k.String()] = definition.DefinitionProperties{
				Ref: fmt.Sprintf("#/definitions/%s%s_%s", prefix, val.Type().String(), hash),
			}
			g.CreateDefinition(val.Interface(), prefix)
		case reflect.Slice:
			hash := utils.GenerateHash(utils.ConvertToString(val.Interface()))
			elem := val.Type().Elem()

			m[k.String()] = definition.DefinitionProperties{
				Ref: fmt.Sprintf("#/definitions/%s%s_%s", prefix, elem.String(), hash),
			}
			g.CreateDefinition(val.Interface(), prefix)

		case reflect.Interface:
			_type := "object"
			if val.Elem().IsValid() {
				_type = getType(val.Type().String())
			}
			m[k.String()] = definition.DefinitionProperties{
				Type: _type,
			}

		default:
			valueType := val.Type()
			if valueType.Kind() == reflect.Interface {
				valueType = val.Elem().Type()
			}

			m[k.String()] = definition.DefinitionProperties{
				Type: getType(valueType.String()),
			}
		}
	}

	return v
}

func createExampleMap(emptyMap interface{}) reflect.Value {
	mapType := reflect.TypeOf(emptyMap)
	if mapType.Kind() != reflect.Map {
		panic("Input must be a map")
	}

	result := reflect.MakeMap(mapType)
	defaultValue := reflect.Zero(mapType.Elem()).Interface()
	result.SetMapIndex(reflect.ValueOf("string"), reflect.ValueOf(defaultValue))
	return result
}
