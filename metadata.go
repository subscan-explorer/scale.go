package scalecodec

import (
	"errors"
	"fmt"

	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
)

type MetadataDecoder struct {
	types.ScaleDecoder
	Version  string               `json:"version"`
	Metadata types.MetadataStruct `json:"metadata"`
}

func (m *MetadataDecoder) Init(data []byte) {
	sData := scaleBytes.ScaleBytes{Data: data}
	m.ScaleDecoder.Init(sData, nil)
}

func (m *MetadataDecoder) Process() error {
	magicBytes := m.NextBytes(4)
	if string(magicBytes) != "meta" {
		m.Data.Reset()
		m.ProcessAndUpdateData("Option<Compact<u32>>") // Option<Compact<u32>>
		magicBytes = m.NextBytes(4)
	}
	if string(magicBytes) == "meta" {
		metadataVersion := utiles.U256(utiles.BytesToHex(m.Data.Data[m.Data.Offset : m.Data.Offset+1]))
		m.Version = m.ProcessAndUpdateData("MetadataVersion").(string)
		m.Metadata = m.ProcessAndUpdateData(m.Version).(types.MetadataStruct)
		m.Metadata.MetadataVersion = int(metadataVersion.Int64())
		if m.Metadata.MetadataVersion == 14 && m.Data.GetRemainingLength() != 0 {
			return fmt.Errorf("metadata V14 has %d trailing bytes", m.Data.GetRemainingLength())
		}
		return nil
	}
	return errors.New("not metadata")

}

func (m *MetadataDecoder) PortableRegistry() (map[int]types.SiType, error) {
	magicBytes := m.NextBytes(4)
	if string(magicBytes) != "meta" { // metadata version >= v15
		m.Data.Reset()
		m.ProcessAndUpdateData("Option<Compact<u32>>")
		magicBytes = m.NextBytes(4)
	}
	if string(magicBytes) == "meta" {
		metadataVersion := utiles.U256(utiles.BytesToHex(m.Data.Data[m.Data.Offset : m.Data.Offset+1]))
		m.Version = m.ProcessAndUpdateData("MetadataVersion").(string)
		if metadataVersion.Int64() < 14 {
			return nil, errors.New("PortableRegistry not support, only metadata version >= v14")
		}
		portable := types.InitPortableRaw(m.ProcessAndUpdateData("PortableRegistry").([]interface{}))
		var registry = make(map[int]types.SiType)
		for index, item := range portable {
			item.RemoveDocs()
			registry[index] = item
		}
		return registry, nil
	}
	return nil, errors.New("not metadata")
}
