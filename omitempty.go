package omitempty

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/JakubCzarlinski/go-logging"
)

func AddOmitJson(filePath string) error {
	// Parse the source code
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return logging.Bubble(err, "failed to parse file")
	}

	// Modify the AST
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.StructType:
			for _, field := range x.Fields.List {
				if field.Tag == nil {
					continue
				}
				if field.Tag.Value == "" || field.Tag.Kind != token.STRING {
					continue
				}

				field.Tag.Value = modifyJSONTag(field.Tag.Value)
			}
		}
		return true
	})

	// Write the output back to the original file
	buf := &bytes.Buffer{}
	err = format.Node(buf, fset, f)
	if err != nil {
		return logging.Bubble(err, "failed to format file")
	}
	outputFile, err := os.Create(filePath)
	if err != nil {
		return logging.Bubble(err, "failed to create file")
	}
	defer outputFile.Close()
	_, err = outputFile.Write(buf.Bytes())
	if err != nil {
		return logging.Bubble(err, "failed to write to file")
	}
	return nil
}

func modifyJSONTag(tagValue string) string {
	tagValue = strings.Trim(tagValue, "`")

	tags := strings.Split(tagValue, " ")
	var modifiedTags []string

	for _, tag := range tags {
		// Only modify JSON tags, leave others as they are.
		if !strings.HasPrefix(tag, "json:") {
			modifiedTags = append(modifiedTags, tag)
			continue
		}

		jsonQuoted := tag[5:]                        // Remove "json:" prefix
		jsonValue := strings.Trim(jsonQuoted, "\"")  // Remove quotes
		jsonOptions := strings.Split(jsonValue, ",") // Split options

		// Check if "omitempty" is already present
		hasOmitempty := false
		for _, opt := range jsonOptions {
			if opt == "omitempty" {
				hasOmitempty = true
				break
			}
		}

		// Add "omitempty" if not present and the field is not ignored
		if !hasOmitempty && jsonOptions[0] != "-" {
			jsonOptions = append(jsonOptions, "omitempty")
		}

		// Reconstruct the JSON tag
		newJSONTag := "json:\"" + strings.Join(jsonOptions, ",") + "\""
		modifiedTags = append(modifiedTags, newJSONTag)
	}

	// Reconstruct the full tag
	return "`" + strings.Join(modifiedTags, " ") + "`"
}
