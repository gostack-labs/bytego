package bytego

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPROTOBUF          = "application/x-protobuf"
	MIMEYAML              = "application/x-yaml"
)

type binder struct {
	validate          Validate
	validateTranslate ValidateTranslate
}

func (b *binder) Bind(c *Ctx, i interface{}) error {
	if err := b.bindDefault(i, "default"); err != nil {
		return err
	}
	if err := b.bindParams(c, i); err != nil {
		return err
	}
	if err := b.bindQueries(c, i); err != nil {
		return err
	}
	if err := b.bindHeaders(c, i); err != nil {
		return err
	}
	if err := b.bindBody(c, i); err != nil {
		return err
	}
	if b.validate != nil {
		err := b.validate(i)
		if err != nil && b.validateTranslate != nil {
			return b.validateTranslate(err)
		}
		return err
	}
	return nil
}

func (b *binder) bindParams(c *Ctx, i interface{}) error {
	params := map[string][]string{}
	for _, p := range c.Params {
		params[p.Key] = []string{p.Value}
	}
	if err := b.bindData(i, params, "param", ""); err != nil {
		return err
	}
	return nil
}

func (b *binder) bindQueries(c *Ctx, i interface{}) error {
	return b.bindData(i, c.Request.URL.Query(), "query", "")
}
func (b *binder) bindHeaders(c *Ctx, i interface{}) error {
	return b.bindData(i, c.Request.Header, "header", "")
}

func (b *binder) bindBody(c *Ctx, i interface{}) error {
	if c.Request.ContentLength == 0 {
		return nil
	}
	switch c.ContentType() {
	case MIMEJSON:
		return json.NewDecoder(c.Request.Body).Decode(i)
	case MIMEXML, MIMEXML2:
		return xml.NewDecoder(c.Request.Body).Decode(i)
	case MIMEPOSTForm, MIMEMultipartPOSTForm:
		if err := c.Request.ParseForm(); err != nil {
			return err
		}
		return b.bindData(i, c.Request.PostForm, "form", "")
	}
	return nil
}

func (b *binder) bindData(dest interface{}, data map[string][]string, tag string, parentTagValue string) error {
	if dest == nil || len(data) == 0 {
		return nil
	}
	dtype := reflect.TypeOf(dest).Elem()
	dval := reflect.ValueOf(dest).Elem()
	if dtype.Kind() == reflect.Map {
		for k, v := range data {
			dval.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
		}
		return nil
	}
	if dtype.Kind() != reflect.Struct {
		return nil
	}

	//struct
	for i := 0; i < dtype.NumField(); i++ {
		field := dtype.Field(i)
		filedVal := dval.Field(i)
		if field.Anonymous {
			if filedVal.Kind() == reflect.Ptr {
				if filedVal.IsNil() {
					ptr := reflect.New(filedVal.Type().Elem())
					filedVal.Set(ptr)
				}
				if err := b.bindData(filedVal.Interface(), data, tag, ""); err != nil {
					return err
				}
			} else if filedVal.Kind() == reflect.Struct {
				if err := b.bindData(filedVal.Addr().Interface(), data, tag, ""); err != nil {
					return err
				}
			}
			continue
		}
		if !filedVal.CanSet() {
			continue
		}
		tagValue, ok := field.Tag.Lookup(tag)
		tagName := b.getTag(tagValue) //header,param,query,param
		if tagName == "-" {
			continue
		}
		var fullTagName string
		if tagName == "" {
			fullTagName = field.Name
		} else {
			fullTagName = tagName
		}
		if parentTagValue != "" {
			fullTagName = parentTagValue + "." + fullTagName
		}

		switch filedVal.Kind() {
		case reflect.Struct:
			if err := b.bindData(filedVal.Addr().Interface(), data, tag, fullTagName); err != nil {
				return err
			}
			continue
		case reflect.Ptr:
			if filedVal.IsNil() && fullTagName != "" {
				for k := range data {
					if strings.HasPrefix(strings.ToLower(k), strings.ToLower(fullTagName+".")) {
						ptr := reflect.New(filedVal.Type().Elem())
						filedVal.Set(ptr)
						break
					}
				}
			}
			if !filedVal.IsNil() {
				if err := b.bindData(filedVal.Interface(), data, tag, fullTagName); err != nil {
					return err
				}
			}
			continue
		case reflect.Slice:
			val, exists := b.findIgnoreCaseData(data, fullTagName)
			if !exists || len(val) == 0 {
				continue
			}
			slice := reflect.MakeSlice(filedVal.Type(), len(val), len(val))
			for i, v := range val {
				if err := b.setField(slice.Index(i), field, v); err != nil {
					return err
				}
			}
			filedVal.Set(slice)
			continue
		}

		if !ok {
			continue
		}
		val, exists := b.findIgnoreCaseData(data, fullTagName)
		if !exists || len(val) == 0 {
			continue
		}
		err := b.setField(filedVal, field, val[0])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *binder) findIgnoreCaseData(data map[string][]string, key string) (val []string, exists bool) {
	val, exists = data[key]
	if !exists {
		for k, v := range data {
			if strings.EqualFold(k, key) { //ignore case
				exists = true
				val = v
				return
			}
		}
	}
	return
}

func (b *binder) bindDefault(dest interface{}, defaultTagName string) error {
	if dest == nil || len(defaultTagName) == 0 {
		return nil
	}
	dtype := reflect.TypeOf(dest).Elem()
	dval := reflect.ValueOf(dest).Elem()
	if dtype.Kind() != reflect.Struct {
		return nil
	}

	//struct
	for i := 0; i < dtype.NumField(); i++ {
		field := dtype.Field(i)
		filedVal := dval.Field(i)
		if field.Anonymous { //embedded
			if filedVal.Kind() == reflect.Ptr {
				if filedVal.IsNil() {
					ptr := reflect.New(filedVal.Type().Elem())
					filedVal.Set(ptr)
				}
				if err := b.bindDefault(filedVal.Interface(), defaultTagName); err != nil {
					return err
				}
			}
		}
		if !filedVal.CanSet() {
			continue
		}

		if filedVal.Kind() == reflect.Struct {
			if err := b.bindDefault(filedVal.Addr().Interface(), defaultTagName); err != nil {
				return err
			}
			continue
		}

		val, ok := field.Tag.Lookup(defaultTagName)
		defaultValue := b.getTag(val)

		switch filedVal.Kind() {
		case reflect.Struct:
			if err := b.bindDefault(filedVal.Addr().Interface(), defaultTagName); err != nil {
				return err
			}
			continue
		case reflect.Ptr:
			if filedVal.IsNil() && defaultValue != "" {
				ptr := reflect.New(filedVal.Type().Elem())
				filedVal.Set(ptr)
			}
			if !filedVal.IsNil() {
				kind := filedVal.Type().Elem().Kind()
				if kind == reflect.Struct {
					if err := b.bindDefault(filedVal.Interface(), defaultTagName); err != nil {
						return err
					}
				} else {
					if err := b.setField(filedVal.Elem(), field, defaultValue); err != nil {
						return err
					}
				}
				continue
			}
		case reflect.Slice:
			if val != "" {
				vals := strings.Split(val, ",")
				if len(vals) > 0 {
					slice := reflect.MakeSlice(filedVal.Type(), len(vals), len(vals))
					for i, v := range vals {
						if err := b.setField(slice.Index(i), field, strings.TrimSpace(v)); err != nil {
							return err
						}
					}
					filedVal.Set(slice)
				}
			}
			continue
		}
		if !ok {
			continue
		}
		err := b.setField(filedVal, field, defaultValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *binder) getTag(tag string) string {
	idx := strings.Index(tag, ",")
	if idx < 0 {
		return tag
	}
	return tag[:idx]
}

func (b *binder) setField(fieldVal reflect.Value, field reflect.StructField, val string) error {
	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(val)
	case reflect.Int:
		return setIntField(fieldVal, val, 0)
	case reflect.Int8:
		return setIntField(fieldVal, val, 8)
	case reflect.Int16:
		return setIntField(fieldVal, val, 16)
	case reflect.Int32:
		return setIntField(fieldVal, val, 32)
	case reflect.Int64:
		switch fieldVal.Interface().(type) {
		case time.Duration:
			return setTimeDuration(fieldVal, val)
		}
		return setIntField(fieldVal, val, 64)
	case reflect.Uint:
		return setUintField(fieldVal, val, 0)
	case reflect.Uint8:
		return setUintField(fieldVal, val, 8)
	case reflect.Uint16:
		return setUintField(fieldVal, val, 16)
	case reflect.Uint32:
		return setUintField(fieldVal, val, 32)
	case reflect.Uint64:
		return setUintField(fieldVal, val, 64)
	case reflect.Bool:
		return setBoolField(fieldVal, val)
	case reflect.Float32:
		return setFloatField(fieldVal, val, 32)
	case reflect.Float64:
		return setFloatField(fieldVal, val, 64)
	case reflect.Struct:
		switch fieldVal.Interface().(type) {
		case time.Time:
			return setTimeField(fieldVal, val, field)
		}
		return json.Unmarshal(stringToBytes(val), fieldVal.Addr().Interface())
	case reflect.Map:
		return json.Unmarshal(stringToBytes(val), fieldVal.Addr().Interface())
	}
	return nil
}

func setIntField(field reflect.Value, val string, bitSize int) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setTimeDuration(field reflect.Value, val string) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	field.Set(reflect.ValueOf(d))
	return nil
}

func setUintField(field reflect.Value, val string, bitSize int) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(field reflect.Value, val string) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(field reflect.Value, val string, bitSize int) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

func setTimeField(value reflect.Value, val string, structField reflect.StructField) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	switch tf := strings.ToLower(timeFormat); tf {
	case "unix", "unixnano":
		tv, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}

		d := time.Duration(1)
		if tf == "unixnano" {
			d = time.Second
		}

		t := time.Unix(tv/int64(d), tv%int64(d))
		value.Set(reflect.ValueOf(t))
		return nil
	}

	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := structField.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(t))
	return nil
}
