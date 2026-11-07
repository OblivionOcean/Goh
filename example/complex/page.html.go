// DO NOT EDIT!
// Generate By Goh

package complex

import (
	"bytes"
	"github.com/OblivionOcean/Goh/utils"
)

func Page(title string, greeting string, activePage string, isAdmin bool, buf *bytes.Buffer) {
	buf.Grow(839)
	buf.WriteString(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>`)
	Goh.EscapeHTML(title, buf)
	buf.WriteString(`</title>
    `)
	buf.WriteString(`
    <meta name="description" content="default description">
    
<meta name="description" content="custom page description">
<meta name="author" content="the child">
`)
	buf.WriteString(`
</head>
<body>
    <header class="site-header">
        <h1>`)
	Goh.EscapeHTML(title, buf)
	buf.WriteString(`</h1>
    </header>
    <nav class="site-nav">
    `)
	if activePage == "/" {
		buf.WriteString(`
    <a href="/" class="active">Home</a>
    `)
	} else {
		buf.WriteString(`
    <a href="/">Home</a>
    `)
	}
	buf.WriteString(`
    `)
	if activePage == "/users" {
		buf.WriteString(`
    <a href="/users" class="active">Users</a>
    `)
	} else {
		buf.WriteString(`
    <a href="/users">Users</a>
    `)
	}
	buf.WriteString(`
    <a href="/about">About</a>
    <a href="/contact">Contact</a>
    `)
	if isAdmin {
		buf.WriteString(`
    <a href="/admin" class="admin-link">Admin</a>
    `)
	}
	buf.WriteString(`
</nav>

    `)
	buf.WriteString(`
    
<p class="greeting">`)
	Goh.EscapeHTML(greeting, buf)
	buf.WriteString(`</p>
<p class="custom-body">This body was overridden by the child.</p>
`)
	buf.WriteString(`
    <footer>&copy; 2026 Example Co.</footer>
</body>
</html>



`)
}
