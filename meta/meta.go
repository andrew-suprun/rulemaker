package meta

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"league.com/rulemaker/model"
	"league.com/rulemaker/util"
)

type Type int

type Meta map[string]Type

func (m Meta) String() string {
	return util.ToJson(m)
}

const (
	Invalid Type = iota
	Bool
	Int
	Float
	String
	Date
	Duration
	Map
	Slice
)

func (t Type) String() string {
	switch t {
	case Invalid:
		return "Invalid" // localizer.Ignore
	case Bool:
		return "Bool" // localizer.Ignore
	case Int:
		return "Int" // localizer.Ignore
	case Float:
		return "Float" // localizer.Ignore
	case String:
		return "String" // localizer.Ignore
	case Date:
		return "Date" // localizer.Ignore
	case Duration:
		return "Duration" // localizer.Ignore
	case Map:
		return "Map" // localizer.Ignore
	}
	return "Unknown Type" // localizer.Ignore
}

const (
	EntityMap         string = "{}" // localizer.Ignore
	AppendToSlice     string = "+"  // localizer.Ignore
	UpdateLastElement string = "-"  // localizer.Ignore
)

func Metainfo(entity interface{}) (result Meta) {
	result = Meta{}
	meta := deref(reflect.TypeOf(entity))
	for i := 0; i < meta.NumField(); i++ {
		field := meta.Field(i)
		collectMetainfo(field.Type, fieldName(field), result)
	}

	return result
}

func (meta Meta) Type(field string) Type {
	if _, ok := meta[field]; ok {
		return meta[field]
	}
	fieldParts := strings.Split(field, ".")
outer:
	for metaField, kind := range meta {
		metaFieldParts := strings.Split(metaField, ".")
		if len(metaFieldParts) != len(fieldParts) {
			continue
		}
		for i := range fieldParts {
			if fieldParts[i] != metaFieldParts[i] && metaFieldParts[i] != EntityMap && metaFieldParts[i] != AppendToSlice && metaFieldParts[i] != UpdateLastElement {
				continue outer
			}
		}
		return kind
	}
	return Invalid
}

var typeOfDate = reflect.TypeOf(time.Time{})
var typeOfDuration = reflect.TypeOf(time.Duration(0))

func collectMetainfo(meta reflect.Type, path string, result Meta) {
	meta = deref(meta)
	if meta == typeOfDate {
		result[path] = Date
		return
	} else if meta == typeOfDuration {
		result[path] = Duration
		return
	}
	switch meta.Kind() {
	case reflect.Map:
		collectMetainfo(meta.Elem(), path+".{}"+meta.Name(), result) // localizer.Ignore
	case reflect.Slice:
		collectMetainfo(meta.Elem(), path+".+", result) // localizer.Ignore
		collectMetainfo(meta.Elem(), path+".-", result) // localizer.Ignore
		result[path] = Slice
	case reflect.Struct:
		for i := 0; i < meta.NumField(); i++ {
			field := meta.Field(i)
			collectMetainfo(field.Type, path+"."+fieldName(field), result)
		}
	case reflect.Bool:
		result[path] = Bool
	case reflect.Int:
		result[path] = Int
	case reflect.Float64:
		result[path] = Float
	case reflect.String:
		result[path] = String
	case reflect.Interface:
		result[path] = Map
	}
}

func fieldName(field reflect.StructField) string {
	if name, ok := field.Tag.Lookup("json"); ok {
		return name
	}
	return strings.ToLower(field.Name)
}

func deref(meta reflect.Type) reflect.Type {
	for meta.Kind() == reflect.Ptr {
		meta = meta.Elem()
	}
	return meta
}

func Get(e model.Entity, field string) (result interface{}) {
	entity := e
	path := strings.Split(field, ".")
	for _, field := range path[:len(path)-1] {
		subEntity := entity[field]
		if subEntity == nil {
			return nil
		}
		if slice, ok := subEntity.([]model.Entity); ok {
			return slice
		}
		entity, _ = subEntity.(model.Entity)
	}
	return entity[path[len(path)-1]]
}

func Set(e model.Entity, field string, value interface{}) {
	entity := e
	path := strings.Split(field, ".")
	for i := 0; i < len(path)-1; i++ {
		field := path[i]
		nextField := path[i+1]
		switch nextField {
		case AppendToSlice:
			newEntity := model.Entity{}
			entity[field] = []model.Entity{newEntity}
			entity = newEntity
			i++
		case UpdateLastElement:
			if slice, ok := entity[field].([]model.Entity); ok {
				entity = slice[len(slice)-1]
			} else {
				newEntity := model.Entity{}
				entity[field] = []model.Entity{newEntity}
				entity = newEntity
			}
			i++
		default:
			subEntity, ok := entity[field]
			if !ok {
				subEntity = model.Entity{}
				entity[field] = subEntity
			}
			entity = subEntity.(model.Entity)
		}
	}
	entity[path[len(path)-1]] = value
}

func CommonType(params []interface{}) (result []interface{}) {
	commonType := reflect.TypeOf("")
	commonKind := reflect.Invalid
	for _, param := range params {
		if param == nil {
			continue
		}
		if _, ok := param.(error); ok {
			return []interface{}{param}
		}
		pType := reflect.TypeOf(param)
		pKind := pType.Kind()
		if pKind == reflect.String {
			if commonKind == reflect.Invalid {
				commonKind = reflect.String
			}
			continue
		}
		if commonKind == reflect.Float64 && pKind == reflect.Int {
			continue
		}
		if commonKind == reflect.Int && pKind == reflect.Float64 {
			commonType = pType
			commonKind = pKind
			continue
		}
		if commonKind != reflect.Invalid && commonKind != reflect.String && commonKind != pKind {
			return []interface{}{fmt.Errorf("Incompatible types %s and %s", commonType, pType)} // localizer.Ignore
		}
		commonType = pType
		commonKind = pKind

	}

	switch commonType {
	case typeOfDate:
		return convertSliceToType(params, Date)
	case typeOfDuration:
		return convertSliceToType(params, Duration)
	}

	switch commonKind {
	case reflect.String:
		return convertSliceToType(params, String)
	case reflect.Bool:
		return convertSliceToType(params, Bool)
	case reflect.Int:
		return convertSliceToType(params, Int)
	case reflect.Float64:
		return convertSliceToType(params, Float)
	}

	return params
}

func convertSliceToType(params []interface{}, kind Type) (result []interface{}) {
	result = make([]interface{}, len(params))
	for i := range params {
		result[i] = ConvertValueToType(params[i], kind)
		if _, ok := result[i].(error); ok {
			return []interface{}{result[i]}
		}
	}
	return result
}

func ConvertValueToType(param interface{}, kind Type) interface{} {
	if param == nil {
		return nil
	}
	switch kind {
	case Bool:
		return convertToBool(param)
	case Int:
		return convertToInt(param)
	case Float:
		return ConvertToFloat(param)
	case String:
		return ConvertToString(param)
	case Date:
		return ConvertToDate(param)
	case Duration:
		return convertToDuration(param)
	case Invalid:
		return param
	}
	return fmt.Errorf("cannot convert value '%v' to %v", param, kind) // localizer.Ignore
}

func ConvertToString(param interface{}) interface{} {
	value := reflect.ValueOf(param)
	if value.Kind() != reflect.String {
		return fmt.Errorf("cannot convert value '%[1]v' of type %[1]T to string", param) // localizer.Ignore
	}
	return reflect.ValueOf(param).String()
}

func convertToBool(param interface{}) interface{} {
	value := reflect.ValueOf(param)
	if value.Kind() == reflect.String {
		str := value.String()
		if str == "true" {
			return true
		} else if str == "false" {
			return false
		}
		return fmt.Errorf("value '%v' is neither 'true' nor 'false'", param) // localizer.Ignore
	}
	if value.Kind() != reflect.Bool {
		return fmt.Errorf("cannot convert value '%v' to Boolean", param) // localizer.Ignore
	}
	return reflect.ValueOf(param).Bool()
}

func convertToInt(param interface{}) interface{} {
	paramKind := reflect.TypeOf(param).Kind()
	if paramKind == reflect.Int {
		return int(reflect.ValueOf(param).Int())
	}
	if paramKind == reflect.Float64 {
		return int(reflect.ValueOf(param).Float())
	}
	if paramKind == reflect.String {
		str := reflect.ValueOf(param).String()
		if str == "" {
			return 0
		}
		i64, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not a number", str) // localizer.Ignore
		}
		return int(i64)
	}
	return fmt.Errorf("cannot convert value '%v' to int", param) // localizer.Ignore
}

func ConvertToFloat(param interface{}) interface{} {
	paramKind := reflect.TypeOf(param).Kind()
	if paramKind == reflect.Float64 {
		return reflect.ValueOf(param).Float()
	}
	if paramKind == reflect.Int {
		return float64(reflect.ValueOf(param).Int())
	}
	if paramKind == reflect.String {
		str := reflect.ValueOf(param).String()
		if str == "" {
			return 0
		}
		if strings.HasPrefix(str, "$") {
			str = strings.ReplaceAll(strings.ReplaceAll(str, ",", ""), "$", "")
		}
		f64, err := strconv.ParseFloat(str, 10)
		if err == nil {
			return f64
		}

	}
	return fmt.Errorf("cannot convert value '%v' to float", param) // localizer.Ignore
}

func ConvertToDate(param interface{}) interface{} {

	paramType := reflect.TypeOf(param)
	if paramType == typeOfDate {
		return param.(time.Time)
	}
	if paramType.Kind() == reflect.String {
		str := reflect.ValueOf(param).String()
		t, err := time.Parse("2006-01-02", str)
		if err == nil {
			return t
		}
		t, err = time.Parse("01/02/2006", str)
		if err == nil {
			return t
		}
		t, err = time.Parse("01-02-2006", str)
		if err == nil {
			return t
		}
		return fmt.Errorf("cannot convert '%s' to date", param) // localizer.Ignore
	}
	return fmt.Errorf("cannot convert '%s' to date", param) // localizer.Ignore
}

func convertToDuration(param interface{}) interface{} {
	paramType := reflect.TypeOf(param)
	if paramType == typeOfDuration {
		return param
	}
	if reflect.TypeOf(param).Kind() == reflect.String {
		d, err := time.ParseDuration(reflect.ValueOf(param).String())
		if err != nil {
			return fmt.Errorf("cannot convert '%s' to duration", param) // localizer.Ignore
		}
		return d
	}
	return fmt.Errorf("cannot convert '%s' to duration", param) // localizer.Ignore
}
