// Package jsonschema uses reflection to generate JSON Schemas from Go types [1].
//
// If json tags are present on struct fields, they will be used to infer
// property names and if a property is required (omitempty is present).
//
// [1] http://json-schema.org/latest/json-schema-validation.html
package jsonschema

import (
	"reflect"
)

const (
	tTypeObject  = "object"
	tTypeString  = "string"
	tTypeInteger = "integer"
	tTypeNumber  = "number"
	tTypeBoolean = "boolean"
	tTypeArray   = "array"
)

// Reflect reflects to Schema from a value.
func Reflect(v interface{}) *Schema {
	valueOf := reflect.ValueOf(v)
	typeOf := reflect.TypeOf(v)

	valueOf = reflect.Indirect(valueOf)

	definitions := Definitions{}

	root := reflectType(definitions, typeOf, valueOf, true)
	root.Version = Version

	return &Schema{Type: root, Definitions: definitions}
}

func reflectType(definitions Definitions, t reflect.Type, v reflect.Value, root bool) *Type {
	if v.Kind() == reflect.Ptr {
		v = v.Elem() // deref ptr

		if !v.IsValid() {
			v = reflect.Zero(t.Elem()) // create zero value
		}
	}

	if v.Kind() == reflect.Interface {
		v = reflect.Indirect(v.Elem())
	}

	switch t {
	case typeTime:
		return reflectTime(definitions, v)
	case typeIP:
		return reflectIP(definitions, v)
	case typeURI:
		return reflectURI(definitions, v)
	}

	switch true {
	case t.Implements(typePBEnum):
		return reflectPBEnum(definitions, v)

	case t.Implements(typeOneOf):
		return reflectOneOf(definitions, v)

	case t.Implements(typeAnyOf):
		return reflectAnyOf(definitions, v)

	case t.Implements(typeAllOf):
		return reflectAllOf(definitions, v)

	case t.Implements(typeEnum):
		return reflectEnum(definitions, v)
	}

	switch v.Kind() {
	case reflect.Struct:
		currentType := reflectStruct(definitions, v)
		if root {
			return currentType
		}

		definitions[v.Type().Name()] = currentType

		return newReference(v.Type().Name())

	case reflect.Slice:
		return reflectSlice(definitions, v)

	case reflect.Map:
		return reflectMap(definitions, v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		return reflectInteger(definitions, v)

	case reflect.Float32, reflect.Float64:
		return reflectNumber(definitions, v)

	case reflect.Bool:
		return reflectBool(definitions, v)

	case reflect.String:
		return reflectString(definitions, v)
	}

	return reflectInterface(definitions, t, v)
}

func reflectStruct(definitions Definitions, v reflect.Value) *Type {
	var currentType = newType(tTypeObject)

	for i := 0; i < v.NumField(); i++ {
		structField := v.Type().Field(i)
		structValue := v.Field(i)

		// unexported field
		if structField.PkgPath != "" {
			continue
		}

		// embedded field
		if structField.Anonymous {
			typ := reflectStruct(definitions, structValue)
			for def, info := range typ.Definitions {
				definitions[def] = info
			}

			for def, info := range typ.Properties {
				currentType.Properties[def] = info
			}
		}

		tags := parseTags(structField.Tag)
		if tags.name == "" || tags.ignored {
			continue
		}

		fieldType := reflectType(definitions, structField.Type, structValue, false)
		if fieldType == nil {
			continue
		}

		applyInfo(fieldType, tags)
		applyValidation(fieldType, tags)

		currentType.Properties[tags.name] = fieldType
	}

	return currentType
}
