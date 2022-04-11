package bytego

import (
	"encoding/json"
	"encoding/xml"
	"log"
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

type binder struct{}

func (b *binder) Bind(c *Ctx, i interface{}) error {
	if err := b.bindParams(c, i); err != nil {
		return err
	}
	if err := b.bindQueries(c, i); err != nil {
		return err
	}
	if err := b.bindHeaders(c, i); err != nil {
		return err
	}
	return b.bindBody(c, i)
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
		return b.bindData(i, c.Request.Form, "form", "")
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
		filed := dtype.Field(i)
		filedVal := dval.Field(i)
		log.Println(filed.Name, filedVal.CanSet(), parentTagValue)

		if !filedVal.CanSet() {
			continue
		}
		tagName := b.getTag(filed.Tag.Get(tag))
		if tagName == "-" {
			continue
		}
		if tagName == "" && !filed.Anonymous {
			tagName = filed.Name
		}
		if parentTagValue != "" {
			tagName = parentTagValue + "." + tagName
		}
		val, exists := data[tagName]
		if !exists {
			for k, v := range data {
				if strings.EqualFold(k, tagName) {
					exists = true
					val = v
				}
			}
		}
		if filedVal.Kind() == reflect.Struct {
			if err := b.bindData(filedVal.Addr().Interface(), data, tag, tagName); err != nil {
				return err
			}
		}
		if !exists {
			continue
		}
		err := b.setField(filedVal, filed, val[0])
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
