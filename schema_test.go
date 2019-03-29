package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//func newReference(typ string) *Type {
//	return &Type{Ref: fmt.Sprintf("#/definitions/%s", typ)}
//}

func TestNewReference(t *testing.T) {
	ref := newReference(tTypeString)
	require.NotNil(t, ref)
	assert.Equal(t, "#/definitions/string", ref.Ref)
}
