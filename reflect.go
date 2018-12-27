// Package jsonschema uses reflection to generate JSON Schemas from Go types [1].
//
// If json tags are present on struct fields, they will be used to infer
// property names and if a property is required (omitempty is present).
//
// [1] http://json-schema.org/latest/json-schema-validation.html
package jsonschema

import (
	"net"
	"net/url"
	"reflect"
	"time"
)

const (
	tTypeString  = "string"
	tTypeInteger = "integer"
	tTypeObject  = "object"
	tTypeNumber  = "number"
	tTypeBoolean = "boolean"
	tTypeArray   = "array"
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

type implicit interface {
	ImplicitType() interface{}
}

type enumType interface {
	Enum() []interface{}
}

var (
	typePBEnum   = reflect.TypeOf((*protoEnum)(nil)).Elem()
	typeEnum     = reflect.TypeOf((*enumType)(nil)).Elem()
	typeOneOf    = reflect.TypeOf((*implicitOneOf)(nil)).Elem()
	typeAnyOf    = reflect.TypeOf((*implicitAnyOf)(nil)).Elem()
	typeAllOf    = reflect.TypeOf((*implicitAllOf)(nil)).Elem()
	typeImplicit = reflect.TypeOf((*implicit)(nil)).Elem()
)

// Available Go defined types for JSON Schema Validation.
// RFC draft-wright-json-schema-validation-00, section 7.3
// custom types
var (
	typeTime      = reflect.TypeOf(time.Time{}) // date-time RFC section 7.3.1
	typeIP        = reflect.TypeOf(net.IP{})    // ipv4 and ipv6 RFC section 7.3.4, 7.3.5
	typeURI       = reflect.TypeOf(url.URL{})   // uri RFC section 7.3.6
	typeByteSlice = reflect.TypeOf([]byte(nil))
)

// A Reflector reflects values into a Schema.
type Reflector struct {
	// AllowAdditionalProperties will cause the Reflector to generate a schema
	// with additionalProperties to 'true' for all struct types. This means
	// the presence of additional keys in JSON objects will not cause validation
	// to fail. Note said additional keys will simply be dropped when the
	// validated JSON is unmarshaled.
	AllowAdditionalProperties bool

	// RequiredFromJSONSchemaTags will cause the Reflector to generate a schema
	// that requires any key tagged with `jsonschema:required`, overriding the
	// default of requiring any key *not* tagged with `json:,omitempty`.
	RequiredFromJSONSchemaTags bool

	// ExpandedStruct will cause the toplevel definitions of the schema not
	// be referenced itself to a definition.
	ExpandedStruct bool
}

// Reflect reflects to Schema from a value.
func (r *Reflector) Reflect(v interface{}) *Schema {
	typeOf := reflect.TypeOf(v)
	valueOf := reflect.ValueOf(v)

	return r.ReflectFromType(typeOf, valueOf)
}

// ReflectFromType generates root schema
func (r *Reflector) ReflectFromType(t reflect.Type, v reflect.Value) *Schema {
	definitions := Definitions{}
	if r.ExpandedStruct {
		st := &Type{
			Version:              Version,
			Type:                 tTypeObject,
			Properties:           map[string]*Type{},
			AdditionalProperties: []byte("false"),
		}
		if r.AllowAdditionalProperties {
			st.AdditionalProperties = []byte("true")
		}
		reflectStructFields(st, definitions, t, v)
		reflectStruct(definitions, t, v)
		delete(definitions, t.Name())
		return &Schema{Type: st, Definitions: definitions}
	}

	s := &Schema{
		Type:        reflectTypeToSchema(definitions, t, v),
		Definitions: definitions,
	}
	return s
}

func reflectPBEnum(def Definitions, t reflect.Type, v reflect.Value) *Type {
	return &Type{OneOf: []*Type{
		{Type: tTypeString},
		{Type: tTypeInteger},
	}}
}

func reflectEnum(def Definitions, t reflect.Type, v reflect.Value) *Type {
	variants := v.Interface().(enumType).Enum()

	vType := reflectTypeToSchema(def, reflect.TypeOf(variants[0]), reflect.ValueOf(variants[0]))

	ret := &Type{
		Type: vType.Type,
		Enum: variants,
	}

	return ret
}

func reflectOneOf(def Definitions, t reflect.Type, v reflect.Value) *Type {
	variants := reflect.Zero(t).
		Interface().(implicitOneOf).OneOf()

	oneOf := make([]*Type, len(variants))

	for idx, variant := range variants {
		oneOf[idx] = reflectTypeToSchema(def, reflect.TypeOf(variant), v)
	}

	return &Type{OneOf: oneOf}
}

func reflectAnyOf(def Definitions, t reflect.Type, v reflect.Value) *Type {
	variants := reflect.Zero(t).
		Interface().(implicitAnyOf).AnyOf()

	anyOf := make([]*Type, len(variants))

	for idx, variant := range variants {
		anyOf[idx] = reflectTypeToSchema(def, reflect.TypeOf(variant), v)
	}

	return &Type{AnyOf: anyOf}
}

func reflectAllOf(def Definitions, t reflect.Type, v reflect.Value) *Type {
	variants := reflect.Zero(t).
		Interface().(implicitAllOf).AllOf()

	allOf := make([]*Type, len(variants))

	for idx, variant := range variants {
		allOf[idx] = reflectTypeToSchema(def, reflect.TypeOf(variant), v)
	}

	return &Type{AllOf: allOf}
}

func reflectImplicit(def Definitions, t reflect.Type, v reflect.Value) *Type {
	typ := reflect.Zero(t).
		Interface().(implicit).ImplicitType()

	return reflectTypeToSchema(def, reflect.TypeOf(typ), v)
}

func reflectTime(def Definitions, t reflect.Type, v reflect.Value) *Type {
	return &Type{Type: tTypeString, Format: "date-time"}
}

func reflectIP(def Definitions, t reflect.Type, v reflect.Value) *Type {
	return &Type{Type: tTypeString, Format: "ipv4"} // ipv4 RFC section 7.3.4
}

func reflectURI(def Definitions, t reflect.Type, v reflect.Value) *Type {
	return &Type{Type: tTypeString, Format: "uri"} // uri RFC section 7.3.6
}

func reflectSlice(def Definitions, t reflect.Type, v reflect.Value) *Type {
	returnType := &Type{}
	if t.Kind() == reflect.Array {
		returnType.MinItems = t.Len()
		returnType.MaxItems = t.Len()
	}

	switch t {
	case typeByteSlice:
		returnType.Type = tTypeString
		returnType.Media = &Type{
			BinaryEncoding: "base64",
		}
	default:
		returnType.Type = "array"
		returnType.Items = reflectTypeToSchema(def, t.Elem(), v)
	}

	return returnType
}

func reflectMap(def Definitions, t reflect.Type, v reflect.Value) *Type {
	rt := &Type{
		Type: "object",
		PatternProperties: map[string]*Type{
			".*": reflectTypeToSchema(def, t.Elem(), v),
		},
	}
	delete(rt.PatternProperties, "additionalProperties")

	return rt
}

func reflectInterface(def Definitions, t reflect.Type, v reflect.Value) *Type {
	return &Type{
		Type:                 "object",
		AdditionalProperties: []byte("true"),
	}
}

func reflectStruct(def Definitions, t reflect.Type, v reflect.Value) *Type {
	switch t {
	case typeTime: // date-time RFC section 7.3.1
		return reflectTime(def, t, v)
	case typeURI: // uri RFC section 7.3.6
		return reflectURI(def, t, v)
	case typeIP:
		return reflectIP(def, t, v)
	}

	st := &Type{
		Type:                 tTypeObject,
		Properties:           map[string]*Type{},
		AdditionalProperties: []byte("false"),
	}
	if true {
		st.AdditionalProperties = []byte("true")
	}
	def[t.Name()] = st
	reflectStructFields(st, def, t, v)

	return &Type{
		Version: Version,
		Ref:     "#/definitions/" + t.Name(),
	}
}

func reflectTypeToSchema(definitions Definitions, t reflect.Type, v reflect.Value) *Type {
	// Already added to definitions?
	if _, ok := definitions[t.Name()]; ok {
		return &Type{Ref: "#/definitions/" + t.Name()}
	}

	// specific interfaces
	switch true {
	case t.Implements(typePBEnum):
		return reflectPBEnum(definitions, t, v)

	case t.Implements(typeOneOf):
		return reflectOneOf(definitions, t, v)

	case t.Implements(typeAnyOf):
		return reflectAnyOf(definitions, t, v)

	case t.Implements(typeAllOf):
		return reflectAllOf(definitions, t, v)

	case t.Implements(typeImplicit):
		return reflectImplicit(definitions, t, v)

	case t.Implements(typeEnum):
		return reflectEnum(definitions, t, v)
	}

	switch t.Kind() {
	case reflect.Struct:
		return reflectStruct(definitions, t, v)

	case reflect.Map:
		return reflectMap(definitions, t, v)

	case reflect.Slice, reflect.Array:
		return reflectSlice(definitions, t, v)

	case reflect.Interface:
		return reflectInterface(definitions, t, v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Type{Type: tTypeInteger}

	case reflect.Float32, reflect.Float64:
		return &Type{Type: tTypeNumber}

	case reflect.Bool:
		return &Type{Type: tTypeBoolean}

	case reflect.String:
		return &Type{Type: tTypeString}

	case reflect.Ptr:
		return reflectTypeToSchema(definitions, t.Elem(), v)
	}
	panic("unsupported type " + t.String())
}

func reflectStructFields(st *Type, definitions Definitions, t reflect.Type, v reflect.Value) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// anonymous and exported type should be processed recursively
		// current type should inherit properties of anonymous one
		if f.Anonymous && f.PkgPath == "" {
			reflectStructFields(st, definitions, f.Type, v)
			continue
		}

		tags := parseTags(f.Tag)
		if tags.name == "" {
			continue
		}

		property := reflectTypeToSchema(definitions, f.Type, v.Field(i))
		property.Title = tags.title
		applyValidation(property, tags)

		st.Properties[tags.name] = property
		if tags.required {
			st.Required = append(st.Required, tags.name)
		}
	}
}
