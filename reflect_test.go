package jsonschema

import (
	"encoding/json"
	"net"
	"net/url"
	"os"
	"testing"
	"time"
)

type OneOfTypeVariant1 struct {
	Type         string
	SomeIntValue int
}

func (o *OneOfTypeVariant2) variant() {}

type OneOfTypeVariant2 struct {
	Type            string
	SomeStringValue string
}

func (o *OneOfTypeVariant1) variant() {}

type OneOfType struct {
	Variant interface {
		variant()
	}
}

func (o *OneOfType) OneOf() []interface{} {
	return []interface{}{
		&OneOfTypeVariant1{},
		&OneOfTypeVariant2{},
	}
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
	Feeling ProtoEnum `json:"feeling,omitempty"`
	Age     int       `json:"age" jsonschema:"minimum=18,maximum=120,exclusiveMaximum=true,exclusiveMinimum=true"`
	Email   string    `json:"email" jsonschema:"format=email"`
	OneOf   OneOfType `json:"one_of"`
}

//var schemaGenerationTests = []struct {
//	reflector *Reflector
//	fixture   string
//}{
//	{&Reflector{}, "fixtures/defaults.json"},
//	{&Reflector{AllowAdditionalProperties: true}, "fixtures/allow_additional_props.json"},
//	{&Reflector{RequiredFromJSONSchemaTags: true}, "fixtures/required_from_jsontags.json"},
//	{&Reflector{ExpandedStruct: true}, "fixtures/defaults_expanded_toplevel.json"},
//}
//
//func TestSchemaOneOf(t *testing.T) {
//	reflector := &Reflector{}
//
//	schema := reflector.Reflect(&OneOfType{})
//
//	if _, ok := schema.Definitions["OneOfTypeVariant1"]; !ok {
//		t.Fatal("OneOfTypeVariant1 definition missed from schema")
//	}
//
//	if _, ok := schema.Definitions["OneOfTypeVariant2"]; !ok {
//		t.Fatal("OneOfTypeVariant2 definition missed from schema")
//	}
//
//	if len(schema.OneOf) != 2 {
//		t.Fatal("count schema one of must be equal 2")
//	}
//}
//
//func TestSchemaGeneration(t *testing.T) {
//	for _, tt := range schemaGenerationTests {
//		name := strings.TrimSuffix(filepath.Base(tt.fixture), ".json")
//		t.Run(name, func(t *testing.T) {
//			f, err := ioutil.ReadFile(tt.fixture)
//			if err != nil {
//				t.Errorf("ioutil.ReadAll(%s): %s", tt.fixture, err)
//				return
//			}
//
//			actualSchema := tt.reflector.Reflect(&TestUser{})
//			expectedSchema := &Schema{}
//
//			if err := json.Unmarshal(f, expectedSchema); err != nil {
//				t.Errorf("json.Unmarshal(%s, %v): %s", tt.fixture, expectedSchema, err)
//				return
//			}
//
//			if !reflect.DeepEqual(actualSchema, expectedSchema) {
//				actualJSON, err := json.MarshalIndent(actualSchema, "", "  ")
//				if err != nil {
//					t.Errorf("json.MarshalIndent(%v, \"\", \"  \"): %v", actualSchema, err)
//					return
//				}
//				t.Errorf("reflector %+v wanted schema %s, got %s", tt.reflector, f, actualJSON)
//			}
//		})
//	}
//}

type Action struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Link     string `json:"link,omitempty"`
	DeepLink string `json:"deeplink,omitempty"`
}

type DesignType struct {
	Type    string                 `json:"type,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type TrackingInfo struct {
	ID        uint64 `json:"id,omitempty"`
	Type      string `json:"type,omitempty"`
	DeepLink  string `json:"deeplink,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
}

type Image struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Image    string `json:"image,omitempty"`
	Link     string `json:"link,omitempty"`
	DeepLink string `json:"deeplink,omitempty"`
}

type Header struct {
	ID         string `json:"id,omitempty"`
	Type       string `json:"type,omitempty"`
	Title      string `json:"title,omitempty"`
	SubTitle   string `json:"subtitle,omitempty"`
	Link       string `json:"link,omitempty"`
	DeepLink   string `json:"deeplink,omitempty"`
	Disclosure *bool  `json:"disclosure,omitempty"`
}

type Footer struct {
	ID         string `json:"id,omitempty"`
	Type       string `json:"type,omitempty"`
	Title      string `json:"title,omitempty"`
	Link       string `json:"link,omitempty"`
	DeepLink   string `json:"deeplink,omitempty"`
	Disclosure *bool  `json:"disclosure,omitempty"`
}

type Widget struct {
	kind         string
	DesignType   *DesignType   `json:"designType,omitempty"`
	Header       *Header       `json:"header,omitempty"`
	Footer       *Footer       `json:"footer,omitempty"`
	Action       *Action       `json:"action,omitempty"`
	TrackingInfo *TrackingInfo `json:"sliceTrackingInfo"`
	Image        *Image        `json:"image,omitempty"`
}

type enum []string

func (dt enum) Enum() []interface{} {
	out := make([]interface{}, len(dt))

	for idx, e := range dt {
		out[idx] = e
	}

	return out
}

type DesignTypeOptions struct {
	SkuFinalPrice  *bool `json:"sku_final_price" title:"Показывать \"finalPrice\""`
	SkuPrice       *bool `json:"sku_price" title:"Показывать \"price\""`
	SkuDiscount    *bool `json:"sku_discount" title:"Показывать \"discount\""`
	SkuTitle       *bool `json:"sku_title" title:"Показывать \"title\""`
	SkuBrand       *bool `json:"sku_brand" title:"Показывать \"brand\""`
	SkuRating      *bool `json:"sku_rating" title:"Показывать \"rating\""`
	SkuIsAddToCart *bool `json:"sku_is_add_to_cart" title:"Показывать \"isAddToCard\""`
	SkuIsFavorite  *bool `json:"sku_is_favorite" title:"Показывать \"isFavorite\""`
	SkuMarketLabel *bool `json:"sku_market_label" title:"Показывать \"marketLabel\""`
}

type DesignType1 struct {
	Type    enum              `json:"type,omitempty" title:"Дезайн	"`
	Options DesignTypeOptions `json:"DesignTypeOptions,omitempty" title:"Настройки"`
}

func TestKek(t *testing.T) {
	r := Reflector{
		RequiredFromJSONSchemaTags: true,
		ExpandedStruct:             true,
		AllowAdditionalProperties:  true,
	}

	s := r.Reflect(DesignType1{
		Type:    enum{"grid1", "grid2", "grid3"},
		Options: DesignTypeOptions{},
	})

	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "\t")

	e.Encode(s)
}
