package zen

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	inputTagName = "form"
	validTagName = "valid"
	validMsgName = "msg"
)

// Headers
const (
	HeaderAcceptEncoding                = "Accept-Encoding"
	HeaderAllow                         = "Allow"
	HeaderAuthorization                 = "Authorization"
	HeaderContentDisposition            = "Content-Disposition"
	HeaderContentEncoding               = "Content-Encoding"
	HeaderContentLength                 = "Content-Length"
	HeaderContentType                   = "Content-Type"
	HeaderCookie                        = "Cookie"
	HeaderSetCookie                     = "Set-Cookie"
	HeaderIfModifiedSince               = "If-Modified-Since"
	HeaderLastModified                  = "Last-Modified"
	HeaderLocation                      = "Location"
	HeaderUpgrade                       = "Upgrade"
	HeaderVary                          = "Vary"
	HeaderWWWAuthenticate               = "WWW-Authenticate"
	HeaderXForwardedProto               = "X-Forwarded-Proto"
	HeaderXHTTPMethodOverride           = "X-HTTP-Method-Override"
	HeaderXForwardedFor                 = "X-Forwarded-For"
	HeaderXRealIP                       = "X-Real-IP"
	HeaderServer                        = "Server"
	HeaderOrigin                        = "Origin"
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)

type (
	// Context warps request and response writer
	Context struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
		context.Context
	}
)

func (s *Server) getContext(rw http.ResponseWriter, req *http.Request) *Context {
	c := contextPool.Get().(*Context)
	c.Req = req
	c.rw.writer = rw
	c.Context = context.Background()
	c.SetValue(fieldKey{}, fields{})
	return c
}

func (s *Server) putBackContext(c *Context) {
	c.params = c.params[0:0]
	c.parsed = false
	c.rw.written = false
	c.Req = nil
	c.rw.writer = nil
	c.Context = nil
	contextPool.Put(c)
}

// parseInput will parse request's form and
func (c *Context) parseInput() error {
	c.parsed = true
	return c.Req.ParseForm()
}

// Dup make a duplicate Context with context.Context
func (c *Context) Dup(ctx context.Context) *Context {
	ret := new(Context)
	ret.Req = c.Req
	ret.rw = c.rw
	ret.Context = ctx
	ret.parsed = c.parsed
	ret.params = c.params
	return ret
}

// WithDeadline ...
func (c *Context) WithDeadline(dead time.Time) (*Context, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(c, dead)
	return c.Dup(ctx), cancel
}

// WithCancel ...
func (c *Context) WithCancel() (*Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c)
	return c.Dup(ctx), cancel
}

// Do job with context
func (c *Context) Do(job func() error) error {
	errChan := make(chan error)
	done := make(chan struct{})
	go func() {
		if err := job(); err != nil {
			errChan <- err
			return
		}
		done <- struct{}{}
	}()

	select {
	case <-c.Done():
		return c.Err()
	case err := <-errChan:
		return err
	case <-done:
		return nil
	}
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
	if !c.parsed {
		c.parseInput()
	}
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
	c.WriteHeader(HeaderContentType, MIMEApplicationJSONCharsetUTF8)

	// Encode json data to rw
	return json.NewEncoder(c.rw).Encode(i)
}

// XML : write xml data to http response writer, with status code 200
func (c *Context) XML(i interface{}) (err error) {
	// write http status code
	c.WriteHeader(HeaderContentType, MIMEApplicationXMLCharsetUTF8)

	// Encode xml data to rw
	return xml.NewEncoder(c.rw).Encode(i)
}

// WriteStatus set response's status code
func (c *Context) WriteStatus(code int) {
	c.rw.WriteHeader(code)
}

// WriteHeader set response header
func (c *Context) WriteHeader(k, v string) {
	c.rw.Header().Add(k, v)
}

// WriteString write raw string
func (c *Context) WriteString(s string) {
	io.WriteString(c.rw, s)
}

// WriteFile serve file
func (c *Context) WriteFile(filepath string) {
	http.ServeFile(c.rw, c.Req, filepath)
}

// WriteData writes some data into the body stream and updates the HTTP code.
func (c *Context) WriteData(contentType string, data []byte) {
	c.WriteHeader(HeaderContentType, contentType)
	c.rw.Write(data)
}

// SetValue set value on context
func (c *Context) SetValue(key, val interface{}) {
	c.Context = context.WithValue(c.Context, key, val)
}

// GetValue of key
func (c *Context) GetValue(key interface{}) interface{} {
	return c.Value(key)
}

// SetField set key val on context fields
func (c *Context) SetField(key string, val interface{}) {
	f, _ := c.Value(fieldKey{}).(fields)
	n := fields{key: val}
	n.Merge(f)
	c.SetValue(fieldKey{}, n)
}

// LogError print error level log with fields
func (c *Context) LogError(args ...interface{}) {
	log.WithFields(log.Fields(c.stackField(2))).Error(args...)
}

// LogErrorf print error level format log with fields
func (c *Context) LogErrorf(format string, args ...interface{}) {
	log.WithFields(log.Fields(c.stackField(2))).Errorf(format, args...)
}

// LogInfo print info level log with fields
func (c *Context) LogInfo(args ...interface{}) {
	log.WithFields(log.Fields(c.stackField(2))).Info(args...)
}

// LogInfof print info level format log with fields
func (c *Context) LogInfof(format string, args ...interface{}) {
	log.WithFields(log.Fields(c.stackField(2))).Infof(format, args...)
}

func (c *Context) stackField(depth int) fields {
	_, caller, line, _ := runtime.Caller(depth)
	stack := fields{
		"caller": fmt.Sprintf("%s:%d", caller, line),
	}
	stack.Merge(c.fields())
	return stack
}

func (c *Context) fields() fields {
	return c.Value(fieldKey{}).(fields)
}
