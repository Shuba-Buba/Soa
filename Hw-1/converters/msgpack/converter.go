package msgpack

import (
	"fmt"

	"github.com/shamaton/msgpack/v2"

	"hw_1_serialization/converters"
)

type Converter struct {
}

func (c *Converter) Serialize(p *converters.MyStruct) ([]byte, error) {
	bytes, err := msgpack.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("json marshal failed")
	}
	return bytes, nil
}

func (c *Converter) Deserialize(raw []byte) (*converters.MyStruct, error) {
	person := &converters.MyStruct{}
	if err := msgpack.Unmarshal(raw, person); err != nil {
		return nil, fmt.Errorf("failed to marshal json bytes")
	}
	return person, nil
}
