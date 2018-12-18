package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"unicode"
	"unicode/utf8"
)

func main() {
	var disco DiscoDoc
	if err := json.NewDecoder(os.Stdin).Decode(&disco); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("package %s\n", disco.Name)

	// schemas
	var tg typGen
	for _, s := range disco.Schemas {
		tg.schemas = append(tg.schemas, s)
	}
	heap.Init(&tg.schemas)

	for len(tg.schemas) > 0 {
		sch := heap.Pop(&tg.schemas).(Schema)
		tg.print(sch)
	}
}

type typGen struct {
	schemas SchemaSlice
}

type SchemaSlice []Schema

func (s *SchemaSlice) Len() int           { return len(*s) }
func (s *SchemaSlice) Less(i, j int) bool { return (*s)[i].ID < (*s)[j].ID }
func (s *SchemaSlice) Swap(i, j int)      { (*s)[i], (*s)[j] = (*s)[j], (*s)[i] }
func (s *SchemaSlice) Push(x interface{}) { *s = append(*s, x.(Schema)) }
func (s *SchemaSlice) Pop() interface{} {
	x := (*s)[s.Len()-1]
	*s = (*s)[:s.Len()-1]
	return x
}

func (tg *typGen) print(sch Schema) error {
	if sch.Type != "object" {
		return fmt.Errorf("expected top level schemas to be objects: %q", sch.ID)
	}

	var fields []Schema
	for id, prop := range sch.Properties {
		prop.ID = id
		fields = append(fields, prop)
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].ID < fields[j].ID })

	fmt.Printf("type %s struct{\n", sch.ID)

	for _, f := range fields {
		t, err := tg.goType(sch.ID, f)
		if err != nil {
			return err
		}
		fmt.Printf("\t%s %s\n", exportName(f.ID), t)
	}

	fmt.Printf("}\n")
	return nil
}

func (tg *typGen) goType(prefix string, sch Schema) (string, error) {
	if r := sch.Ref; r != "" {
		return r, nil
	}
	if sch.Type == "object" {
		if ap := sch.AdditionalProperties; ap != nil {
			elem, err := tg.goType(prefix+exportName(sch.ID), *ap)
			if err != nil {
				return "", err
			}
			return "map[string]" + elem, nil
		}
		if prop := sch.Properties; prop != nil {
			typ := prefix + exportName(sch.ID)
			heap.Push(&tg.schemas, Schema{
				ID:         typ,
				Type:       "object",
				Properties: prop,
			})
			return typ, nil
		}
		return "", fmt.Errorf("unrecognized object type")
	}
	if sch.Type == "array" {
		elem, err := tg.goType(prefix+exportName(sch.ID), *sch.Items)
		if err != nil {
			return "", err
		}
		return "[]" + elem, nil
	}

	t, ok := primitiveType[[2]string{sch.Type, sch.Format}]
	if !ok {
		return "", fmt.Errorf("unknown (type, format): (%q, %q)", sch.Type, sch.Format)
	}
	return t, nil
}

func exportName(s string) string {
	if s == "" {
		return ""
	}
	r, w := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[w:]
}

// map[[type, format]] type
var primitiveType = map[[2]string]string{
	{"boolean", ""}:       "bool",
	{"integer", "int32"}:  "int32",
	{"integer", "uint32"}: "uint32",
	{"number", "double"}:  "float64",
	{"number", "float"}:   "float32",
	{"string", ""}:        "string",
	{"string", "byte"}:    "[]byte",
	{"string", "int64"}:   "int64",
	{"string", "uint64"}:  "uint64",
}
