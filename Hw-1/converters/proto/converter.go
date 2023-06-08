package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	"hw_1_serialization/converters"
)

type Converter struct{}

func (c *Converter) Serialize(p *converters.MyStruct) ([]byte, error) {
	bytes, err := proto.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("Error with proto marshal")
	}
	return bytes, nil
}

func (c *Converter) Deserialize(raw []byte) (*converters.MyStruct, error) {
	person := &converters.MyStruct{}
	if err := proto.Unmarshal(raw, person); err != nil {
		return nil, fmt.Errorf("Error with marshal proto")
	}
	return person, nil
}
