package models

// primitive types as response
type Integer int

// struct as response
type StructResponse1 struct {
	PrimitiveType int                        `json:"primitive_type"`
	SliceType     []int                      `json:"slice_type"`
	MapType       map[string]string          `json:"map_type_string_string"`
	MapType2      map[string]StructResponse2 `json:"map_type_string_struct"`
	// MapType2      map[string]StructResponse1 `json:"map_type_string_struct"` recursive
	PointerType   *int        `json:"pointer_type_int"`
	PointerType2  *Integer    `json:"pointer_type_int_response"`
	InterfaceType interface{} `json:"interface_type"`
}

type StructResponse2 struct {
	PrimitiveType int `json:"primitive_type"`
}
