package jsonschema

import (
	"reflect"
	"strconv"
)

const (
	tagName     = "name"
	tagNameJson = "json"
	tagTitle    = "title"
	tagRequired = "required"
	tagIgnore   = "ignore"
	// string
	tagStringMinLength = "minLength"
	tagStringMaxLength = "maxLength"
	tagStringFormat    = "format"
	// number
	tagNumberMultipleOf       = "multipleOf"
	tagNumberMinimum          = "minimum"
	tagNumberMaximum          = "maximum"
	tagNumberExclusiveMaximum = "exclusiveMaximum"
	tagNumberExclusiveMinimum = "exclusiveMinimum"
	// array
	tagArrayMinItems    = "minItems"
	tagArrayMaxItems    = "maxItems"
	tagArrayUniqueItems = "uniqueItems"
)

type tags struct {
	name     string
	title    string
	required bool
	ignored  bool
	// string specific
	minLength int
	maxLength int
	format    string
	// number specific
	multipleOf       int
	minimum          int
	maximum          int
	exclusiveMaximum bool
	exclusiveMinimum bool
	// array specific
	minItems    int
	maxItems    int
	uniqueItems bool
}

func parseTags(tag reflect.StructTag) tags {
	t := tags{}

	var ok bool
	if t.name, ok = tag.Lookup(tagName); !ok {
		t.name = tag.Get(tagNameJson)
	}

	t.title = tag.Get(tagTitle)
	t.ignored, _ = strconv.ParseBool(tag.Get(tagIgnore))
	t.required, _ = strconv.ParseBool(tag.Get(tagRequired))

	// string specific
	t.minLength, _ = strconv.Atoi(tag.Get(tagStringMinLength))
	t.maxLength, _ = strconv.Atoi(tag.Get(tagStringMaxLength))
	t.format = tag.Get(tagStringFormat)

	// number specific
	t.multipleOf, _ = strconv.Atoi(tag.Get(tagNumberMultipleOf))
	t.minimum, _ = strconv.Atoi(tag.Get(tagNumberMinimum))
	t.maximum, _ = strconv.Atoi(tag.Get(tagNumberMaximum))
	t.exclusiveMinimum, _ = strconv.ParseBool(tag.Get(tagNumberExclusiveMinimum))
	t.exclusiveMaximum, _ = strconv.ParseBool(tag.Get(tagNumberExclusiveMaximum))

	// array specific
	t.minItems, _ = strconv.Atoi(tag.Get(tagArrayMinItems))
	t.maxItems, _ = strconv.Atoi(tag.Get(tagArrayMaxItems))
	t.uniqueItems, _ = strconv.ParseBool(tag.Get(tagArrayUniqueItems))

	return t
}

func applyValidation(dst *Type, t tags) {
	switch dst.Type {
	case tTypeString:
		dst.MinLength = t.minLength
		dst.MaxLength = t.maxLength
		dst.Format = t.format
	case tTypeNumber:
		dst.MultipleOf = t.multipleOf
		dst.Minimum = t.minimum
		dst.Maximum = t.maximum
		dst.ExclusiveMinimum = t.exclusiveMinimum
		dst.ExclusiveMaximum = t.exclusiveMaximum
	case tTypeArray:
		dst.MinItems = t.minItems
		dst.MaxItems = t.maxItems
		dst.UniqueItems = t.uniqueItems
	}
}