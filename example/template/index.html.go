// DO NOT EDIT!
// Generate By Goh

package template

import (
	"bytes"

	"github.com/OblivionOcean/Goh/utils"
)

func UserList(title string, userList []string, buf *bytes.Buffer) {
	buf.Grow(357)
	buf.WriteString(`
    <!DOCTYPE html>
    <html>

    <head>
        <title>
            `)
	Goh.EscapeHTML(title, buf)
	buf.WriteString(`
        </title>
    </head>

    <body>
        <h1>
            `)
	Goh.EscapeHTML(title, buf)
	buf.WriteString(`
        </h1>
        <ul>
            `)
	for _, user := range userList {
		buf.WriteString(`
                `)
		if user != "Alice" {
			buf.WriteString(`
                    <li>
                        `)
			Goh.EscapeHTML(user, buf)
			buf.WriteString(`
                    </li>
                    `)
		}
		buf.WriteString(`
                        `)
	}
	buf.WriteString(`
        </ul>
    </body>

    </html>`)
}
