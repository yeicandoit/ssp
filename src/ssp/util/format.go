package util

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

func mapIndexConvert(t reflect.Type, s string) (interface{}, error) {
	if t.Kind() == reflect.String {
		return s, nil
	}

	if t.Kind() >= reflect.Int && t.Kind() <= reflect.Int64 {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		switch t.Kind() {
		case reflect.Int:
			return int(i), nil
		case reflect.Int8:
			return int8(i), nil
		case reflect.Int16:
			return int16(i), nil
		case reflect.Int32:
			return int32(i), nil
		case reflect.Int64:
			return int64(i), nil
		}
	}
	if t.Kind() >= reflect.Uint && t.Kind() <= reflect.Uint64 {
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		switch t.Kind() {
		case reflect.Uint:
			return uint(i), nil
		case reflect.Uint8:
			return uint8(i), nil
		case reflect.Uint16:
			return uint16(i), nil
		case reflect.Uint32:
			return uint32(i), nil
		case reflect.Uint64:
			return uint64(i), nil
		}
	}

	return nil, errors.New("Unsupported type: " + t.String())
}

// i should be struct/map/array or pointer to it
func GetInValue(i interface{}, keys []string) interface{} {
	if len(keys) == 0 {
		return i
	}

	t := reflect.TypeOf(i)
	switch t.Kind() {
	case reflect.Ptr:
		iv := reflect.Indirect(reflect.ValueOf(i))
		if iv.CanInterface() {
			return GetInValue(iv.Interface(), keys)
		} else {
			return nil
		}
	case reflect.Map:
		index, err := mapIndexConvert(t.Key(), keys[0])
		if err != nil {
			return nil
		}

		v := reflect.ValueOf(i).MapIndex(reflect.ValueOf(index))
		if v.IsValid() {
			return GetInValue(v.Interface(), keys[1:])
		} else {
			return nil
		}
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		ai := i.([]interface{})
		index, err := strconv.Atoi(keys[0])
		if err != nil || len(ai) <= index {
			return nil
		}
		return GetInValue(ai[index], keys[1:])
	case reflect.Struct:
		v := reflect.ValueOf(i).FieldByName(keys[0])
		if v.IsValid() {
			return GetInValue(v.Interface(), keys[1:])
		} else {
			return nil
		}
	default:
		return nil
	}
}

func DumpString(i interface{}) string {
	buffer := dumpValue(reflect.ValueOf(i))
	return buffer.String()
}

func dumpValue(i reflect.Value) *bytes.Buffer {
	buffer := bytes.NewBuffer(make([]byte, 0))
	if i.Kind() != reflect.Invalid {
		buffer.WriteString(i.Type().Name() + ":\t")
	}
	dumpLevel(i, 0, false, buffer)
	return buffer
}

func dumpLevel(i reflect.Value, lv int, is_struct_val bool, buffer *bytes.Buffer) {
	r := i.Kind()

	switch {
	case r == reflect.UnsafePointer || r == reflect.Uintptr: // unsafe ptr ignore
	case r == reflect.Ptr:
		v := reflect.Indirect(i)
		dumpLevel(v, lv, is_struct_val, buffer)
	case r == reflect.Array || r == reflect.Slice:
		putHeadTab(buffer, lv, is_struct_val)
		buffer.WriteString("[\n")

		for j := 0; j < i.Len(); j++ {
			dumpLevel(i.Index(j), lv+1, false, buffer)
		}
		putTailTab(buffer, lv)
		buffer.WriteString("]\n")
	case r == reflect.Map:
		putHeadTab(buffer, lv, is_struct_val)
		ks := i.MapKeys()
		if len(ks) == 0 {
			buffer.WriteString("{}\n")
		} else {
			buffer.WriteString("{\n")
			for _, ik := range ks {
				putChars(buffer, '\t', lv+1)
				buffer.WriteString(keyValue(ik) + ":\t")
				iv := i.MapIndex(ik)
				dumpLevel(iv, lv+1, true, buffer)
			}
			putTailTab(buffer, lv)
			buffer.WriteString("}\n")
		}
	case r == reflect.String:
		putHeadTab(buffer, lv, is_struct_val)
		buffer.WriteString(fmt.Sprintf("\"%s\"\n", i.String()))
	case r == reflect.Struct:
		putHeadTab(buffer, lv, is_struct_val)
		buffer.WriteString("(\n")
		for j := 0; j < i.NumField(); j++ {
			putChars(buffer, '\t', lv+1)
			iv := i.Field(j)
			buffer.WriteString(i.Type().Field(j).Name + ":\t")
			dumpLevel(iv, lv+1, true, buffer)
		}
		putTailTab(buffer, lv)
		buffer.WriteString(")\n")
	case r == reflect.Bool:
		putHeadTab(buffer, lv, is_struct_val)
		buffer.WriteString(fmt.Sprintln(i.Bool()))
	case r >= reflect.Int && r <= reflect.Int64:
		putHeadTab(buffer, lv, is_struct_val)
		buffer.WriteString(fmt.Sprintln(i.Int()))
	case r >= reflect.Uint && r <= reflect.Uint64:
		putHeadTab(buffer, lv, is_struct_val)
		buffer.WriteString(fmt.Sprintln(i.Uint()))
	case r >= reflect.Float32 && r <= reflect.Float64:
		putHeadTab(buffer, lv, is_struct_val)
		buffer.WriteString(fmt.Sprintln(i.Float()))
	default: // ignore interface/channel/complex
		putHeadTab(buffer, lv, is_struct_val)
		buffer.WriteString(i.String() + "\n")
	}
}

func keyValue(v reflect.Value) string {
	t := v.Kind()
	switch {
	case t == reflect.String:
		return "\"" + v.String() + "\""
	case t >= reflect.Int && t <= reflect.Int64:
		return fmt.Sprint(v.Int())
	case t >= reflect.Uint && t <= reflect.Uint64:
		return fmt.Sprint(v.Uint())
	default:
		return v.String()
	}
}

func putHeadTab(buffer *bytes.Buffer, lv int, is_struct_val bool) {
	if !is_struct_val {
		putChars(buffer, '\t', lv)
	}
}

func putTailTab(buffer *bytes.Buffer, lv int) {
	putChars(buffer, '\t', lv)
}

func putChars(buffer *bytes.Buffer, v byte, n int) {
	for i := 0; i < n; i++ {
		buffer.WriteByte(v)
	}
}

type ReflectValue []reflect.Value

func (vals ReflectValue) Len() int      { return len(vals) }
func (vals ReflectValue) Swap(i, j int) { vals[i], vals[j] = vals[j], vals[i] }
func (vals ReflectValue) Less(i, j int) bool {
	less, err := func(i, j int) (less bool, err error) {
		defer func() {
			err = errors.New(fmt.Sprint(recover()))
		}()
		less = vals[i].Int() < vals[j].Int()
		return
	}(i, j)
	if err == nil {
		return less
	}

	less, err = func(i, j int) (less bool, err error) {
		defer func() {
			err = errors.New(fmt.Sprint(recover()))
		}()
		less = vals[i].Uint() < vals[j].Uint()
		return
	}(i, j)
	if err == nil {
		return less
	}

	return strings.Compare(vals[i].String(), vals[j].String()) == -1
}

func DumpSimple(i interface{}) string {
	if i == nil {
		return "<nil>"
	}
	t := reflect.TypeOf(i)
	switch t.Kind() {
	case reflect.Ptr:
		return DumpSimple(reflect.Indirect(reflect.ValueOf(i)).Interface())
	case reflect.Map:
		buffer := bytes.NewBuffer(make([]byte, 0))
		keys := reflect.ValueOf(i).MapKeys()
		count := len(keys)
		sort.Sort(ReflectValue(keys))
		buffer.WriteString("KeyType: " + t.Key().String() + "\tValType: " + t.Elem().String() + "\tCount: " + fmt.Sprint(count) + "\n")
		for i := 0; i < len(keys); i++ {
			buffer.WriteString(fmt.Sprintf("%v ", keys[i].Interface()))
		}
		return buffer.String()
	case reflect.Struct:
		buffer := bytes.NewBuffer(make([]byte, 0))
		buffer.WriteString("{\n")
		v := reflect.ValueOf(i)
		for j := 0; j < v.NumField(); j++ {
			buffer.WriteString("\t")
			iv := v.Field(j)
			if iv.CanInterface() {
				buffer.WriteString(v.Type().Field(j).Name + ":")
				buffer.WriteString(fmt.Sprintf("%T\n", iv.Interface()))
			} else {
				buffer.WriteString(v.Type().Field(j).Name + ":" + iv.Type().Name() + "\n")
			}
		}
		buffer.WriteString("}\n")
		return buffer.String()
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		return fmt.Sprintf("%s (len=%)d\n", t.Name(), len(i.([]interface{})))
	default:
		return fmt.Sprintf("%+v\n", i)
	}
}
