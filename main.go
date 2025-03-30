package main

import (
	"flag"
	"fmt"
	Goh "github.com/OblivionOcean/Goh/src"
	"os"
	"os/exec"
	"path"
)

// execCommand runs a shell command.
func execCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	// Define command-line flags
	dest := flag.String("dest", "", "generated golang files dir, it will be the same with source if not set")
	ext := flag.String("ext", ".html", "source file extensions, comma splitted if many")
	pkg := flag.String("pkg", "template", "the generated template package name, default is template")
	src := flag.String("src", "./", "the html template file or directory")

	// Parse the command-line flags
	flag.Parse()

	// If destination is not set, use the source directory
	if *dest == "" {
		*dest = *src
	}

	// Read the directory contents
	files, err := os.ReadDir(*src)
	if err != nil {
		panic(err.Error())
	}

	// Process each file in the directory
	for i := 0; i < len(files); i++ {
		if files[i].IsDir() {
			continue
		}
		if path.Ext(files[i].Name()) == *ext {
			fullPath := path.Join(*src, files[i].Name())
			g := &Goh.CodeGenerator{
				Destination: *dest,
				PackageName: *pkg,
			}
			g.NewGenerator(fullPath)
		}
	}

	// Ensure the destination directory exists
	if _, err := os.Stat(*dest); os.IsNotExist(err) {
		if err := os.MkdirAll(*dest, 0755); err != nil {
			panic(err.Error())
		}
	}
	os.Chdir(*dest)

	// Run goimports and gofmt on the generated code
	if err := execCommand("goimports -w ."); err != nil {
		fmt.Println("Error running goimports:", err)
	}
	if err := execCommand("gofmt -w ."); err != nil {
		fmt.Println("Error running gofmt:", err)
	}
}
