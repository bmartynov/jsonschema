package jsonschema

import (
	"encoding/json"
	"os"
	"testing"
)

const (
	enumAlignCenter = "center"
)

type EnumAlign string

func (EnumAlign) Enum() []interface{} {
	return []interface{}{
		enumAlignCenter,
	}
}

type config struct {
	host  string
	Items []pageConfig `json:"items,omitempty"`
}

type backgroundConfig struct {
	BackgroundColor ColorPicker `json:"backgroundColor,omitempty"`
}

type buttonConfig struct {
	Title string `json:"title"`
}

type imageConfig struct {
	Align EnumAlign   `json:"align,omitempty"`
	Image ColorPicker `json:"image"`
	Width float64     `json:"width,omitempty"`
}

type textConfig struct {
	Align           string      `json:"align,omitempty"`
	BackgroundColor ColorPicker `json:"backgroundColor,omitempty"`
	MarkupType      string      `json:"markupType,omitempty"`
	Subtitle        string      `json:"subtitle,omitempty"`
	TextColor       ColorPicker `json:"textColor,omitempty"`
	Title           string      `json:"title,omitempty"`
	Width           float64     `json:"width,omitempty"`
}

type pageConfig struct {
	Background *backgroundConfig `json:"background,omitempty"`
	Button     *buttonConfig     `json:"button,omitempty"`
	Image      *imageConfig      `json:"image,omitempty"`
	Text       *textConfig       `json:"text,omitempty"`
}

func TestComplexTypes(t *testing.T) {
	schema := Reflect(config{})

	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "\t")

	e.Encode(schema)

}
