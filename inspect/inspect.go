package inspect

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func Vars(objs map[string]interface{}) {
	inspectVarsPath, exists := os.LookupEnv("INSPECT_VARS")
	if !exists {
		return
	}

	data, jsonErr := json.Marshal(objs)
	if jsonErr != nil {
		return
	}

	if _, err := os.Stat(inspectVarsPath); os.IsNotExist(err) {
		os.Mkdir(inspectVarsPath, 0755)
	}

	file, fileErr := os.OpenFile(inspectVarsPath, os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.Encode(data)
}

func Dump(obj interface{}) string {
	jsonBytes, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return err.Error()
	}
	return strings.ReplaceAll(string(jsonBytes), "\"", "")
}

func PrintDump(obj interface{}) {
	fmt.Println(Dump(obj))
}

func AllKeys(objsMapList []map[string]interface{}) []string {
	var allKeys []string
	for _, objsMap := range objsMapList {
		for _, oKey := range reflect.ValueOf(objsMap).MapKeys() {
			key := oKey.Interface().(string)
			if !Contains(allKeys, key) {
				allKeys = append(allKeys, key)
			}
		}
	}
	return allKeys
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func DumpTable(slice interface{}) string {
	options := TableOptions{}
	return options.DumpTable(slice)
}

func (options TableOptions) DumpTable(slice interface{}) string {
	tableString := &strings.Builder{}
	options.Writer = tableString
	options.PrintDumpTable(slice)
	return tableString.String()
}

func (options TableOptions) PrintDumpTable(slice interface{}) {
	objs := AsInterfaces(slice)
	writer := options.Writer
	if writer == nil {
		writer = os.Stdout
	}
	table := tablewriter.NewWriter(writer)
	table.SetAutoWrapText(false)
	objsMapList := asInterfaceMap(objs)
	headers := options.Headers
	if headers == nil {
		headers = AllKeys(objsMapList)
	}
	table.SetHeader(headers)
	for _, e := range objsMapList {
		values := make([]string, 0, len(headers))
		for _, k := range headers {
			value := fmt.Sprintf("%v", e[k])
			values = append(values, value)
		}
		table.Append(values)
	}
	if options.Filter != nil {
		options.Filter(table)
	}
	table.Render()
}

func PrintDumpTable(objs interface{}) {
	TableOptions{Writer: os.Stdout}.PrintDumpTable(objs)
}

type TableFilter func(*tablewriter.Table)

type TableOptions struct {
	Headers []string
	Writer  io.Writer
	Filter  TableFilter
}

//archived: https://github.com/fatih/structs
var (
	// DefaultTagName is the default tag name for struct fields which provides
	// a more granular to tweak certain structs. Lookup the necessary functions
	// for more info.
	DefaultTagName = "structs" // struct's field default tag name
)

type Struct struct {
	raw     interface{}
	value   reflect.Value
	TagName string
}

func AsInterfaces(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	// Keep the distinction between nil and empty slice input
	if s.IsNil() {
		return nil
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

func asInterfaceMap(objs []interface{}) []map[string]interface{} {
	var objsMapList []map[string]interface{}
	for _, element := range objs {
		objsMapList = append(objsMapList, Map(element))
	}
	return objsMapList
}

func strctVal(s interface{}) reflect.Value {
	v := reflect.ValueOf(s)

	// if pointer get the underlying element≤
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("not struct")
	}

	return v
}

func New(s interface{}) *Struct {
	return &Struct{
		raw:     s,
		value:   strctVal(s),
		TagName: DefaultTagName,
	}
}
func Map(s interface{}) map[string]interface{} {
	return New(s).Map()
}
func (s *Struct) Map() map[string]interface{} {
	out := make(map[string]interface{})
	s.FillMap(out)
	return out
}
func (s *Struct) structFields() []reflect.StructField {
	t := s.value.Type()

	var f []reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// we can't access the value of unexported fields
		if field.PkgPath != "" {
			continue
		}

		// don't check if it's omitted
		if tag := field.Tag.Get(s.TagName); tag == "-" {
			continue
		}

		f = append(f, field)
	}

	return f
}

type tagOptions []string

func (t tagOptions) Has(opt string) bool {
	for _, tagOpt := range t {
		if tagOpt == opt {
			return true
		}
	}

	return false
}
func parseTag(tag string) (string, tagOptions) {
	// tag is one of followings:
	// ""
	// "name"
	// "name,opt"
	// "name,opt,opt2"
	// ",opt"

	res := strings.Split(tag, ",")
	return res[0], res[1:]
}

func (s *Struct) FillMap(out map[string]interface{}) {
	if out == nil {
		return
	}

	fields := s.structFields()

	for _, field := range fields {
		name := field.Name
		val := s.value.FieldByName(name)
		isSubStruct := false
		var finalVal interface{}

		tagName, tagOpts := parseTag(field.Tag.Get(s.TagName))
		if tagName != "" {
			name = tagName
		}

		// if the value is a zero value and the field is marked as omitempty do
		// not include
		if tagOpts.Has("omitempty") {
			zero := reflect.Zero(val.Type()).Interface()
			current := val.Interface()

			if reflect.DeepEqual(current, zero) {
				continue
			}
		}

		if !tagOpts.Has("omitnested") {
			finalVal = s.nested(val)

			v := reflect.ValueOf(val.Interface())
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			switch v.Kind() {
			case reflect.Map, reflect.Struct:
				isSubStruct = true
			}
		} else {
			finalVal = val.Interface()
		}

		if tagOpts.Has("string") {
			s, ok := val.Interface().(fmt.Stringer)
			if ok {
				out[name] = s.String()
			}
			continue
		}

		if isSubStruct && (tagOpts.Has("flatten")) {
			for k := range finalVal.(map[string]interface{}) {
				out[k] = finalVal.(map[string]interface{})[k]
			}
		} else {
			out[name] = finalVal
		}
	}
}

func (s *Struct) nested(val reflect.Value) interface{} {
	var finalVal interface{}

	v := reflect.ValueOf(val.Interface())
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		n := New(val.Interface())
		n.TagName = s.TagName
		m := n.Map()

		// do not add the converted value if there are no exported fields, ie:
		// time.Time
		if len(m) == 0 {
			finalVal = val.Interface()
		} else {
			finalVal = m
		}
	case reflect.Map:
		// get the element type of the map
		mapElem := val.Type()
		switch val.Type().Kind() {
		case reflect.Ptr, reflect.Array, reflect.Map,
			reflect.Slice, reflect.Chan:
			mapElem = val.Type().Elem()
			if mapElem.Kind() == reflect.Ptr {
				mapElem = mapElem.Elem()
			}
		}

		// only iterate over struct types, ie: map[string]StructType,
		// map[string][]StructType,
		if mapElem.Kind() == reflect.Struct ||
			(mapElem.Kind() == reflect.Slice &&
				mapElem.Elem().Kind() == reflect.Struct) {
			m := make(map[string]interface{}, val.Len())
			for _, k := range val.MapKeys() {
				m[k.String()] = s.nested(val.MapIndex(k))
			}
			finalVal = m
			break
		}

		// TODO(arslan): should this be optional?
		finalVal = val.Interface()
	case reflect.Slice, reflect.Array:
		if val.Type().Kind() == reflect.Interface {
			finalVal = val.Interface()
			break
		}

		// TODO(arslan): should this be optional?
		// do not iterate of non struct types, just pass the value. Ie: []int,
		// []string, co... We only iterate further if it's a struct.
		// i.e []foo or []*foo
		if val.Type().Elem().Kind() != reflect.Struct &&
			!(val.Type().Elem().Kind() == reflect.Ptr &&
				val.Type().Elem().Elem().Kind() == reflect.Struct) {
			finalVal = val.Interface()
			break
		}

		slices := make([]interface{}, val.Len())
		for x := 0; x < val.Len(); x++ {
			slices[x] = s.nested(val.Index(x))
		}
		finalVal = slices
	default:
		finalVal = val.Interface()
	}

	return finalVal
}
