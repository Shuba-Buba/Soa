package avro

import (
	"fmt"

	"github.com/hamba/avro"

	"hw_1_serialization/converters"
)

type Converter struct {
	schema avro.Schema
}

func (c *Converter) SetSchema() error {
	schema, err := avro.Parse(`{
		"type": "record",
		"name": "me",
		"namespace": "org.hamba.avro",
		"fields" : [
			{"name": "String_", "type": "string"},
			{"name": "Int_", "type": "int"},
			{"name": "Map_", "type": {"type":"map", "values": "string"}},
			{"name": "Array_", "type": {"type":"array", "items": "string"}},
			{"name": "Float_", "type": "float"}
		]
	}`)
	if err != nil {
		return fmt.Errorf("failed to parse struct")
	}
	c.schema = schema
	return nil
}

func (c *Converter) Serialize(p *converters.MyStruct) ([]byte, error) {
	bytes, err := avro.Marshal(c.schema, p)
	if err != nil {
		return nil, fmt.Errorf("avro marshal failed")
	}
	return bytes, nil
}

func (c *Converter) Deserialize(raw []byte) (*converters.MyStruct, error) {
	person := &converters.MyStruct{}
	err := avro.Unmarshal(c.schema, raw, person)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal avro bytes")
	}
	return person, nil
}
