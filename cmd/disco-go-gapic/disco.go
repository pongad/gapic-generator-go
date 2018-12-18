package main

type DiscoDoc struct {
	Name    string
	Schemas map[string]Schema
}

type Schema struct {
	ID string

	Type, Format         string
	Ref                  string `json:"$ref"`
	Items                *Schema
	AdditionalProperties *Schema

	Properties map[string]Schema
}
