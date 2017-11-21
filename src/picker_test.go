package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PickFirstLevelIntegers(t *testing.T) {
	input := `{ "id" : 1, "age" : 14, "path" : "john" }`
	config := `{
		"properties" : [
			{"path" : "id", "type" : "i" },
			{"path" : "age", "type" : "i"}
		]
	}`
	expected := `{"id" : 1, "age" : 14}`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}
func Test_PickFirstLevelFloat(t *testing.T) {
	input := `{ "id" : 1, "age" : 14, "name" : "john", "height" : 12.8 }`
	config := `{
		"properties" : [
			{"path" : "height", "type" : "f" }
		]
	}`
	expected := `{"height" : 12.8}`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}
func Test_PickFirstLevelIntAndFloat(t *testing.T) {
	input := `{ "id" : 1, "age" : 14, "name" : "john", "height" : 12.8 }`
	config := `{
		"properties" : [
			{"path" : "height", "type" : "f" },
			{"path" : "age", "type" : "i" }
		]
	}`
	expected := `{"height" : 12.8, "age" : 14}`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}

func Test_PickFirstLevelBool(t *testing.T) {
	input := `{ "id" : 1, "age" : 14, "name" : "john", "isAdmin" : true }`
	config := `{
		"properties" : [
			{"path" : "isAdmin", "type" : "b" }
		]
	}`
	expected := `{"isAdmin" : true}`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}
func Test_PickFirstLevelString(t *testing.T) {
	input := `{ "id" : 1, "age" : 14, "name" : "john", "isAdmin" : true }`
	config := `{
		"properties" : [
			{"path" : "name", "type" : "s" }
		]
	}`
	expected := `{"name" : "john"}`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}
func Test_PickFirstLevelObject(t *testing.T) {
	input := `{ "id" : 1, "address" : {"country" : "india", "pin" : 600041 } }`
	config := `{
		"properties" : [
			{"path" : "address", "type" : "o" }
		]
	}`
	expected := `{"address" : {"country" : "india", "pin" : 600041 }}`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}

func Test_PickFirstLevelArrayOfIntegers(t *testing.T) {
	input := `{ "id" : 1, "name" : "john", "marks" : [1,2,3] }`
	config := `{
		"properties" : [
			{"path" : "marks", "type" : "[i]" }
		]
	}`
	expected := `{"marks" : [1,2,3]}`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}
func Test_PickFirstLevelArrayOfFloats(t *testing.T) {
	input := `{ "id" : 1, "name" : "john", "heights" : [1.2,2.4,3.4] }`
	config := `{
		"properties" : [
			{"path" : "heights", "type" : "[f]" }
		]
	}`
	expected := `{"heights" : [1.2,2.4,3.4] }`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}

func Test_PickFirstLevelArrayOfStrings(t *testing.T) {
	input := `{ "id" : 1, "name" : "john", "tags" : ["fair", "asian", "tall"] }`
	config := `{
		"properties" : [
			{"path" : "tags", "type" : "[s]" }
		]
	}`
	expected := `{"tags" : ["fair", "asian", "tall"] }`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}

func Test_PickFirstLevelArrayOfBools(t *testing.T) {
	input := `{ "id" : 1, "name" : "john", "hits" : [false, true, true, false] }`
	config := `{
		"properties" : [
			{"path" : "hits", "type" : "[b]" }
		]
	}`
	expected := `{ "hits" : [false, true, true, false] }`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}

func Test_PickFirstLevelArrayOfObjects(t *testing.T) {
	input := `{
		"url" : "http://foo.com",
		"properties" : [
			{"path" : "id", "type" : "i" },
			{"path" : "age", "type" : "i"}
		]
	}`
	config := `{
		"properties" : [
			{"path" : "properties", "type" : "[o]" }
		]
	}`
	expected := `{
		"properties" : [
			{"path" : "id", "type" : "i" },
			{"path" : "age", "type" : "i"}
		]	
	}`

	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}

var nestedJSONInput = `
{
	"foo": 1,
	"bar": 2,
	"test": "Hello, world!",
	"baz": 123.1,
	"array": [
		{"foo": 1, "size" : 1, "test" : { "id" : 1 }},
		{"bar": 2, "size" : 2, "test" : { "id" : 2 }},
		{"baz": 3, "size" : 3, "test" : { "id" : 3 }}
	],
	"subobj": {
		"foo": 10,
		"subarray": [1,2,3],
		"subsubobj": {
			"bar": 2,
			"baz": 3,
			"array": ["hello", "world"]
		}
	},
	"bool": true
}
`

func Test_PickNestedJsonValues(t *testing.T) {
	var pickTests = []struct {
		propPath string
		propType string
		expected string
	}{
		{"subobj/subsubobj/bar", "i", `{"bar" : 2}`},
		{"subobj/foo", "i", `{"foo" : 10}`},
	}
	for _, tt := range pickTests {
		configFormat := `{
			"properties" : [
				{"path" : "%s", "type" : "%s"}
			]
		}`
		config := fmt.Sprintf(configFormat, tt.propPath, tt.propType)

		res, err := PickUsingJSONConfig(strings.NewReader(nestedJSONInput), config)

		assert.Nil(t, err)
		assert.JSONEq(t, tt.expected, res.ToJSON())
	}
}

func Test_PickValuesFromObjArray(t *testing.T) {
	config := `{
		"properties" : [
			{"path" : "array/size", "type" : "[]o"}
		]
	}`
	expected := `{
		"size" : [1,2,3]
	}`

	res, err := PickUsingJSONConfig(strings.NewReader(nestedJSONInput), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())

}
func Test_PickObjPropertyFromObjArray(t *testing.T) {
	config := `{
		"properties" : [
			{"path" : "array/test/id", "type" : "[]op"}
		]
	}`
	expected := `{
		"id" : [1,2,3]
	}`

	res, err := PickUsingJSONConfig(strings.NewReader(nestedJSONInput), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())

}

func Test_PickFromArrayOfObjects(t *testing.T) {
	input := `[{"id" : 1}, {"id" : 2}, {"id" : 3}]`

	config := `{
		"properties" : [
			{"path" : "./id", "type" : "[]op"}
		]
	}`
	expected := `{
		"id" : [1,2,3]
	}`
	res, err := PickUsingJSONConfig(strings.NewReader(input), config)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, res.ToJSON())
}

func Test_PickWithDeserialization(t *testing.T) {
	input := `[{"id" : 1}, {"id" : 2}, {"id" : 3}]`

	config := `{
		"properties" : [
			{"path" : "./id", "type" : "[]op"}
		]
	}`
	expected := []int{1, 2, 3}
	var actual []int
	err := PickDeserializedUsingJSONConfig(strings.NewReader(input), config, "id", &actual)

	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}
