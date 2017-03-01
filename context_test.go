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
	}{
		{"case1",
			fields{
				mustNewRequest("GET", "/GET?email=golang@gmail.com", nil),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				Params{Param{Key: "uid", Value: "10086"}},
				false,
			},
			args{
				&struct {
					Email string `form:"email" valid:"[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,64}" msg:"Illegal email"`
				}{},
			},
			false,
		},

		{"case2",
			fields{
				mustNewRequest("GET", "/GET?name=zen", nil),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				Params{Param{Key: "uid", Value: "10086"}},
				true,
			},
			args{
				&struct {
					Name string `form:"name" valid:"[a-zA-Z0-9]{6}" msg:"Name should have len 6"`
				}{},
			},
			true,
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
			if err := c.ParseValidateForm(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Context.ParseValidateForm() error = %v, wantErr %v", err, tt.wantErr)
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
		{"case1",
			fields{
				mustNewRequest("GET", "/GET", strings.NewReader(`{"name":"zen"}`)),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				Params{Param{Key: "uid", Value: "10086"}},
				true,
			},
			args{
				&struct {
					Name string `json:"name"`
				}{},
			},
			false,
		},

		{"case2",
			fields{
				mustNewRequest("GET", "/GET", strings.NewReader(`{"flag":"zen"}`)),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				Params{Param{Key: "uid", Value: "10086"}},
				true,
			},
			args{
				&struct {
					Flag bool `json:"flag"`
				}{},
			},
			true,
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
		{"case1",
			fields{
				mustNewRequest("GET", "/GET", strings.NewReader(`<x><name>hello</name></x>`)),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				Params{Param{Key: "uid", Value: "10086"}},
				true,
			},
			args{
				&struct {
					Name string `xml:"name"`
				}{},
			},
			false,
		},

		{"case2",
			fields{
				mustNewRequest("GET", "/GET", strings.NewReader(`{"flag":"zen"}`)),
				&responseWriter{writer: new(mockResponseWriter), written: false},
				Params{Param{Key: "uid", Value: "10086"}},
				true,
			},
			args{
				&struct {
					Flag bool `xml:"flag"`
				}{},
			},
			true,
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

	type tst struct {
		Name   string
		Age    int
		Gender bool
		Level  uint
		Money  float32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"case1",
			args{
				reflect.ValueOf(new(tst)).Elem().Field(1),
				"6",
			},
			false,
		},

		{
			"case2",
			args{
				reflect.ValueOf(new(tst)).Elem().Field(1),
				"a",
			},
			true,
		},

		{
			"case3",
			args{
				reflect.ValueOf(new(tst)).Elem().Field(0),
				"zen",
			},
			false,
		},
		{
			"case4",
			args{
				reflect.ValueOf(new(tst)).Elem().Field(2),
				"true",
			},
			false,
		},
		{
			"case5",
			args{
				reflect.ValueOf(new(tst)).Elem().Field(2),
				"null",
			},
			true,
		},
		{
			"case6",
			args{
				reflect.ValueOf(new(tst)).Elem().Field(3),
				"4",
			},
			false,
		},
		{
			"case7",
			args{
				reflect.ValueOf(new(tst)).Elem().Field(3),
				"-4",
			},
			true,
		},
		{
			"case8",
			args{
				reflect.ValueOf(new(tst)).Elem().Field(4),
				"3.1415",
			},
			false,
		},
		{
			"case9",
			args{
				reflect.ValueOf(new(tst)).Elem().Field(4),
				"-a3.1415",
			},
			true,
		},
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
		params Params
		parsed bool
	}
	type args struct {
		i interface{}
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantBody string
	}{
		{"case1",
			fields{
				mustNewRequest("GET", "/GET", strings.NewReader("name=zen&version=1.0")),
				nil,
				false,
			},
			args{
				map[string]string{"name": "zen"},
			},
			false,
			"{\"name\":\"zen\"}\n",
		},
	}
	for _, tt := range tests {
		rw := new(mockResponseWriter)

		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     &responseWriter{writer: rw, written: false},
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if err := c.JSON(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("Context.JSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if rw.body.String() != tt.wantBody {
				t.Errorf("Context.JSON() body = %s, want %s", rw.body.String(), tt.wantBody)
			}
		})
	}
}

func TestContext_XML(t *testing.T) {
	type fields struct {
		Req    *http.Request
		params Params
		parsed bool
	}
	type x struct {
		Name string `xml:"name"`
	}

	type args struct {
		i interface{}
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantBody string
	}{
		{"case1",
			fields{
				mustNewRequest("GET", "/GET", strings.NewReader("name=zen&version=1.0")),
				nil,
				false,
			},
			args{
				x{"zen"},
			},
			false,
			"<x><name>zen</name></x>",
		},
	}
	for _, tt := range tests {
		rw := new(mockResponseWriter)

		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     &responseWriter{writer: rw, written: false},
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			if err := c.XML(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("Context.XML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if rw.body.String() != tt.wantBody {
				t.Errorf("Context.XML() body = %s, want %s", rw.body.String(), tt.wantBody)
			}
		})
	}
}

func TestContext_WriteStatus(t *testing.T) {
	type fields struct {
		Req    *http.Request
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
		{"case1",
			fields{
				mustNewRequest("GET", "/GET", strings.NewReader("name=zen&version=1.0")),
				nil,
				false,
			},
			args{
				404,
			},
		},
		{"case2",
			fields{
				mustNewRequest("GET", "/GET", strings.NewReader("name=zen&version=1.0")),
				nil,
				false,
			},
			args{
				302,
			},
		},
	}
	for _, tt := range tests {
		rw := new(mockResponseWriter)
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Req:    tt.fields.Req,
				rw:     &responseWriter{writer: rw, written: false},
				params: tt.fields.params,
				parsed: tt.fields.parsed,
			}
			c.WriteStatus(tt.args.code)
			if rw.code != tt.args.code {
				t.Errorf("Context.WriteStatus() code = %d, want %d", rw.code, tt.args.code)
			}
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
