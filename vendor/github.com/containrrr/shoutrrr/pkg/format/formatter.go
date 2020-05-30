package format

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/containrrr/shoutrrr/pkg/util"
)

// GetConfigMap returns a string map of a given Config struct
func GetConfigMap(service types.Service) (map[string]string, int) {

	cr := reflect.ValueOf(service).Elem().FieldByName("config")
	cr = reflect.NewAt(cr.Type(), unsafe.Pointer(cr.UnsafeAddr())).Elem()
	config := cr.Interface().(types.ServiceConfig)

	formatter := formatter{
		EnumFormatters: config.Enums(),
		MaxDepth:       10,
	}
	return formatter.getStructMap(config, 0)
}

type formatter struct {
	EnumFormatters map[string]types.EnumFormatter
	MaxDepth       uint8
	Errors         []error
}

func (fmtr *formatter) getStructMap(structItem interface{}, depth uint8) (map[string]string, int) {
	defs := reflect.TypeOf(structItem)
	values := reflect.ValueOf(structItem)

	if defs.Kind() == reflect.Ptr {
		values = values.Elem()
		defs = defs.Elem()
	}

	numFields := values.NumField()
	valueMap := make(map[string]string, numFields)
	nextDepth := depth + 1
	maxKeyLen := 0

	for i := 0; i < numFields; i++ {
		fieldDef := defs.Field(i)
		if fieldDef.Anonymous {
			// This is an embedded field, which should not be part of the Config output
			continue
		}

		value := fmt.Sprintf("(%s)", fieldDef.Type.Name())
		valueLen := len(value)

		ef, isEnum := fmtr.EnumFormatters[fieldDef.Name]
		if isEnum {
			fieldVal := values.Field(i)
			kind := fieldVal.Kind()
			if kind == reflect.Int {
				valueStr := ef.Print(int(fieldVal.Int()))
				value = ColorizeEnum(valueStr)
				valueLen = len(valueStr)
			} else {
				err := fmt.Errorf("incorrect enum type '%s' for field '%s'", kind, fieldDef.Name)
				fmtr.Errors = append(fmtr.Errors, err)
			}
		} else if nextDepth < fmtr.MaxDepth {
			value, valueLen = fmtr.getFieldValueString(values.Field(i), nextDepth)
		}

		if tag, ok := fieldDef.Tag.Lookup("desc"); ok {

			prePad := strings.Repeat(" ", util.Max(40-valueLen, 1))
			postPad := strings.Repeat(" ", util.Max(60-len(tag), 1))

			value += " " + prePad + ColorizeDesc(tag) + postPad
		}

		if tag, ok := fieldDef.Tag.Lookup("tpl"); ok {
			value += fmt.Sprintf(" <Template: %s>", ColorizeString(tag))
		}

		if tag, ok := fieldDef.Tag.Lookup("default"); ok {
			value += fmt.Sprintf(" <Default: %s>", ColorizeValue(tag, isEnum))
		}

		if isEnum {
			value += fmt.Sprintf(" [%s]", strings.Join(ef.Names(), ", "))
		}

		valueMap[fieldDef.Name] = value
		keyLen := len(fieldDef.Name)
		if keyLen > maxKeyLen {
			maxKeyLen = keyLen
		}
	}

	return valueMap, maxKeyLen
}

func (fmtr *formatter) getFieldValueString(field reflect.Value, depth uint8) (string, int) {

	nextDepth := depth + 1
	kind := field.Kind()

	if util.IsUnsignedDecimal(kind) {
		strVal := fmt.Sprintf("%d", field.Uint())
		return ColorizeNumber(fmt.Sprintf("%s", strVal)), len(strVal)
	}
	if util.IsSignedDecimal(kind) {
		strVal := fmt.Sprintf("%d", field.Int())
		return ColorizeNumber(fmt.Sprintf("%s", strVal)), len(strVal)
	}
	if kind == reflect.String {
		strVal := field.String()
		return ColorizeString(strVal), len(strVal)
	}
	if kind == reflect.Bool {
		val := field.Bool()
		if val {
			return ColorizeTrue(PrintBool(val)), 3
		}
		return ColorizeFalse(PrintBool(val)), 2

	}

	if util.IsCollection(kind) {
		len := field.Len()
		items := make([]string, len)
		totalLen := 4
		var itemLen int
		for i := 0; i < field.Len(); i++ {
			items[i], itemLen = fmtr.getFieldValueString(field.Index(i), nextDepth)
			totalLen += itemLen
		}
		return fmt.Sprintf("[ %s ]", strings.Join(items, ", ")), totalLen
	}

	if kind == reflect.Map {
		items := make([]string, field.Len())
		iter := field.MapRange()
		index := 0
		// initial value for totalLen is surrounding curlies and spaces, and separating commas
		totalLen := 4 + (field.Len() - 1)
		for iter.Next() {
			key, keyLen := fmtr.getFieldValueString(iter.Key(), nextDepth)
			value, valueLen := fmtr.getFieldValueString(iter.Value(), nextDepth)
			items[index] = fmt.Sprintf("%s: %s", key, value)
			totalLen += keyLen + valueLen + 2
		}

		return fmt.Sprintf("{ %s }", strings.Join(items, ", ")), totalLen
	}
	if kind == reflect.Struct {
		structMap, _ := fmtr.getStructMap(field, depth+1)
		structFieldCount := len(structMap)
		items := make([]string, structFieldCount)
		index := 0
		totalLen := 4 + (structFieldCount - 1)
		for key, value := range structMap {
			items[index] = fmt.Sprintf("%s: %s", key, value)
			index++
			totalLen += len(key) + 2 + len(value)
		}
		return fmt.Sprintf("< %s >", strings.Join(items, ", ")), totalLen
	}
	strVal := kind.String()
	return fmt.Sprintf("<?%s>", strVal), len(strVal) + 5
}
