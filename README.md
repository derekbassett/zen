# zen is a lightweight go framework for web development [![CircleCI](https://circleci.com/gh/philchia/zen/tree/master.svg?style=svg)](https://circleci.com/gh/philchia/zen/tree/master)

[![Coverage Status](https://coveralls.io/repos/github/philchia/zen/badge.svg?branch=master)](https://coveralls.io/github/philchia/zen?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/philchia/zen)](https://goreportcard.com/report/github.com/philchia/zen)
[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![GoDoc](https://godoc.org/github.com/philchia/zen?status.svg)](https://godoc.org/github.com/philchia/zen)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://opensource.org/licenses/MIT)

zen is a web framework written by go, you will love it if you preffer high performance and lightweight!!!

## Features

* High performance HTTP router
* Restful API
* Parameters in path
* Group APIs
* Form validate and struct bind
* JSON and XML data bind
* Built in panic handler
* Middleware at root or group level
* Handy response functions

## Installation

```bash
go get github.com/philchia/zen
```

## Examples

### Start a server

```go
func main() {
    server := zen.New()

    if err := server.Run(":8080"); err != nil {
        log.Println(err)
    }
}
```

### Using GET, POST, PUT, PATCH, DELETE

```go
    server := zen.New()
    server.Get("/test",handler)
    server.Post("/test", handler)
    server.Put("/test",handler)
    server.Patch("/test", handler)
    server.Delete("/test",handler)
```

### Group route

```go
    server := zen.New()

    user := server.Group("/user")
    {
        user.Get("/test",handler)
        user.Post("/test", handler)
        user.Put("/test",handler)
        user.Patch("/test", handler)
        user.Delete("/test",handler)
    }
```

### Add a middleware

```go
    server := zen.New()
    server.AddInterceptor(func(c *zen.Context) {
        log.Println("root middleware")
    })
```

### Group layer middleware

```go
    server := zen.New()

    user := server.Group("/user")
    {
        user.AddInterceptor(func(c *zen.Context) {
        log.Println("user middleware")
        })
    }
```

### Parameters in path

```go
    server := zen.New()
    server.Get("/user/:uid",func (c *zen.Context) {
        c.JSON(map[string]string{"uid": c.Param("uid")})
    })
    if err := server.Run(":8080"); err != nil {
    log.Println(err)
    }
```

### Parse and validate input

```go
func handler(c *zen.Context) {
    var input struct {
        Name string `form:"name" json:"name"`
        Age  int    `form:"age" json:"age"`
        Mail string `form:"mail" valid:"[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,64}" msg:"Illegal email" json:"mail"`
    }

    if err := c.ParseValidForm(&input); err != nil {
        c.JSON(map[string]string{"err": err.Error()})
        return
    }
    log.Println(input)
    c.JSON(input)
}
```

### Handle panic

```go
    server := zen.New()
    server.HandlePanic(func(c *zen.Context, err interface{}) {
        c.RawStr(fmt.Sprint(err))
    })
    if err := server.Run(":8080"); err != nil {
    log.Println(err)
    }
```

### Handle 404

```go
    server := zen.New()
    server.HandleNotFound(func(c *zen.Context) {
        c.WriteStatus(StatusNotFound)
        c.RawStr(StatusText(StatusNotFound))
    })
    if err := server.Run(":8080"); err != nil {
    log.Println(err)
    }
```

### Context support

```go
    server := zen.New()
    server.HandleNotFound(func(c *zen.Context) {
        db, _ := sql.Open("mysql", "dsn")
        db.QueryContext(c, "SELECT * FROM table;")
    })
    if err := server.Run(":8080"); err != nil {
    log.Println(err)
    }
```

### Shutdown

```go
    server := zen.New()
    server.Shutdown()
```

## License

zen is published under MIT license