package jsonschema

import (
	"net"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type enumImpl struct{}

func (enumImpl) Enum() []interface{} {
	return []interface{}{"1", "2", "3"}
}

type implicitOneOfImpl struct{}

func (implicitOneOfImpl) OneOf() []interface{} {
	return []interface{}{"1", "2", "3"}
}

type implicitAnyOfImpl struct{}

func (implicitAnyOfImpl) AnyOf() []interface{} {
	return []interface{}{"1", "2", "3"}
}

type implicitAllOfImpl struct{}

func (implicitAllOfImpl) AllOf() []interface{} {
	return []interface{}{"1", "2", "3"}
}

type GrandfatherType struct {
	FamilyName string `json:"family_name" jsonschema:"required"`
}

type SomeBaseType struct {
	SomeBaseProperty int `json:"some_base_property"`
	// The jsonschema required tag is nonsensical for private and ignored properties.
	// Their presence here tests that the fields *will not* be required in the output
	// schema, even if they are tagged required.
	somePrivateBaseProperty   string          `json:"i_am_private" jsonschema:"required"`
	SomeIgnoredBaseProperty   string          `json:"-" jsonschema:"required"`
	SomeSchemaIgnoredProperty string          `jsonschema:"-,required"`
	Grandfather               GrandfatherType `json:"grand"`

	SomeUntaggedBaseProperty           bool `jsonschema:"required"`
	someUnexportedUntaggedBaseProperty bool
}

type nonExported struct {
	PublicNonExported  int
	privateNonExported int
}

type ProtoEnum int32

func (ProtoEnum) EnumDescriptor() ([]byte, []int) { return []byte(nil), []int{0} }

const (
	Unset ProtoEnum = iota
	Great
)

type TestUser struct {
	SomeBaseType
	nonExported

	ID      int                    `json:"id" jsonschema:"required"`
	Name    string                 `json:"name" jsonschema:"required,minLength=1,maxLength=20"`
	Friends []int                  `json:"friends,omitempty"`
	Tags    map[string]interface{} `json:"tags,omitempty"`

	TestFlag       bool
	IgnoredCounter int `json:"-"`

	// Tests for RFC draft-wright-json-schema-validation-00, section 7.3
	BirthDate time.Time `json:"birth_date,omitempty"`
	Website   url.URL   `json:"website,omitempty"`
	IPAddress net.IP    `json:"network_address,omitempty"`

	// Tests for RFC draft-wright-json-schema-hyperschema-00, section 4
	Photo []byte `json:"photo,omitempty" jsonschema:"required"`

	// Tests for jsonpb enum support
	Feeling ProtoEnum         `json:"feeling,omitempty"`
	Age     int               `json:"age" jsonschema:"minimum=18,maximum=120,exclusiveMaximum=true,exclusiveMinimum=true"`
	Email   string            `json:"email" jsonschema:"format=email"`
	OneOf   implicitOneOfImpl `json:"oneOf"`
	AllOf   implicitAnyOfImpl `json:"allOf"`
	AnyOf   implicitAnyOfImpl `json:"anyOf"`
	Enum    enumImpl          `json:"enum"`
}

type String string

type ColorPicker struct {
	String
}

type SomeStruct struct {
	ColorPicker ColorPicker `json:"colorPicker"`
	TextArea    TextArea    `json:"textArea"`
}

type TextArea struct {
	String
}

func TestEmbeddedTypes(t *testing.T) {
	schema := Reflect(SomeStruct{})

	assert.NotNil(t, schema.Definitions["ColorPicker"])
	assert.Equal(t, tTypeString, schema.Definitions["ColorPicker"].Type)
	assert.Equal(t, String(""), schema.Definitions["ColorPicker"].Default)

	assert.NotNil(t, schema.Definitions["TextArea"])
	assert.Equal(t, tTypeString, schema.Definitions["TextArea"].Type)
	assert.Equal(t, String(""), schema.Definitions["TextArea"].Default)
}

func TestReflect(t *testing.T) {
	t.Run("ReflectStruct_returns_CorrectType", func(t *testing.T) {
		a := assert.New(t)
		r := require.New(t)

		d := Definitions{}
		_ = d

		tu := TestUser{
			SomeBaseType: SomeBaseType{},
			ID:           666,
			Name:         "some name",
			Friends:      []int{1, 2, 3, 4, 5, 6},
			Tags: map[string]interface{}{
				"tag1": "value1",
				"tag2": "value2",
			},
			TestFlag:       true,
			IgnoredCounter: 666,
			BirthDate:      time.Now(),
			Website:        url.URL{Scheme: "https", Host: "google.com"},
			IPAddress:      net.IPv4(127, 0, 0, 1),
			Photo:          []byte{},
			Feeling:        Great,
			Age:            666,
			Email:          "some@email.com",
		}

		schema := Reflect(tu)

		r.Contains(schema.Definitions, "GrandfatherType")
		grandfatherType := schema.Definitions["GrandfatherType"]

		r.Contains(grandfatherType.Properties, "family_name")
		a.Equal(grandfatherType.Properties["family_name"].Type, tTypeString)
		a.Equal(tu.Grandfather.FamilyName, grandfatherType.Properties["family_name"].Default)

		r.Contains(schema.Properties, "id")
		idProperty := schema.Properties["id"]
		a.Equal(tTypeInteger, idProperty.Type)
		a.Equal(tu.ID, idProperty.Default)

		r.Contains(schema.Properties, "name")
		nameProperty := schema.Properties["name"]
		a.Equal(tTypeString, nameProperty.Type)
		a.Equal(tu.Name, nameProperty.Default)

		r.Contains(schema.Properties, "friends")
		friendsProperty := schema.Properties["friends"]
		a.Equal(tTypeArray, friendsProperty.Type)

		r.Contains(schema.Properties, "tags")
		tagsProperty := schema.Properties["tags"]
		a.Equal(tTypeObject, tagsProperty.Type)

		r.Contains(schema.Properties, "birth_date")
		birthDateProperty := schema.Properties["birth_date"]
		a.Equal(tTypeString, birthDateProperty.Type)
		//a.Equal("date-time", birthDateProperty.Format)

		r.Contains(schema.Properties, "website")
		websiteProperty := schema.Properties["website"]
		a.Equal(tTypeString, websiteProperty.Type)
		//a.Equal("uri", websiteProperty.Format)

		r.Contains(schema.Properties, "network_address")
		networkAddressProperty := schema.Properties["network_address"]
		a.Equal(tTypeString, networkAddressProperty.Type)
		//a.Equal("ipv4", networkAddressProperty.Format)

		r.Contains(schema.Properties, "photo")
		photoProperty := schema.Properties["photo"]
		a.Equal(tTypeString, photoProperty.Type)
		//a.Equal("base64", photoProperty.Media.BinaryEncoding)

		// TODO: implement
		//r.Contains(schema.Properties, "feeling")
		//feelingProperty := schema.Properties["feeling"]
		//a.Equal(tTypeArray, feelingProperty.Type)
		//a.Equal("base64", feelingProperty.Format)

		r.Contains(schema.Properties, "age")
		ageProperty := schema.Properties["age"]
		a.Equal(tTypeInteger, ageProperty.Type)
		a.Equal(tu.ID, ageProperty.Default)

		r.Contains(schema.Properties, "email")
		emailProperty := schema.Properties["email"]
		a.Equal(tTypeString, emailProperty.Type)

		r.Contains(schema.Properties, "oneOf")
		oneOfProperty := schema.Properties["oneOf"]
		a.Equal(tTypeString, oneOfProperty.Type)

		r.Contains(schema.Properties, "allOf")
		allOffProperty := schema.Properties["allOf"]
		a.Equal(tTypeString, allOffProperty.Type)

		r.Contains(schema.Properties, "anyOf")
		anyOfProperty := schema.Properties["anyOf"]
		a.Equal(tTypeString, anyOfProperty.Type)

		r.Contains(schema.Properties, "enum")
		enumProperty := schema.Properties["enum"]
		a.Equal(tTypeString, enumProperty.Type)
	})

	t.Run("ReflectTime_returns_ValidType", func(t *testing.T) {
		d := Definitions{}
		v := reflect.ValueOf(time.Now())

		typ := reflectTime(d, v)
		require.NotNil(t, typ)

		assert.Equal(t, typ.Type, tTypeString)
		assert.Equal(t, typ.Format, "date-time")
	})
	t.Run("ReflectIP_returns_ValidType", func(t *testing.T) {
		d := Definitions{}
		v := reflect.ValueOf(net.IP{})

		typ := reflectIP(d, v)
		require.NotNil(t, typ)

		assert.Equal(t, typ.Type, tTypeString)
		assert.Equal(t, typ.Format, "ipv4")
	})
	t.Run("ReflectURI_returns_ValidType", func(t *testing.T) {
		d := Definitions{}
		v := reflect.ValueOf(url.URL{})

		typ := reflectURI(d, v)
		require.NotNil(t, typ)

		assert.Equal(t, typ.Type, tTypeString)
		assert.Equal(t, typ.Format, "uri")
	})
	t.Run("ReflectPBEnum_returns_ValidType", func(t *testing.T) {
		t.Skip("implement")
		//d := Definitions{}
		//v := reflect.ValueOf()
		//
		//typ := reflectPBEnum(d)
		//require.NotNil(t, typ)
	})
	t.Run("ReflectEnum_returns_ValidType", func(t *testing.T) {
		a := assert.New(t)
		r := require.New(t)

		d := Definitions{}
		enumImpl := enumImpl{}
		enumVariants := enumImpl.Enum()

		v := reflect.ValueOf(enumImpl)

		typ := reflectEnum(d, v)
		r.NotNil(typ)
		r.Len(typ.Enum, len(enumVariants))
		a.Equal(tTypeString, typ.Type)

		a.Equal(typ.Enum[0], enumVariants[0])

		a.Equal(typ.Enum[1], enumVariants[1])

		a.Equal(typ.Enum[2], enumVariants[2])
	})
	t.Run("ReflectOneOf_returns_ValidType", func(t *testing.T) {
		a := assert.New(t)
		r := require.New(t)

		d := Definitions{}
		oneOfImpl := implicitOneOfImpl{}
		oneOfImplVariants := oneOfImpl.OneOf()

		v := reflect.ValueOf(oneOfImpl)

		typ := reflectOneOf(d, v)
		r.NotNil(typ)

		a.Len(typ.OneOf, len(oneOfImplVariants))
		a.Equal(typ.OneOf[0].Type, tTypeString)
		a.Equal(typ.OneOf[0].Default, oneOfImplVariants[0])

		a.Equal(typ.OneOf[1].Type, tTypeString)
		a.Equal(typ.OneOf[1].Default, oneOfImplVariants[1])

		a.Equal(typ.OneOf[2].Type, tTypeString)
		a.Equal(typ.OneOf[2].Default, oneOfImplVariants[2])
	})
	t.Run("ReflectAnyOf_returns_ValidType", func(t *testing.T) {
		a := assert.New(t)
		r := require.New(t)

		d := Definitions{}
		anyOfImpl := implicitAnyOfImpl{}
		anyOfImplVariants := anyOfImpl.AnyOf()

		v := reflect.ValueOf(anyOfImpl)

		typ := reflectAnyOf(d, v)
		r.NotNil(typ)

		a.Len(typ.AnyOf, len(anyOfImplVariants))
		a.Equal(typ.AnyOf[0].Type, tTypeString)
		a.Equal(typ.AnyOf[0].Default, anyOfImplVariants[0])

		a.Equal(typ.AnyOf[1].Type, tTypeString)
		a.Equal(typ.AnyOf[1].Default, anyOfImplVariants[1])

		a.Equal(typ.AnyOf[2].Type, tTypeString)
		a.Equal(typ.AnyOf[2].Default, anyOfImplVariants[2])
	})
	t.Run("ReflectAllOf_returns_ValidType", func(t *testing.T) {
		a := assert.New(t)
		r := require.New(t)

		d := Definitions{}
		allOfImpl := implicitAllOfImpl{}
		allOfImplVariants := allOfImpl.AllOf()

		v := reflect.ValueOf(allOfImpl)

		typ := reflectAllOf(d, v)
		r.NotNil(typ)

		a.Len(typ.AllOf, len(allOfImplVariants))
		a.Equal(typ.AllOf[0].Type, tTypeString)
		a.Equal(typ.AllOf[0].Default, allOfImplVariants[0])

		a.Equal(typ.AllOf[1].Type, tTypeString)
		a.Equal(typ.AllOf[1].Default, allOfImplVariants[1])

		a.Equal(typ.AllOf[2].Type, tTypeString)
		a.Equal(typ.AllOf[2].Default, allOfImplVariants[2])
	})

	t.Run("ReflectSlice", func(t *testing.T) {
		a := assert.New(t)
		r := require.New(t)

		t.Run("ReflectSlice_returns_ValidTypeOnBoolSLice", func(t *testing.T) {
			d := Definitions{}
			slice := []bool{true, false, false}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeBoolean)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnIntSLice", func(t *testing.T) {
			d := Definitions{}
			slice := []int{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeInteger)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnInt8SLice", func(t *testing.T) {
			d := Definitions{}
			slice := []int8{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeInteger)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnInt16SLice", func(t *testing.T) {
			d := Definitions{}
			slice := []int16{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeInteger)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnInt32SLice", func(t *testing.T) {
			d := Definitions{}
			slice := []int32{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeInteger)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnInt64SLice", func(t *testing.T) {
			d := Definitions{}
			slice := []int64{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeInteger)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnUintSLice", func(t *testing.T) {
			d := Definitions{}
			slice := []uint{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeInteger)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnUint8SLice", func(t *testing.T) {
			t.Skip("int8 slice handles as []byte. fix it")
			d := Definitions{}
			slice := []uint8{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeInteger)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnUint16SLice", func(t *testing.T) {
			d := Definitions{}
			slice := []uint16{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeInteger)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnUint32SLice", func(t *testing.T) {
			d := Definitions{}
			slice := []uint32{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeInteger)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnUint64SLice", func(t *testing.T) {
			d := Definitions{}
			slice := []uint64{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeInteger)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnFloat32SLice", func(t *testing.T) {
			d := Definitions{}
			slice := []float32{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeNumber)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnFloat64SLice", func(t *testing.T) {
			d := Definitions{}
			slice := []float64{1, 2, 3}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeNumber)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnArraySLice", func(t *testing.T) {
			d := Definitions{}
			array := [4]int{1, 2, 3, 4}

			v := reflect.ValueOf(array)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(tTypeArray, typ.Type)
			r.NotNil(typ.Items)

			a.Equal(tTypeInteger, typ.Items.Type)
			a.Equal(4, typ.MaxItems)
			a.Equal(4, typ.MinItems)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnByteSLice", func(t *testing.T) {
			d := Definitions{}
			array := make([]byte, 16)

			v := reflect.ValueOf(array)

			typ := reflectSlice(d, v)
			r.NotNil(typ)

			a.Equal(tTypeString, typ.Type)
			a.Equal("base64", typ.Media.BinaryEncoding)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnInterfaceSLice", func(t *testing.T) {
			d := Definitions{}
			slice := []interface{}{"1", "2", "3"}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(tTypeArray, typ.Type)
			r.NotNil(typ.Items)

			a.Equal(tTypeObject, typ.Items.Type)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnMapSLice", func(t *testing.T) {
			d := Definitions{}
			slice := []interface{}{
				map[string]interface{}{"1": 1},
				map[string]interface{}{"2": 2},
				map[string]interface{}{"3": 3},
			}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeObject)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnStringSLice", func(t *testing.T) {
			d := Definitions{}
			slice := []string{"1", "2", "3"}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeString)
		})

		t.Run("ReflectSlice_returns_ValidTypeOnStructSLice", func(t *testing.T) {
			t.Skip("implement: handle slice of structs")
			d := Definitions{}
			slice := []interface{}{}

			v := reflect.ValueOf(slice)

			typ := reflectSlice(d, v)
			r.NotNil(typ)
			a.Equal(typ.Type, tTypeArray)
			r.NotNil(typ.Items)

			a.Equal(typ.Items.Type, tTypeObject)
		})

	})
	t.Run("ReflectMap_returns_ValidType", func(t *testing.T) {
		d := Definitions{}
		v := reflect.ValueOf(map[string]interface{}{})

		typ := reflectMap(d, v)
		require.NotNil(t, typ)

		assert.Equal(t, typ.Type, tTypeObject)
		assert.Contains(t, typ.PatternProperties, ".*")
	})
	t.Run("ReflectInteger_returns_ValidType", func(t *testing.T) {
		d := Definitions{}
		v := reflect.ValueOf(int(666))

		typ := reflectInteger(d, v)
		require.NotNil(t, typ)

		assert.Equal(t, typ.Type, tTypeInteger)
		assert.Equal(t, typ.Default, 666)
	})
	t.Run("ReflectNumber_returns_ValidType", func(t *testing.T) {
		d := Definitions{}
		v := reflect.ValueOf(float64(666))

		typ := reflectNumber(d, v)
		require.NotNil(t, typ)

		assert.Equal(t, typ.Type, tTypeNumber)
		assert.Equal(t, float64(666), typ.Default)
	})
	t.Run("ReflectBool_returns_ValidType", func(t *testing.T) {
		d := Definitions{}
		v := reflect.ValueOf(float64(666))

		typ := reflectNumber(d, v)
		require.NotNil(t, typ)

		assert.Equal(t, typ.Type, tTypeNumber)
		assert.Equal(t, float64(666), typ.Default)
	})
	t.Run("ReflectString_returns_ValidType", func(t *testing.T) {
		d := Definitions{}
		v := reflect.ValueOf("666")

		typ := reflectString(d, v)
		require.NotNil(t, typ)

		assert.Equal(t, typ.Type, tTypeString)
		assert.Equal(t, "666", typ.Default)
	})
	t.Run("ReflectInterface_returns_ValidType", func(t *testing.T) {
		d := Definitions{}

		var sValue interface{} = "666"
		vValue := reflect.ValueOf(sValue)
		vType := reflect.TypeOf(sValue)

		typ := reflectInterface(d, vType, vValue)
		require.NotNil(t, typ)

		assert.Equal(t, typ.Type, tTypeObject)
	})
}
