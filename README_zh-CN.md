# Goh
[![GoDoc](https://pkg.go.dev/badge/github.com/OblivionOcean/Goh)](https://pkg.go.dev/github.com/OblivionOcean/Goh)
[![Go Report Card](https://goreportcard.com/badge/github.com/OblivionOcean/Goh)](https://goreportcard.com/report/github.com/OblivionOcean/Goh)

Goh 是一款Go语言的预编译快速模板引擎。

[English](https://github.com/OblivionOcean/Goh/blob/master/README.md) | 简体中文
## 目录
- [特性](#特性)
- [性能测试](#性能测试)
- [安装](#安装)
- [使用](#使用)
- [语法](#语法)
  
## 特性
- [x] 预编译模板引擎，提升运行速度。
- [x] 几乎兼容·Go语言的语法。
- [x] 0依赖。
- [ ] 更改模板文件后自动重新编译。

## 性能测试
从 https://github.com/slinso/goTemplateBenchmark 获取，目前为本地测试结果，代码与Hero部分的测试代码相同
```
goos: windows
goarch: amd64
pkg: github.com/SlinSo/goTemplateBenchmark
cpu: Intel(R) Core(TM) i7-10700 CPU @ 2.90GHz
# 复杂模板测试
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
# 简单模板测试
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

## 安装
```shell
go get -u github.com/OblivionOcean/Goh
go install github.com/OblivionOcean/Goh

# 依赖
go get golang.org/x/tools/cmd/goimports
go install golang.org/x/tools/cmd/goimports
```
## 使用
```
~ $ Goh
Usage of ./Goh:
  -dest string
        generated golang files dir, it will be the same with source if not set
  -ext string
        source file extensions, comma splitted if many (default ".html")
  -pkg string
        the generated template package name, default is template (default "template")
  -src string
        the html template file or directory (default "./")
```
> 完整的使用方法请参考[实例程序](https://github.com/OblivionOcean/Goh/tree/master/example)

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
## 语法
> 文档修改自[https://github.com/shiyanhui/hero](https://github.com/shiyanhui/hero)

Goh总共有九种语句，他们分别是：

- 函数定义语句 `<%: func define %>`
  - 该语句定义了该模板所对应的函数，如果一个模板中没有函数定义语句，那么最终结果不会生成对应的函数。
  - 该函数最后一个参数必须为`*bytes.Buffer`或者`io.Writer`, hero会自动识别该参数的名字，并把把结果写到该参数里。
  - 例:
    - `<%: func UserList(userList []string, buffer *bytes.Buffer) %>`
    - `<%: func UserList(userList []string, w io.Writer) %>`
    - `<%: func UserList(userList []string, w io.Writer) (int, error) %>`

- 模板继承语句 `<%~ "parent template" %>`
  - 该语句声明要继承的模板。
  - 例: `<%~ "index.html" >`

- 模板include语句 `<%+ "sub template" %>`
  - 该语句把要include的模板加载进该模板，工作原理和`C++`中的`#include`有点类似。
  - 例: `<%+ "user.html" >`

- 包导入语句 `<%! go code %>`
  - 该语句用来声明所有在函数外的代码，包括依赖包导入、全局变量、const等。

  - 该语句不会被子模板所继承

  - 例:

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

- 块语句 `<%@ blockName { %> <% } %>`
> [!WARNING]
> 暂不支持该语法，请使用其他方式代替。

- Go代码语句 `<% go code %>`

  - 该语句定义了函数内部的代码部分。

  - 例:

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

- 原生值语句 `<%==[t] variable %>`、`<%- variable %>`

  - 该语句把变量转换为string。

  - `t`是变量的类型，hero会自动根据`t`来选择转换函数。`t`的待选值有:
    - `b`: bool
    - `i`: int, int8, int16, int32, int64
    - `u`: byte, uint, uint8, uint16, uint32, uint64
    - `f`: float32, float64
    - `s`: string
    - `bs`: []byte
    - `v`: interface

    注意：
    - 如果`t`没有设置，那么`t`默认为`s`.
    - 最好不要使用`v`，因为其对应的转换函数为`fmt.Sprintf("%v", variable)`，该函数很慢。

  - 例:

    ```go
    <%== "hello" %>
    <%==i 34  %>
    <%==u Add(a, b) %>
    <%==s user.Name %>
    ```

- 转义值语句 `<%= statement %>`

  - 该语句把变量转换为string后，又通过`html.EscapesString`记性转义。
  - `t`跟上面原生值语句中的`t`一样。
  - 例:

    ```go
    <%= a %>
    <%= a + b %>
    <%= Add(a, b) %>
    <%= user.Name %>
    ```

- 注释语句 `<%# note %>`

  - 该语句注释相关模板，注释不会被生成到go代码里边去。
  - 例: `<# 这是一个注释 >`.

## 感谢
[Shiyanhui/hero](https://github.com/shiyanhui/hero)
