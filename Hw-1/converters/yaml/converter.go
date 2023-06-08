package yaml

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"hw_1_serialization/converters"
)

type Converter struct {
}

func (c *Converter) Serialize(p *converters.MyStruct) ([]byte, error) {
	bytes, err := yaml.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("json marshal failed: %v", err)
	}
	return bytes, nil
}

func (c *Converter) Deserialize(raw []byte) (*converters.MyStruct, error) {
	person := &converters.MyStruct{}
	if err := yaml.Unmarshal(raw, person); err != nil {
		return nil, fmt.Errorf("failed to marshal json bytes")
	}
	return person, nil
}
