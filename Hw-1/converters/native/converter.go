package native

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"hw_1_serialization/converters"
)

type Converter struct{}

func (c *Converter) Serialize(p *converters.MyStruct) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		return nil, fmt.Errorf("Error in xml marshal failed")
	}
	return buf.Bytes(), nil
}

func (c *Converter) Deserialize(raw []byte) (*converters.MyStruct, error) {
	var buf bytes.Buffer
	buf.Write(raw)

	person := &converters.MyStruct{}
	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(person); err != nil {
		return nil, fmt.Errorf("Error unmarshal native bytes")
	}
	return person, nil
}
