package zen

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestServer_getContext(t *testing.T) {
	type fields struct {
		notFoundHandler HandlerFunc
		panicHandler    PanicHandler
		filters         []HandlerFunc
		contextPool     *sync.Pool
	}
	type args struct {
		rw  http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantNil bool
	}{
		{"case1",
			fields{
				contextPool: &sync.Pool{
					New: func() interface{} {
						c := Context{
							params: Params{},
							rw:     &responseWriter{},
						}
						return &c
					},
				},
			},
			args{
				nil, nil,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				notFoundHandler: tt.fields.notFoundHandler,
				panicHandler:    tt.fields.panicHandler,
				contextPool:     *tt.fields.contextPool,
			}
			if got := s.getContext(tt.args.rw, tt.args.req); (got == nil) != tt.wantNil {
				t.Errorf("Server.getContext() = %v, want nil? %v", got, tt.wantNil)
			} else {
				s.putBackContext(got)
			}
		})
	}
}

func BenchmarkGetContext(b *testing.B) {
	s := &Server{

		contextPool: sync.Pool{
			New: func() interface{} {
				c := Context{
					params: Params{},
					rw:     &responseWriter{},
				}
				return &c
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := s.getContext(nil, nil)
		s.putBackContext(c)
	}
}

func mustNewRequest(method string, urlStr string, body io.Reader) *http.Request {
	ret, _ := http.NewRequest(method, urlStr, body)
	return ret
}

func TestContext_parseInput(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"case1",
			fields{
				mustNewRequest("GET", "/GET", strings.NewReader("name=zen&version=1.0")),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				nil,
				false,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if err := c.parseInput(); (err != nil) != tt.wantErr {
				t.Errorf("Context.parseInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContext_Form(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"case1",
			fields{
				mustNewRequest("GET", "/GET?name=zen", nil),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				nil,
				false,
			},
			args{"name"},
			"zen",
		},
		{"case2",
			fields{
				mustNewRequest("GET", "/GET?name=zen", nil),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				nil,
				false,
			},
			args{"age"},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if got := c.Form(tt.args.key); got != tt.want {
				t.Errorf("Context.Form() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContext_Param(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"case1",
			fields{
				mustNewRequest("GET", "/GET?name=zen", nil),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				Params{Param{Key: "uid", Value: "10086"}},
				false,
			},
			args{"uid"},
			"10086",
		},
		{"case1",
			fields{
				mustNewRequest("GET", "/GET?name=zen", nil),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				Params{Param{Key: "uid", Value: "10086"}},
				false,
			},
			args{"name"},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if got := c.Param(tt.args.key); got != tt.want {
				t.Errorf("Context.Param() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContext_ParseValidateForm(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		input interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if err := c.ParseValidateForm(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Context.ParseValidateForm() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContext_parseValidateForm(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		input interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if err := c.parseValidateForm(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Context.parseValidateForm() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContext_BindJSON(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		input interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if err := c.BindJSON(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Context.BindJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContext_BindXML(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		input interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if err := c.BindXML(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Context.BindXML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_scan(t *testing.T) {
	type args struct {
		v reflect.Value
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := scan(tt.args.v, tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_valid(t *testing.T) {
	type args struct {
		s        string
		validate string
		msg      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := valid(tt.args.s, tt.args.validate, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("valid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContext_JSON(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if err := c.JSON(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("Context.JSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContext_XML(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if err := c.XML(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("Context.XML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContext_ASN1(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if err := c.ASN1(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("Context.ASN1() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContext_WriteStatus(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		code int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			c.WriteStatus(tt.args.code)
		})
	}
}

func TestContext_RawStr(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			c.RawStr(tt.args.s)
		})
	}
}

func TestContext_Data(t *testing.T) {
	type fields struct {
		Req    *http.Request
		rw     *responseWriter
		params Params
		parsed bool
	}
	type args struct {
		cType string
		data  []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     tt.fields.rw,
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			c.Data(tt.args.cType, tt.args.data)
		})
	}
}
