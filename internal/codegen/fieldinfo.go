package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// FieldTypeMap maps "StructName.FieldName" to Go type strings.
// It is populated by ParseFieldTypes from the actual source files.
var FieldTypeMap = map[string]string{}

// ParseFieldTypes parses all Go files in the given directory and populates
// FieldTypeMap with struct field type information. Embedded structs are
// resolved so that e.g. JobWithEnvironment inherits Job's fields.
func ParseFieldTypes(dir string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", dir, err)
	}

	// First pass: collect direct fields for each struct.
	type structInfo struct {
		fields   map[string]string // field name -> type string
		embeds   []string          // embedded struct names
	}
	structs := map[string]*structInfo{}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}
				for _, spec := range genDecl.Specs {
					typeSpec := spec.(*ast.TypeSpec)
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					name := typeSpec.Name.Name
					info := &structInfo{
						fields: map[string]string{},
					}
					for _, field := range structType.Fields.List {
						typeStr := exprToString(field.Type)
						if len(field.Names) == 0 {
							// Embedded field
							info.embeds = append(info.embeds, typeStr)
							continue
						}
						for _, ident := range field.Names {
							info.fields[ident.Name] = typeStr
						}
					}
					structs[name] = info
				}
			}
		}
	}

	// Second pass: resolve embeds and populate the global map.
	for name, info := range structs {
		for fieldName, fieldType := range info.fields {
			FieldTypeMap[name+"."+fieldName] = fieldType
		}
		for _, embed := range info.embeds {
			if embedded, ok := structs[embed]; ok {
				for fieldName, fieldType := range embedded.fields {
					FieldTypeMap[name+"."+fieldName] = fieldType
				}
			}
		}
	}

	return nil
}

// exprToString converts an ast.Expr to its Go source representation.
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + exprToString(e.Elt)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.MapType:
		return "map[" + exprToString(e.Key) + "]" + exprToString(e.Value)
	case *ast.InterfaceType:
		return "any"
	default:
		return fmt.Sprintf("%T", expr)
	}
}

