package jap

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func getValueAndField(data interface{}, path string) (*reflect.Value, *reflect.StructField, error) {
	path = strings.TrimPrefix(path, ".")

	parts := strings.Split(path, ".")

	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	var structField *reflect.StructField

	for _, part := range parts {
		re := regexp.MustCompile(`^([a-zA-Z0-9_]+)?\[(\d+)\]$`)
		matches := re.FindStringSubmatch(part)

		if matches != nil {
			fieldName := matches[1]
			index, _ := strconv.Atoi(matches[2])

			if fieldName != "" {
				field, found := typ.FieldByName(fieldName)
				if !found {
					return nil, nil, fmt.Errorf("field %s not found", fieldName)
				}
				structField = &field
				val = val.FieldByName(fieldName)
				typ = val.Type()
			}

			if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
				return nil, nil, fmt.Errorf("field %s is not a slice or array", part)
			}

			if index < 0 || index >= val.Len() {
				return nil, nil, fmt.Errorf("index %d out of range", index)
			}
			val = val.Index(index)

		} else {
			field, found := typ.FieldByName(part)
			if !found {
				return nil, nil, fmt.Errorf("field %s not found", part)
			}

			structField = &field
			val = val.FieldByName(part)
			typ = val.Type()
		}

		if val.Kind() == reflect.Ptr {
			val = val.Elem()
			typ = val.Type()
		}

		if !val.IsValid() {
			return nil, nil, fmt.Errorf("invalid path: %s", part)
		}
	}

	return &val, structField, nil
}

// Generate creates a cisco command line from the given struct.
// The command gets generated from the cmd tag of the struct and parsed with a printf.
func Generate(parsed any) (string, error) {
	var config string

	t := reflect.TypeOf(parsed)
	for i := 0; i < t.NumField(); i++ {
		cmd, err := GenerateField(t.Field(i), reflect.ValueOf(&parsed).Elem().Elem().Field(i))
		if err != nil {
			return "", err
		}
		if cmd == "" {
			continue
		}
		cmd = "  " + cmd + "\n"
		config = config + cmd
	}
	config = config + "!"
	return config, nil
}

func GenerateFieldByPath(parsed any, path string) (string, error) {
	val, structField, err := getValueAndField(parsed, path)
	if err != nil {
		return "", err
	}

	cmd, err := GenerateField(*structField, *val)
	if err != nil {
		return "", err
	}

	return cmd, nil
}

func GenerateField(field reflect.StructField, value reflect.Value) (string, error) {
	tag := field.Tag.Get("cmd")
	if tag != "" {
		defaultval := field.Tag.Get("default")
		cmd, err := generateCMD(value, tag, defaultval)
		if err != nil {
			return "", err
		}
		return cmd, nil
	}
	return "", nil
}

func generateCMD(field reflect.Value, tag, defaultval string) (string, error) {
	var cmd string
	switch field.Type().Kind() {
	case reflect.Struct:
		return Generate(field.Interface())
	case reflect.String:
		value := field.String()
		if value == "" {
			return "", nil
		}
		cmd = fmt.Sprintf(tag, value)
	case reflect.Int:
		value := field.Int()
		if value == 0 {
			return "", nil
		}
		cmd = fmt.Sprintf(tag, value)
	case reflect.Bool:
		value := field.Bool()
		if !value && defaultval != "" {
			return "no " + tag, nil
		} else if !value {
			return "", nil
		}
		cmd = tag
	case reflect.Float64:
		value := field.Float()
		if value == 0.0 {
			return "", nil
		}
		cmd = fmt.Sprintf(tag, value)
	case reflect.Slice:
		switch field.Type().String() {
		case "[]string":
			slice, _ := field.Interface().([]string)
			if len(slice) == 0 {
				return "", nil
			}
			cmds := ""
			for i2, s := range slice {
				cmd1 := fmt.Sprintf(tag, s)
				if i2 == 0 {
					cmds = cmds + cmd1
				} else {
					cmds = cmds + "\n" + "  " + cmd1
				}
			}
			cmd = cmd + cmds
		case "[]int":
			slice, _ := field.Interface().([]int)
			if len(slice) == 0 {
				return "", nil
			}
			var sliceStr []string
			for _, s := range slice {
				text := strconv.Itoa(s)
				sliceStr = append(sliceStr, text)
			}
			cmd = fmt.Sprintf(tag, strings.Join(sliceStr, ","))
		default:
			if field.Type().Kind() == reflect.Slice && tag != "" {

				for i2 := 0; i2 < field.Len(); i2++ {
					cmds := tag
					for i1 := 0; i1 < field.Type().Elem().NumField(); i1++ {
						defaultvali := field.Index(0).Type().Field(i1).Tag.Get("default")
						gcmd, _ := generateCMD(field.Index(i2).Field(i1), field.Index(0).Type().Field(i1).Tag.Get("cmd"), defaultvali)
						cmds = cmds + gcmd
						//cmds = fmt.Sprintf(field1.Index(0).Type().Field(i1).Tag.Get("cmd"), )
					}
					if i2 == 0 {
						cmd = cmd + cmds
					} else {
						cmd = cmd + "\n" + "  " + cmds
					}
				}
			}
		}
	default:
		panic(field.Type().Kind().String() + " not implemented")
	}

	return cmd, nil
}
