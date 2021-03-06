package jsonschema

import (
	"net"
	"net/url"
	"reflect"
	"time"
)

// Available Go defined types for JSON Schema Validation.
// RFC draft-wright-json-schema-validation-00, section 7.3
// custom types
var (
	typeTime      = reflect.TypeOf(time.Time{}) // date-time RFC section 7.3.1
	typeIP        = reflect.TypeOf(net.IP{})    // ipv4 and ipv6 RFC section 7.3.4, 7.3.5
	typeURI       = reflect.TypeOf(url.URL{})   // uri RFC section 7.3.6
	typeByteSlice = reflect.TypeOf([]byte(nil))
	typePBEnum    = reflect.TypeOf((*protoEnum)(nil)).Elem()
	typeEnum      = reflect.TypeOf((*enumType)(nil)).Elem()
	typeOneOf     = reflect.TypeOf((*implicitOneOf)(nil)).Elem()
	typeAnyOf     = reflect.TypeOf((*implicitAnyOf)(nil)).Elem()
	typeAllOf     = reflect.TypeOf((*implicitAllOf)(nil)).Elem()
)

// Go code generated from protobuf enum types should fulfil this interface.
type protoEnum interface {
	EnumDescriptor() ([]byte, []int)
}

type implicitOneOf interface {
	OneOf() []interface{}
}

type implicitAnyOf interface {
	AnyOf() []interface{}
}

type implicitAllOf interface {
	AllOf() []interface{}
}

type enumType interface {
	Enum() []interface{}
}

func reflectTime(definition Definitions, v reflect.Value) *Type {
	t := Type{
		Type:   tTypeString,
		Format: "date-time",
	}

	return &t
}

// ipv4 RFC section 7.3.4
func reflectIP(definition Definitions, v reflect.Value) *Type {
	return &Type{
		Type:   tTypeString,
		Format: "ipv4",
	}
}

// uri RFC section 7.3.6
func reflectURI(definition Definitions, v reflect.Value) *Type {
	return &Type{
		Type:   tTypeString,
		Format: "uri",
	}
}

func reflectPBEnum(definition Definitions, v reflect.Value) *Type {
	return &Type{OneOf: []*Type{
		{Type: tTypeString},
		{Type: tTypeInteger},
	}}
}

func reflectEnum(definition Definitions, v reflect.Value) *Type {
	variants := v.Interface().(enumType).Enum()

	variantValueOf := reflect.ValueOf(variants[0])
	variantTypeOf := reflect.TypeOf(variants[0])

	vType := reflectType(definition, variantTypeOf, variantValueOf, false)

	typ := &Type{
		Type: vType.Type,
		Enum: variants,
	}

	handleDefaultValue(typ, v)

	return typ
}

func reflectOneOf(definition Definitions, v reflect.Value) *Type {
	variants := v.Interface().(implicitOneOf).OneOf()

	oneOf := make([]*Type, len(variants))

	for idx, variant := range variants {
		oneOf[idx] = reflectType(definition,
			reflect.TypeOf(variant),
			reflect.ValueOf(variant), false)
	}

	typ := &Type{
		Type:  oneOf[0].Type,
		OneOf: oneOf,
	}

	handleDefaultValue(typ, v)

	return typ
}

func reflectAnyOf(definition Definitions, v reflect.Value) *Type {
	variants := v.Interface().(implicitAnyOf).AnyOf()

	anyOf := make([]*Type, len(variants))

	for idx, variant := range variants {
		anyOf[idx] = reflectType(definition,
			reflect.TypeOf(variant),
			reflect.ValueOf(variant), false)
	}

	typ := &Type{
		Type:  anyOf[0].Type,
		AnyOf: anyOf,
	}

	handleDefaultValue(typ, v)

	return typ
}

func reflectAllOf(definition Definitions, v reflect.Value) *Type {
	variants := v.Interface().(implicitAllOf).AllOf()

	allOf := make([]*Type, len(variants))

	for idx, variant := range variants {
		allOf[idx] = reflectType(definition,
			reflect.TypeOf(variant),
			reflect.ValueOf(variant), false)
	}

	typ := &Type{
		Type:  allOf[0].Type,
		AllOf: allOf,
	}

	handleDefaultValue(typ, v)

	return typ
}

func getSliceValue(v reflect.Value) reflect.Value {
	if v.Len() > 0 {
		return v.Index(0)
	}

	return reflect.New(v.Type().Elem())
}

func reflectSlice(definition Definitions, v reflect.Value) *Type {
	returnType := newType("")

	if v.Type().Kind() == reflect.Array {
		returnType.MinItems = v.Type().Len()
		returnType.MaxItems = v.Type().Len()
	}

	elemValue := reflect.New(v.Type().Elem())

	switch v.Type() {
	case typeByteSlice:
		returnType.Type = tTypeString
		returnType.Media = &Type{
			BinaryEncoding: "base64",
		}
	default:
		returnType.Type = "array"
		returnType.Items = reflectType(definition, elemValue.Type(), elemValue, false)
	}

	defaults := make([]interface{}, 0)
	for i := 0; i < v.Len(); i++ {
		defaults = append(defaults, v.Index(i).Interface())
	}

	return returnType
}

func reflectMap(definitions Definitions, v reflect.Value) *Type {
	val := v.Type().Elem()

	rt := &Type{
		Type: tTypeObject,
		PatternProperties: map[string]*Type{
			".*": reflectType(definitions, val, reflect.New(val), false),
		},
	}
	delete(rt.PatternProperties, "additionalProperties")

	return rt
}

func reflectInteger(definitions Definitions, v reflect.Value) *Type {
	typ := &Type{
		Type: tTypeInteger,
	}

	handleDefaultValue(typ, v)

	return typ
}

func reflectNumber(definitions Definitions, v reflect.Value) *Type {
	typ := &Type{
		Type: tTypeNumber,
	}

	handleDefaultValue(typ, v)

	return typ
}

func reflectBool(definitions Definitions, v reflect.Value) *Type {
	typ := &Type{
		Type: tTypeBoolean,
	}

	handleDefaultValue(typ, v)

	return typ
}

func reflectString(definitions Definitions, v reflect.Value) *Type {
	typ := &Type{
		Type: tTypeString,
	}

	handleDefaultValue(typ, v)

	return typ
}

func reflectInterface(definitions Definitions, t reflect.Type, v reflect.Value) *Type {
	typ := &Type{
		Type:                 tTypeObject,
		AdditionalProperties: []byte("true"),
	}

	handleDefaultValue(typ, v)

	return typ
}

func handleDefaultValue(dst *Type, v reflect.Value) {
	if v.IsValid() {
		dst.Default = v.Interface()
	}
}
