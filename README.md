# Goh
[![GoDoc](https://pkg.go.dev/badge/github.com/OblivionOcean/Goh)](https://pkg.go.dev/github.com/OblivionOcean/Goh)
[![Go Report Card](https://goreportcard.com/badge/github.com/OblivionOcean/Goh)](https://goreportcard.com/report/github.com/OblivionOcean/Goh)

Goh is a pre-compiled and fast template engine for the Go language.

English | [简体中文](https://github.com/OblivionOcean/Goh/blob/master/README_zh-CN.md)

## Table of Contents
- [Features](#features)
- [Performance Tests](#performance-tests)
- [Installation](#installation)
- [Usage](#usage)
- [Syntax](#syntax)

## Features
- [x] Pre-compiled template engine to improve running speed.
- [x] Almost compatible with the syntax of the Go language.
- [x] Zero-dependency.
- [ ] Automatically re-compile after changing the template file.

## Performance Tests
Obtained from https://github.com/slinso/goTemplateBenchmark. They are local test results generated from the test code in Hero.

```txt
goos: windows
goarch: amd64
pkg: github.com/SlinSo/goTemplateBenchmark
cpu: Intel(R) Core(TM) i7-10700 CPU @ 2.90GHz
# Simple Template Tests
BenchmarkComplexGolang-16                  36800             31428 ns/op            6562 B/op        290 allocs/op
BenchmarkComplexGolangText-16              88148             13370 ns/op            2235 B/op        107 allocs/op
BenchmarkComplexEgo-16                    486294              2411 ns/op             568 B/op         31 allocs/op
BenchmarkComplexQuicktemplate-16         1367928               878.1 ns/op             0 B/op          0 allocs/op
BenchmarkComplexTempl-16                  788673              1400 ns/op             408 B/op         11 allocs/op
BenchmarkComplexFtmpl-16                  293755              3982 ns/op            3534 B/op         38 allocs/op
BenchmarkComplexFtmplInclude-16           317361              4142 ns/op            3534 B/op         38 allocs/op
BenchmarkComplexMustache-16                90567             13748 ns/op            7274 B/op        156 allocs/op
BenchmarkComplexGorazor-16                361304              3195 ns/op            3688 B/op         24 allocs/op
BenchmarkComplexJetHTML-16                189176              5928 ns/op             532 B/op          5 allocs/op
BenchmarkComplexHero-16                  1410391               863.5 ns/op             0 B/op          0 allocs/op
BenchmarkComplexGoh-16                   2304783               535.4 ns/op             0 B/op          0 allocs/op
BenchmarkComplexJade-16                  1826784               651.8 ns/op             0 B/op          0 allocs/op
BenchmarkComplexGoDirectBuffer-16        2890996               414.6 ns/op             0 B/op          0 allocs/op
BenchmarkComplexGoHyperscript-16         1717754               778.6 ns/op             0 B/op          0 allocs/op
BenchmarkComplexGoStaticString-16       84003024                14.44 ns/op            0 B/op          0 allocs/op
# Simple Template Tests
BenchmarkGolang-16                        300493              3691 ns/op             768 B/op         35 allocs/op
BenchmarkGolangText-16                   1000000              1073 ns/op             128 B/op          7 allocs/op
BenchmarkGoDirectBuffer-16              21959280                55.81 ns/op            0 B/op          0 allocs/op
BenchmarkGoCustomHtmlAPI-16             14034298                85.06 ns/op            0 B/op          0 allocs/op
BenchmarkGoFunc3-16                     14962965                68.62 ns/op            0 B/op          0 allocs/op
BenchmarkEgo-16                          2577276               464.3 ns/op            85 B/op          8 allocs/op
BenchmarkHB-16                            280617              4445 ns/op            2448 B/op         51 allocs/op
BenchmarkQuicktemplate-16                7013572               168.9 ns/op             0 B/op          0 allocs/op
BenchmarkFtmpl-16                        1000000              1000 ns/op             774 B/op         12 allocs/op
BenchmarkAce-16                           179811              6605 ns/op            1121 B/op         40 allocs/op
BenchmarkAmber-16                         268149              3800 ns/op             849 B/op         36 allocs/op
BenchmarkMustache-16                      523143              2636 ns/op            1722 B/op         30 allocs/op
BenchmarkPongo2-16                        350612              3862 ns/op            2074 B/op         32 allocs/op
BenchmarkHandlebars-16                    162860              7261 ns/op            3423 B/op         75 allocs/op
BenchmarkGorazor-16                      1562088               772.3 ns/op           512 B/op          5 allocs/op
BenchmarkSoy-16                           639549              2200 ns/op            1224 B/op         19 allocs/op
BenchmarkJetHTML-16                      1960117               600.4 ns/op             0 B/op          0 allocs/op
BenchmarkHero-16                        10452396               113.9 ns/op             0 B/op          0 allocs/op
BenchmarkGoh-16                         14838537                81.97 ns/op            0 B/op          0 allocs/op
BenchmarkJade-16                        15025261                78.85 ns/op            0 B/op          0 allocs/op
BenchmarkTempl-16                        4015622               293.1 ns/op            96 B/op          2 allocs/op
BenchmarkGomponents-16                    479330              2882 ns/op            1112 B/op         56 allocs/op
ok      github.com/SlinSo/goTemplateBenchmark   65.553s
```

## Installation
```shell
go get -u github.com/OblivionOcean/Goh
go install github.com/OblivionOcean/Goh

# Dependencies
go get golang.org/x/tools/cmd/goimports
go install golang.org/x/tools/cmd/goimports
```
## Usage
```shell
~ $ Goh
Usage of./Goh:
  -dest string
        generated golang files dir, it will be the same as source if not set
  -ext string
        source file extensions, comma-separated if many (default ".html")
  -pkg string
        the generated template package name, default is template (default "template")
  -src string
        the html template file or directory (default "./")
```
> For the complete usage method, please refer to [example programs](https://github.com/OblivionOcean/Goh/tree/master/example)


```html
<%: func UserList(title string, userList []string, buf *bytes.Buffer) %>
    <!DOCTYPE html>
    <html>

    <head>
        <title>
            <%= title %>
        </title>
    </head>

    <body>
        <h1>
            <%= title %>
        </h1>
        <ul>
            <% for _, user :=range userList { %>
                <% if user !="Alice" { %>
                    <li>
                        <%= user %>
                    </li>
                    <% } %>
                        <% } %>
        </ul>
    </body>

    </html>
```

```go
package main

import (
	"bytes"
	"net/http"

	"github.com/OblivionOcean/Goh/example/template"
)

func main() {
	http.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
		var userList = []string{
			"Alice",
			"Bob",
			"Tom",
		}

		buffer := new(bytes.Buffer)
		template.UserList("User List", userList, buffer)

		w.Write(buffer.Bytes())
	})

	http.ListenAndServe(":8080", nil)
}

```
## Syntax
> The document is modified from [https://github.com/shiyanhui/hero](https://github.com/shiyanhui/hero)

Goh has a total of nine statements, which are:

- Function Definition Statement `<%: func define %>`
  - This statement defines the function corresponding to the template. If there is no function definition statement in a template, the final result will not generate the corresponding function.
  - The last parameter of this function must be `*bytes.Buffer` or `io.Writer`. Hero will automatically recognize the name of this parameter and write the result to this parameter.
  - Examples:
    - `<%: func UserList(userList []string, buffer *bytes.Buffer) %>`
    - `<%: func UserList(userList []string, w io.Writer) %>`
    - `<%: func UserList(userList []string, w io.Writer) (int, error) %>`

- Template Inheritance Statement `<%~ "parent template" %>`
  - This statement declares the template to be inherited.
  - Example: `<%~ "index.html" >`
    
- Template Include Statement `<%+ "sub template" %>`
  - This statement loads the template to be included into this template. Its working principle is somewhat similar to `#include` in `C++`.
  - Example: `<%+ "user.html" >`
  
- Package Import Statement `<%! go code %>`
  - This statement is used to declare all codes outside the function, including dependency package imports, global variables, const, etc.
  - This statement will not be inherited by sub - templates.
  - Example:
    
    ```go
    <%!
    	import (
          	"fmt"
        	"strings"
        )

    	var a int

    	const b = "hello, world"

    	func Add(a, b int) int {
        	return a + b
    	}

    	type S struct {
        	Name string
    	}

    	func (s S) String() string {
        	return s.Name
    	}
    %>
    ```
  
- Block Statement `<%@ blockName { %> <%@ } %>`
> [!WARNING]
> Unlike Hero, closing a block requires the use of '<%@ } %>' instead of '<% } %>'`
> This change mainly improves compiler compilation performance and shortens compilation time

  - Block statement represents a block. Child template overwrites blocks to extend parent template.

  - Example:

    ```html
    <!DOCTYPE html>
    <html>
        <head>
            <meta charset="utf-8">
        </head>

        <body>
            <%@ body { %>
            <% } %>
        </body>
    </html>
    ```

- Go Code Statement `<% go code %>`
  - This statement defines the code part inside the function.
  - Example:
     ```go
    <% for _, user := range userList { %>
        <% if user != "Alice" { %>
        	<%= user %>
        <% } %>
    <% } %>

    <%
    	a, b := 1, 2
    	c := Add(a, b)
    %>
    ```
- Native Value Statement `<%==[t] variable %>`, `<%- variable %>`
  - This statement converts the variable to a string.
    - `t` is the type of the variable. Hero will automatically select the conversion function according to `t`. The candidate values of `t` are:
    - `b`: bool
    - `i`: int, int8, int16, int32, int64
    - `u`: byte, uint, uint8, uint16, uint32, uint64
    - `f`: float32, float64
    - `s`: string
    - `bs`: []byte
    - `v`: interface
    Note:
    - If `t` is not set, then `t` defaults to s.
    - It is best not to use `v` because its corresponding conversion function is `fmt.Sprintf("%v", variable)`, which is very slow.
  Examples:
    - `<%== "hello" %>`
    - `<%==i 34  %>`
    - `<%==u Add(a, b) %>`
    - `<%==s user.Name %>`
- Escaped Value Statement `<%= statement %>`
  - This statement converts the variable to a string and then escapes it through html.EscapesString.
  - t is the same as t in the above native value statement.
  - Examples:
    - `<%= a %>`
    - `<%= a + b %>`
    - `<%= Add(a, b) %>`
    - `<%= user.Name %>`
- Comment Statement `<%# note %>`
  - This statement comments on the relevant template. The comment will not be generated into the Go code.
  - Example: `<# This is a comment >`

## Acknowledgments
[Shiyanhui/hero](https://github.com/shiyanhui/hero)
