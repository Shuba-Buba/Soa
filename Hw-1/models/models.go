package models

type Serializators struct {
	Types []Serializator `json:"types"`
}

type Serializator struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func (this *Serializators) GetArray() []Serializator {
	return this.Types
}
