package types

import (
	"fmt"
	"sync"
	"testing"

	"github.com/itering/scale.go/source"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
)

func TestRegCustomTypesConcurrency(t *testing.T) {
	wg := sync.WaitGroup{}
	count := 0
	for {
		count++
		go func() {
			RegCustomTypes(map[string]source.TypeStruct{fmt.Sprintf("%d", count): {Type: "string", TypeString: "u32"}})
			wg.Add(1)
			wg.Done()
		}()
		if count > 100 {
			break
		}
	}

	wg.Wait()
}

func TestCodecCacheConcurrency(t *testing.T) {
	RegCustomTypes(map[string]source.TypeStruct{
		"MyVec": {Type: "string", TypeString: "Vec<U32>"},
	})
	raw := "080100000002000000"
	errCh := make(chan error, 1000)
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := ScaleDecoder{}
			m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
			result := m.ProcessAndUpdateData("MyVec")
			values, ok := result.([]interface{})
			if !ok || len(values) != 2 {
				errCh <- fmt.Errorf("decode mismatch: %v", result)
				return
			}
			encoded := Encode("MyVec", []interface{}{1, 2})
			if encoded != raw {
				errCh <- fmt.Errorf("encode mismatch: %s", encoded)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}
