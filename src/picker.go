package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/buger/jsonparser"

	"github.com/pkg/errors"
)

const pathSeparator = "/"

// Property
type Property struct {
	Type  string  `json:"type"`
	Path  string  `json:"path"`
	Alias *string `json:"alias"`
}

// Name : returns the property name
func (p Property) Name() string {
	paths := strings.Split(p.Path, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[len(paths)-1]
}

// Config : Config
type Config struct {
	Properties []Property `json:"properties"`
}

// Response Response
type Response map[string]interface{}

// ToJSON : in JSON string
func (r *Response) ToJSON() string {
	return string(r.ToBytes())
}

// ToBytes : in Bytes
func (r *Response) ToBytes() []byte {
	d, err := json.Marshal(r)
	if err != nil {
		return []byte{}
	}
	return d
}

// JSONParser :  Wrapper
type JSONParser struct {
	body []byte
}

// Int : Get
func (p *JSONParser) Int(prop Property) (int64, error) {
	return jsonparser.GetInt(p.body, resovlePropertyPath(prop.Path)...)
}

// Float : Get
func (p *JSONParser) Float(prop Property) (float64, error) {
	return jsonparser.GetFloat(p.body, resovlePropertyPath(prop.Path)...)
}

// Bool : Get
func (p *JSONParser) Bool(prop Property) (bool, error) {
	return jsonparser.GetBoolean(p.body, resovlePropertyPath(prop.Path)...)
}

// String : Get
func (p *JSONParser) String(prop Property) (string, error) {
	return jsonparser.GetString(p.body, resovlePropertyPath(prop.Path)...)
}

// Object : Get
func (p *JSONParser) Object(prop Property) (map[string]interface{}, error) {
	v, _, _, err := jsonparser.Get(p.body, resovlePropertyPath(prop.Path)...)
	if err != nil {
		return nil, err
	}
	obj := map[string]interface{}{}
	err = json.Unmarshal(v, &obj)
	return obj, err
}

// IntSlice : Get
func (p *JSONParser) IntSlice(prop Property) ([]int64, error) {
	values := []int64{}
	var parserErr error
	_, err := jsonparser.ArrayEach(p.body, func(value []byte, dataType jsonparser.ValueType, _ int, _ error) {
		t := fmt.Sprintf(`{ "v" : %s}`, value)
		v, err := jsonparser.GetInt([]byte(t), "v")
		if err != nil {
			parserErr = err
			return
		}
		values = append(values, v)
	}, resovlePropertyPath(prop.Path)...)
	if err != nil {
		return values, err
	}
	return values, parserErr
}

// FloatSlice  : Get
func (p *JSONParser) FloatSlice(prop Property) ([]float64, error) {
	values := []float64{}
	var parserErr error
	_, err := jsonparser.ArrayEach(p.body, func(value []byte, dataType jsonparser.ValueType, _ int, _ error) {
		t := fmt.Sprintf(`{ "v" : %s}`, value)
		v, err := jsonparser.GetFloat([]byte(t), "v")
		if err != nil {
			parserErr = err
			return
		}
		values = append(values, v)
	}, resovlePropertyPath(prop.Path)...)
	if err != nil {
		return values, err
	}
	return values, parserErr
}

// BoolSlice : Get
func (p *JSONParser) BoolSlice(prop Property) ([]bool, error) {
	values := []bool{}
	var parserErr error
	_, err := jsonparser.ArrayEach(p.body, func(value []byte, dataType jsonparser.ValueType, _ int, _ error) {
		t := fmt.Sprintf(`{ "v" : %s}`, value)
		v, err := jsonparser.GetBoolean([]byte(t), "v")
		if err != nil {
			parserErr = err
			return
		}
		values = append(values, v)
	}, resovlePropertyPath(prop.Path)...)
	if err != nil {
		return values, err
	}
	return values, parserErr
}

// StringSlice  : Get
func (p *JSONParser) StringSlice(prop Property) ([]string, error) {
	values := []string{}
	var parserErr error
	_, err := jsonparser.ArrayEach(p.body, func(value []byte, dataType jsonparser.ValueType, _ int, _ error) {
		values = append(values, string(value))
	}, resovlePropertyPath(prop.Path)...)
	if err != nil {
		return values, err
	}
	return values, parserErr
}

// ObjectSlice : Get
func (p *JSONParser) ObjectSlice(prop Property) ([]map[string]interface{}, error) {
	values := []map[string]interface{}{}
	var parserErr error
	if prop.Path == "." {
		_, err := jsonparser.ArrayEach(p.body, func(value []byte, dataType jsonparser.ValueType, _ int, _ error) {
			t := fmt.Sprintf(`{ "v" : %s}`, value)
			jp := &JSONParser{body: []byte(t)}
			v, err := jp.Object(Property{Path: "v"})
			if err != nil {
				parserErr = err
				return
			}
			values = append(values, v)
		})
		if err != nil {
			return values, err
		}
		return values, parserErr
	}
	_, err := jsonparser.ArrayEach(p.body, func(value []byte, dataType jsonparser.ValueType, _ int, _ error) {
		t := fmt.Sprintf(`{ "v" : %s}`, value)
		jp := &JSONParser{body: []byte(t)}
		v, err := jp.Object(Property{Path: "v"})
		if err != nil {
			parserErr = err
			return
		}
		values = append(values, v)
	}, resovlePropertyPath(prop.Path)...)
	if err != nil {
		return values, err
	}
	return values, parserErr
}

// SliceObject : Get
func (p *JSONParser) SliceObject(prop Property) ([]interface{}, error) {
	newPath, err := getObjectSliceKey(prop)
	if err != nil {
		return nil, err
	}
	newProperty := Property{
		Alias: prop.Alias,
		Type:  prop.Type,
		Path:  newPath}

	objects, err := p.ObjectSlice(newProperty)
	if err != nil {
		return nil, err
	}
	values := []interface{}{}
	for _, o := range objects {
		if v, ok := o[prop.Name()]; ok {
			values = append(values, v)
		}
	}
	return values, nil
}

// SliceObjectProperty : Get
func (p *JSONParser) SliceObjectProperty(prop Property) ([]interface{}, error) {
	newPath, objPropName, err := getObjectSlicePropertyKey(prop)
	if err != nil {
		return nil, err
	}
	newProperty := Property{
		Alias: prop.Alias,
		Type:  prop.Type,
		Path:  newPath}
	objects, err := p.ObjectSlice(newProperty)
	if err != nil {
		return nil, err
	}
	values := []interface{}{}
	for _, obj := range objects {
		if objProp, ok := obj[objPropName]; ok {
			if objPropMap, ok := objProp.(map[string]interface{}); ok {
				if v, ok := objPropMap[prop.Name()]; ok {
					values = append(values, v)
					continue
				}
			}
			values = append(values, objProp)
		}
	}
	return values, nil
}

func getObjectSliceKey(prop Property) (string, error) {
	paths := resovlePropertyPath(prop.Path)
	pathsLen := len(paths)
	if len(paths) < 2 {
		return "", fmt.Errorf("invalid slice object key")
	}
	if paths[0] == "." {
		return strings.Join(paths, pathSeparator), nil
	}
	limit := (pathsLen - 1)
	return strings.Join(paths[:limit], pathSeparator), nil
}

func getObjectSlicePropertyKey(prop Property) (string, string, error) {
	objectSliceKey, err := getObjectSliceKey(prop)
	if err != nil {
		return "", "", nil
	}
	paths := resovlePropertyPath(objectSliceKey)
	pathsLen := len(paths)
	if pathsLen < 2 {
		return "", "", fmt.Errorf("invalid slice object property key")
	}
	limit := (pathsLen - 1)
	return strings.Join(paths[:limit], pathSeparator), paths[limit], nil
}

// PickUsingJSONConfig : Pick Using JSON Config String
func PickUsingJSONConfig(input io.Reader, configJSON string) (*Response, error) {
	var config Config
	err := json.Unmarshal([]byte(configJSON), &config)
	if err != nil {
		return nil, errors.Wrap(err, "invalid config JSON")
	}
	return PickUsingConfig(input, config)
}

// PickUsingConfig : Typed Version of PickUsingJSONConfig
func PickUsingConfig(input io.Reader, config Config) (*Response, error) {
	body, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read input")
	}
	jp := &JSONParser{body: body}
	res := Response{}
	for _, p := range config.Properties {
		key, value, err := pickProperty(jp, p)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("unable to pick the property '%s'", p.Name()))
		}
		res[key] = value
	}
	return &res, nil
}

func pickProperty(jp *JSONParser, prop Property) (string, interface{}, error) {
	switch propType := prop.Type; propType {
	case "i":
		v, err := jp.Int(prop)
		return handlePickPropertyResult(prop, v, err)
	case "[i]":
		v, err := jp.IntSlice(prop)
		return handlePickPropertyResult(prop, v, err)
	case "f":
		v, err := jp.Float(prop)
		return handlePickPropertyResult(prop, v, err)
	case "[f]":
		v, err := jp.FloatSlice(prop)
		return handlePickPropertyResult(prop, v, err)
	case "b":
		v, err := jp.Bool(prop)
		return handlePickPropertyResult(prop, v, err)
	case "[b]":
		v, err := jp.BoolSlice(prop)
		return handlePickPropertyResult(prop, v, err)
	case "s":
		v, err := jp.String(prop)
		return handlePickPropertyResult(prop, v, err)
	case "[s]":
		v, err := jp.StringSlice(prop)
		return handlePickPropertyResult(prop, v, err)
	case "o":
		v, err := jp.Object(prop)
		return handlePickPropertyResult(prop, v, err)
	case "[o]":
		v, err := jp.ObjectSlice(prop)
		return handlePickPropertyResult(prop, v, err)
	case "[]o":
		v, err := jp.SliceObject(prop)
		return handlePickPropertyResult(prop, v, err)
	case "[]op":
		v, err := jp.SliceObjectProperty(prop)
		return handlePickPropertyResult(prop, v, err)
	default:
		return pickPropertyError(fmt.Errorf("un-supported property type '%s'", propType))
	}
}

func handlePickPropertyResult(prop Property, value interface{}, err error) (string, interface{}, error) {
	if err != nil {
		return pickPropertyError(err)
	}
	name := prop.Name()
	if prop.Alias != nil {
		name = *prop.Alias
	}
	return name, value, nil
}

func pickPropertyError(err error) (string, interface{}, error) {
	return "", nil, err
}

func resovlePropertyPath(path string) []string {
	return strings.Split(path, pathSeparator)
}

// PickDeserializedUsingJSONConfig : Pick JSON and Deserialize
func PickDeserializedUsingJSONConfig(input io.Reader, configJSON string, propName string, value interface{}) error {
	res, err := PickUsingJSONConfig(input, configJSON)
	if err != nil {
		return err
	}
	if body, ok := (*res)[propName]; ok {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return err
		}
		return json.Unmarshal(bodyJSON, value)
	}
	return fmt.Errorf("deserialization failed : property name not found")
}

// PickDeserializedUsingConfig : Typed Version of PickDeserializedUsingJSONConfig
func PickDeserializedUsingConfig(input io.Reader, config Config, propName string, value interface{}) error {
	res, err := PickUsingConfig(input, config)
	if err != nil {
		return err
	}
	if propName == "" {
		return json.Unmarshal(res.ToBytes(), value)
	}
	if body, ok := (*res)[propName]; ok {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return err
		}
		return json.Unmarshal(bodyJSON, value)
	}
	return fmt.Errorf("deserialization failed : property name not found")
}
