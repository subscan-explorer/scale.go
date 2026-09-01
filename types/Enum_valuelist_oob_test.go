package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/itering/scale.go/source"
	scaleBytes "github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
)

// A ValueList enum decodes its variant index from a single untrusted input byte
// (0-255). If that index is beyond the value list, the decoder previously did an
// unguarded slice access e.ValueList[e.Index] and panicked with an opaque
// "runtime error: index out of range". This mirrors the descriptive bounds check
// already present on the TypeMapping.Types path so the recovering caller gets the
// offending type/index instead of a bare runtime panic.
func TestEnumValueListIndexOutOfRange(t *testing.T) {
	RegCustomTypes(map[string]source.TypeStruct{
		"OOBValueEnum": {Type: "enum", ValueList: []string{"A", "B", "C"}},
	})

	// Out-of-range index byte (0x10 = 16, list length 3) -> descriptive panic.
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected a panic for an out-of-range enum index")
			}
			msg := fmt.Sprint(r)
			if !strings.Contains(msg, "index out of range") ||
				!strings.Contains(msg, "value list length") {
				t.Fatalf("panic message not descriptive: %v", r)
			}
		}()
		m := ScaleDecoder{}
		m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("0x10")}, nil)
		m.ProcessAndUpdateData("OOBValueEnum")
	}()

	// In-range index still decodes correctly.
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("0x01")}, nil)
	if got := m.ProcessAndUpdateData("OOBValueEnum"); got != "B" {
		t.Fatalf("in-range decode = %v, want B", got)
	}
}
