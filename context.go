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
		Rw     http.ResponseWriter
		params Params
		parsed bool
		context.Context
	}
)

func getContext(rw http.ResponseWriter, req *http.Request) *Context {
	c := contextPool.Get().(*Context)
	c.Req = req
	c.Rw = rw
	c.Context = context.TODO()
	c.SetValue(fieldKey{}, fields{})
	return c
}

func putBackContext(ctx *Context) {
	ctx.params = ctx.params[0:0]
	ctx.parsed = false
	ctx.Req = nil
	ctx.Rw = nil
	ctx.Context = nil
	contextPool.Put(ctx)
}

// parseInput will parse request's form and
func (ctx *Context) parseInput() error {
	ctx.parsed = true
	return ctx.Req.ParseForm()
}

// Dup make a duplicate Context with context.Context
func (ctx *Context) Dup(c context.Context) *Context {
	ret := new(Context)
	ret.Req = ctx.Req
	ret.Rw = ctx.Rw
	ret.Context = c
	ret.parsed = ctx.parsed
	ret.params = ctx.params
	return ret
}

// WithDeadline ...
func (ctx *Context) WithDeadline(dead time.Time) (*Context, context.CancelFunc) {
	c, cancel := context.WithDeadline(ctx, dead)
	return ctx.Dup(c), cancel
}

// WithCancel ...
func (ctx *Context) WithCancel() (*Context, context.CancelFunc) {
	c, cancel := context.WithCancel(ctx)
	return ctx.Dup(c), cancel
}

// Do job with context
func (ctx *Context) Do(job func() error) error {
	errChan := make(chan error)
	done := make(chan struct{})
	go func() {
		if err := job(); err != nil {
			errChan <- err
			return
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	case <-done:
		return nil
	}
}

// Form return request form value with given key
func (ctx *Context) Form(key string) string {
	if !ctx.parsed {
		ctx.parseInput()
	}
	return ctx.Req.FormValue(key)
}

// Param return url param with given key
func (ctx *Context) Param(key string) string {
	return ctx.params.ByName(key)
}

// ParseValidateForm will parse request's form and map into a interface{} value
func (ctx *Context) ParseValidateForm(input interface{}) error {
	if !ctx.parsed {
		ctx.parseInput()
	}
	return ctx.parseValidateForm(input)
}

// BindJSON will parse request's json body and map into a interface{} value
func (ctx *Context) BindJSON(input interface{}) error {
	if err := json.NewDecoder(ctx.Req.Body).Decode(input); err != nil {
		return err
	}
	return nil
}

// BindXML will parse request's xml body and map into a interface{} value
func (ctx *Context) BindXML(input interface{}) error {
	if err := xml.NewDecoder(ctx.Req.Body).Decode(input); err != nil {
		return err
	}
	return nil
}

func (ctx *Context) parseValidateForm(input interface{}) error {
	inputValue := reflect.ValueOf(input).Elem()
	inputType := inputValue.Type()

	for i := 0; i < inputValue.NumField(); i++ {
		tag := inputType.Field(i).Tag
		formName := tag.Get(inputTagName)
		validate := tag.Get(validTagName)
		validateMsg := tag.Get(validMsgName)
		field := inputValue.Field(i)
		formValue := ctx.Req.Form.Get(formName)

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
func (ctx *Context) JSON(i interface{}) (err error) {
	// write http status code
	ctx.WriteHeader(HeaderContentType, MIMEApplicationJSONCharsetUTF8)

	// Encode json data to rw
	return json.NewEncoder(ctx.Rw).Encode(i)
}

// XML : write xml data to http response writer, with status code 200
func (ctx *Context) XML(i interface{}) (err error) {
	// write http status code
	ctx.WriteHeader(HeaderContentType, MIMEApplicationXMLCharsetUTF8)

	// Encode xml data to rw
	return xml.NewEncoder(ctx.Rw).Encode(i)
}

// WriteStatus set response's status code
func (ctx *Context) WriteStatus(code int) {
	ctx.Rw.WriteHeader(code)
}

// WriteHeader set response header
func (ctx *Context) WriteHeader(k, v string) {
	ctx.Rw.Header().Add(k, v)
}

// WriteString write raw string
func (ctx *Context) WriteString(s string) {
	io.WriteString(ctx.Rw, s)
}

// WriteFile serve file
func (ctx *Context) WriteFile(filepath string) {
	http.ServeFile(ctx.Rw, ctx.Req, filepath)
}

// WriteData writes some data into the body stream and updates the HTTP code.
func (ctx *Context) WriteData(contentType string, data []byte) {
	ctx.WriteHeader(HeaderContentType, contentType)
	ctx.Rw.Write(data)
}

// SetValue set value on context
func (ctx *Context) SetValue(key, val interface{}) {
	ctx.Context = context.WithValue(ctx.Context, key, val)
}

// GetValue of key
func (ctx *Context) GetValue(key interface{}) interface{} {
	return ctx.Value(key)
}

// SetField set key val on context fields
func (ctx *Context) SetField(key string, val interface{}) {
	f, _ := ctx.Value(fieldKey{}).(fields)
	n := fields{key: val}
	n.Merge(f)
	ctx.SetValue(fieldKey{}, n)
}

// LogError print error level log with fields
func (ctx *Context) LogError(args ...interface{}) {
	log.WithFields(log.Fields(ctx.stackField(2))).Error(args...)
}

// LogErrorf print error level format log with fields
func (ctx *Context) LogErrorf(format string, args ...interface{}) {
	log.WithFields(log.Fields(ctx.stackField(2))).Errorf(format, args...)
}

// LogInfo print info level log with fields
func (ctx *Context) LogInfo(args ...interface{}) {
	log.WithFields(log.Fields(ctx.stackField(2))).Info(args...)
}

// LogInfof print info level format log with fields
func (ctx *Context) LogInfof(format string, args ...interface{}) {
	log.WithFields(log.Fields(ctx.stackField(2))).Infof(format, args...)
}

func (ctx *Context) stackField(depth int) fields {
	_, caller, line, _ := runtime.Caller(depth)
	stack := fields{
		"caller": fmt.Sprintf("%s:%d", caller, line),
	}
	stack.Merge(ctx.fields())
	return stack
}

func (ctx *Context) fields() fields {
	return ctx.Value(fieldKey{}).(fields)
}
