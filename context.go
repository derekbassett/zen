package zen

import (
	"encoding/asn1"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
)

const (
	inputTagName = "form"
	validTagName = "valid"
	validMsgName = "msg"
)

//commonly used mime-types
const (
	applicationJSON = "application/json"
	applicationXML  = "application/xml"
	textXML         = "text/xml"
	applicationASN1 = "application/asn1"

	contentType = "Content-Type"
)

type (
	// Context warps request and response writer
	Context struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
)

func (s *Server) getContext(rw http.ResponseWriter, req *http.Request) *Context {
	c := s.contextPool.Get().(*Context)
	c.Req = req
	c.rw.writer = rw
	return c
}

func (s *Server) putBackContext(c *Context) {
	c.params = c.params[0:0]
	c.parsed = false
	c.Req = nil
	c.rw.writer = nil

	s.contextPool.Put(c)
}

// parseInput will parse request's form and
func (c *Context) parseInput() error {
	return c.Req.ParseForm()
}

// Form return request form value with given key
func (c *Context) Form(key string) string {
	if !c.parsed {
		c.parseInput()
	}
	return c.Req.FormValue(key)
}

// Param return url param with given key
func (c *Context) Param(key string) string {
	return c.params.ByName(key)
}

// ParseValidateForm will parse request's form and map into a interface{} value
func (c *Context) ParseValidateForm(input interface{}) error {
	return c.parseValidateForm(input)
}

// BindJSON will parse request's json body and map into a interface{} value
func (c *Context) BindJSON(input interface{}) error {
	if err := json.NewDecoder(c.Req.Body).Decode(input); err != nil {
		return err
	}
	return nil
}

// BindXML will parse request's xml body and map into a interface{} value
func (c *Context) BindXML(input interface{}) error {
	if err := xml.NewDecoder(c.Req.Body).Decode(input); err != nil {
		return err
	}
	return nil
}

func (c *Context) parseValidateForm(input interface{}) error {
	inputValue := reflect.ValueOf(input).Elem()
	inputType := inputValue.Type()

	for i := 0; i < inputValue.NumField(); i++ {
		tag := inputType.Field(i).Tag
		formName := tag.Get(inputTagName)
		validate := tag.Get(validTagName)
		validateMsg := tag.Get(validMsgName)
		field := inputValue.Field(i)
		formValue := c.Req.Form.Get(formName)

		// scan form string value into field
		if err := scan(field, formValue); err != nil {
			return err
		}
		// validate form with regex
		if err := valid(formValue, validate, validateMsg); err != nil {
			return err
		}

	}
	return nil
}

func scan(v reflect.Value, s string) error {

	if !v.CanSet() {
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(s)

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(x)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)

	}
	return nil
}

func valid(s string, validate, msg string) error {
	if validate == "" {
		return nil
	}
	rxp, err := regexp.Compile(validate)
	if err != nil {
		return err
	}

	if !rxp.MatchString(s) {
		return errors.New(msg)
	}

	return nil
}

// JSON : write json data to http response writer, with status code 200
func (c *Context) JSON(i interface{}) (err error) {
	// write http status code
	c.WriteHeader(contentType, applicationJSON)

	// Encode json data to rw
	err = json.NewEncoder(c.rw).Encode(i)

	//return
	return
}

// XML : write xml data to http response writer, with status code 200
func (c *Context) XML(i interface{}) (err error) {
	// write http status code
	c.WriteHeader(contentType, applicationXML)

	// Encode xml data to rw
	err = xml.NewEncoder(c.rw).Encode(i)

	//return
	return
}

// ASN1 : write asn1 data to http response writer, with status code 200
func (c *Context) ASN1(i interface{}) (err error) {
	// write http status code
	c.WriteHeader(contentType, applicationASN1)

	// Encode asn1 data to rw
	bts, err := asn1.Marshal(i)
	if err != nil {
		return
	}
	//return
	_, err = c.rw.Write(bts)
	return
}

// WriteStatus set response's status code
func (c *Context) WriteStatus(code int) {
	c.rw.WriteHeader(code)
}

// WriteHeader set response header
func (c *Context) WriteHeader(k, v string) {
	c.rw.Header().Add(k, v)
}

// RawStr write raw string
func (c *Context) RawStr(s string) {
	io.WriteString(c.rw, s)
}

// File serve file
func (c *Context) File(filepath string) {
	http.ServeFile(c.rw, c.Req, filepath)
}

// Data writes some data into the body stream and updates the HTTP code.
func (c *Context) Data(cType string, data []byte) {
	c.WriteHeader(contentType, cType)
	c.rw.Write(data)
}
