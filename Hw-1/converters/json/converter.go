package json

import (
	"encoding/json"
	"fmt"

	"hw_1_serialization/converters"
)

type Converter struct {
}

func (c *Converter) Serialize(p *converters.MyStruct) ([]byte, error) {
	bytes, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("Error in json marshal failed")
	}
	return bytes, nil
}

func (c *Converter) Deserialize(raw []byte) (*converters.MyStruct, error) {
	person := &converters.MyStruct{}
	err := json.Unmarshal(raw, person)
	if err != nil {
		return nil, fmt.Errorf("Error in marshal json bytes")
	}
	return person, nil
}
