package XML

import (
	"encoding/xml"
	"fmt"

	"hw_1_serialization/converters"
)

type Converter struct {
}

func (c *Converter) Serialize(p *converters.MyStruct) ([]byte, error) {
	bytes, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("xml marshal failed: %v", err)
	}
	return bytes, nil
}

func (c *Converter) Deserialize(raw []byte) (*converters.MyStruct, error) {
	person := &converters.MyStruct{}
	if err := xml.Unmarshal(raw, person); err != nil {
		return nil, fmt.Errorf("failed to unmarshal xml bytes: %v", err)
	}
	return person, nil
}
