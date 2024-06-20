/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
)

var fallbackType = "unknown"
var backquoteEscapeRegexp = regexp.MustCompile(`([$\\])`)
var validJSNameRegexp = regexp.MustCompile(`(?m)^[\pL_][\pL\pN_]*$`)
var octalPrefixRegexp = regexp.MustCompile(`^0[0-7]`)
var unicode8Regexp = regexp.MustCompile(`\\\\|\\U[\da-fA-F]{8}`)

var jsNumberOperatorPrecedence = map[token.Token]int{
	token.MUL:     6,
	token.QUO:     6,
	token.REM:     6,
	token.ADD:     5,
	token.SUB:     5,
	token.SHL:     4,
	token.SHR:     4,
	token.AND:     3,
	token.AND_NOT: 3,
	token.OR:      2,
	token.XOR:     1,
}

func generateTS(s *strings.Builder, spec ast.Spec) {

	ts, ok := spec.(*ast.TypeSpec)
	if ok {
		writeTypeSpec(s, ts)
	}
}

func writeTypeSpec(s *strings.Builder, ts *ast.TypeSpec) {

	st, isStruct := ts.Type.(*ast.StructType)
	if isStruct {
		s.WriteString("export interface ")
		s.WriteString(ts.Name.Name)

		if ts.TypeParams != nil {
			writeTypeParamsFields(s, ts.TypeParams.List)
		}

		s.WriteString(" {\n")
		writeStructFields(s, st.Fields.List, 0)
		s.WriteString("}\n\n")
	}

	id, isIdent := ts.Type.(*ast.Ident)
	if isIdent {
		s.WriteString("export type ")
		s.WriteString(ts.Name.Name)
		s.WriteString(" = ")
		s.WriteString(getIdent(id.Name))
		s.WriteString(";\n\n")
	}

	if !isStruct && !isIdent {
		s.WriteString("export type ")
		s.WriteString(ts.Name.Name)
		s.WriteString(" = ")
		writeType(s, ts.Type, nil, 0, true)
		s.WriteString(";\n\n")
	}
}

func writeTypeParamsFields(s *strings.Builder, fields []*ast.Field) {
	s.WriteByte('<')
	for i, f := range fields {
		for j, ident := range f.Names {
			s.WriteString(ident.Name)
			s.WriteString(" extends ")
			writeType(s, f.Type, nil, 0, true)

			if i != len(fields)-1 || j != len(f.Names)-1 {
				s.WriteString(", ")
			}
		}
	}
	s.WriteByte('>')
}

func writeStructFields(s *strings.Builder, fields []*ast.Field, depth int) {
	for _, f := range fields {
		optional := false
		required := false
		readonly := false

		var fieldName string
		if len(f.Names) == 0 { // anonymous field
			if name, valid := getAnonymousFieldName(f.Type); valid {
				fieldName = name
			}
		}
		if len(f.Names) != 0 && f.Names[0] != nil && len(f.Names[0].Name) != 0 {
			fieldName = f.Names[0].Name
		}
		if len(fieldName) == 0 || 'A' > fieldName[0] || fieldName[0] > 'Z' {
			continue
		}

		var name string
		var tstype string
		if f.Tag != nil {
			tags, err := structtag.Parse(f.Tag.Value[1 : len(f.Tag.Value)-1])
			if err != nil {
				panic(err)
			}

			jsonTag, err := tags.Get("json")
			if err == nil {
				name = jsonTag.Name
				if name == "-" {
					continue
				}

				optional = jsonTag.HasOption("omitempty")
			}
			yamlTag, err := tags.Get("yaml")
			if err == nil {
				name = yamlTag.Name
				if name == "-" {
					continue
				}

				optional = yamlTag.HasOption("omitempty")
			}

			tstypeTag, err := tags.Get("tstype")
			if err == nil {
				tstype = tstypeTag.Name
				if tstype == "-" || tstypeTag.HasOption("extends") {
					continue
				}
				required = tstypeTag.HasOption("required")
				readonly = tstypeTag.HasOption("readonly")
			}
		}

		writeIndent(s, depth+1)
		quoted := !validJSName(name)
		if quoted {
			s.WriteByte('\'')
		}
		if readonly {
			s.WriteString("readonly ")
		}
		s.WriteString(name)
		if quoted {
			s.WriteByte('\'')
		}

		switch t := f.Type.(type) {
		case *ast.StarExpr:
			optional = !required
			f.Type = t.X
		}

		if optional {
			s.WriteByte('?')
		}

		s.WriteString(": ")

		if tstype == "" {
			writeType(s, f.Type, nil, depth, false)
		} else {
			s.WriteString(tstype)
		}
		s.WriteByte(';')
		s.WriteByte('\n')

	}
}

func getAnonymousFieldName(f ast.Expr) (name string, valid bool) {
	switch ft := f.(type) {
	case *ast.Ident:
		name = ft.Name
		if ft.Obj != nil && ft.Obj.Decl != nil {
			dcl, ok := ft.Obj.Decl.(*ast.TypeSpec)
			if ok {
				valid = dcl.Name.IsExported()
			}
		} else {
			// Types defined in the Go file after the parsed file in the same package
			valid = token.IsExported(name)
		}
	case *ast.IndexExpr:
		return getAnonymousFieldName(ft.X)
	case *ast.IndexListExpr:
		return getAnonymousFieldName(ft.X)
	case *ast.SelectorExpr:
		valid = ft.Sel.IsExported()
		name = ft.Sel.String()
	case *ast.StarExpr:
		return getAnonymousFieldName(ft.X)
	}

	return
}

func writeType(
	s *strings.Builder,
	t ast.Expr,
	p ast.Expr,
	depth int,
	optionalParens bool,
) {
	switch t := t.(type) {
	case *ast.StarExpr:
		if optionalParens {
			s.WriteByte('(')
		}
		writeType(s, t.X, t, depth, false)
		s.WriteString(" | undefined")
		if optionalParens {
			s.WriteByte(')')
		}
	case *ast.ArrayType:
		if v, ok := t.Elt.(*ast.Ident); ok && v.String() == "byte" {
			s.WriteString("string")
			break
		}
		writeType(s, t.Elt, t, depth, true)
		s.WriteString("[]")
	case *ast.StructType:
		s.WriteString("{\n")
		writeStructFields(s, t.Fields.List, depth+1)
		writeIndent(s, depth+1)
		s.WriteByte('}')
	case *ast.Ident:
		if t.String() == "any" {
			s.WriteString(getIdent("unknown"))
		} else {
			s.WriteString(getIdent(t.String()))
		}
	case *ast.MapType:
		s.WriteString("{ [key: ")
		writeType(s, t.Key, t, depth, false)
		s.WriteString("]: ")
		writeType(s, t.Value, t, depth, false)
		s.WriteByte('}')
	case *ast.BasicLit:
		switch t.Kind {
		case token.INT:
			if octalPrefixRegexp.MatchString(t.Value) {
				t.Value = "0o" + t.Value[1:]
			}
		case token.CHAR:
			var char rune
			if strings.HasPrefix(t.Value, `'\x`) ||
				strings.HasPrefix(t.Value, `'\u`) ||
				strings.HasPrefix(t.Value, `'\U`) {
				i32, err := strconv.ParseInt(t.Value[3:len(t.Value)-1], 16, 32)
				if err != nil {
					panic(err)
				}
				char = rune(i32)
			} else {
				var data []byte
				data = append(data, '"')
				data = append(data, []byte(t.Value[1:len(t.Value)-1])...)
				data = append(data, '"')
				var s string
				err := json.Unmarshal(data, &s)
				if err != nil {
					panic(err)
				}
				char = []rune(s)[0]
			}
			if char > 0xFFFF {
				t.Value = fmt.Sprintf("0x%08X /* %s */", char, t.Value)
			} else {
				t.Value = fmt.Sprintf("0x%04X /* %s */", char, t.Value)
			}
		case token.STRING:
			if strings.HasPrefix(t.Value, "`") {
				t.Value = backquoteEscapeRegexp.ReplaceAllString(t.Value, `\$1`)
			} else {
				t.Value = unicode8Regexp.ReplaceAllStringFunc(t.Value, func(s string) string {
					if len(s) == 10 {
						s = fmt.Sprintf("\\u{%s}", strings.ToUpper(s[2:]))
					}
					return s
				})
			}
		}
		s.WriteString(t.Value)
	case *ast.ParenExpr:
		s.WriteByte('(')
		writeType(s, t.X, t, depth, false)
		s.WriteByte(')')
	case *ast.BinaryExpr:
		inParen := false
		switch p := p.(type) {
		case *ast.BinaryExpr:
			if jsNumberOperatorPrecedence[t.Op] < jsNumberOperatorPrecedence[p.Op] {
				inParen = true
			}
		}
		if inParen {
			s.WriteByte('(')
		}
		writeType(s, t.X, t, depth, false)
		s.WriteByte(' ')
		if t.Op == token.AND_NOT {
			s.WriteString("& ~")
		} else {
			s.WriteString(t.Op.String())
			s.WriteByte(' ')
		}
		writeType(s, t.Y, t, depth, false)
		if inParen {
			s.WriteByte(')')
		}
	case *ast.InterfaceType:
		writeInterfaceFields(s, t.Methods.List, depth+1)
	case *ast.CallExpr, *ast.FuncType, *ast.ChanType:
		s.WriteString(fallbackType)
	case *ast.UnaryExpr:
		switch t.Op {
		case token.TILDE:
			// We just ignore the tilde token, in Typescript extended types are
			// put into the generic typing itself, which we can't support yet.
			writeType(s, t.X, t, depth, false)
		case token.XOR:
			s.WriteString("~")
			writeType(s, t.X, t, depth, false)
		case token.ADD, token.SUB, token.NOT:
			s.WriteString(t.Op.String())
			writeType(s, t.X, t, depth, false)
		default:
			err := fmt.Errorf("unhandled unary expr: %v\n %T", t, t)
			fmt.Println(err)
			panic(err)
		}
	case *ast.IndexListExpr:
		writeType(s, t.X, t, depth, false)
		s.WriteByte('<')
		for i, index := range t.Indices {
			writeType(s, index, t, depth, false)
			if i != len(t.Indices)-1 {
				s.WriteString(", ")
			}
		}
		s.WriteByte('>')
	case *ast.IndexExpr:
		writeType(s, t.X, t, depth, false)
		s.WriteByte('<')
		writeType(s, t.Index, t, depth, false)
		s.WriteByte('>')
	default:
		err := fmt.Errorf("unhandled: %s\n %T", t, t)
		fmt.Println(err)
		panic(err)
	}
}

func writeInterfaceFields(
	s *strings.Builder,
	fields []*ast.Field,
	depth int,
) {
	// Usually interfaces in Golang don't have fields, but generic (union) interfaces we can map to Typescript.

	if len(fields) == 0 { // Type without any fields (probably only has methods)
		s.WriteString(fallbackType)
		return
	}

	didContainNonFuncFields := false
	for _, f := range fields {
		if _, isFunc := f.Type.(*ast.FuncType); isFunc {
			continue
		}
		if didContainNonFuncFields {
			s.WriteString(" &\n")
		} else {
			s.WriteByte(
				'\n',
			) // We need to write a newline so comments of generic components render nicely.
			didContainNonFuncFields = true
		}

		writeIndent(s, depth+1)
		writeType(s, f.Type, nil, depth, false)
	}

	if !didContainNonFuncFields {
		s.WriteString(fallbackType)
	}
}

func writeIndent(s *strings.Builder, depth int) {
	for i := 0; i < depth; i++ {
		s.WriteString("  ")
	}
}

func validJSName(n string) bool {
	return validJSNameRegexp.MatchString(n)
}

func getIdent(s string) string {
	switch s {
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"complex64", "complex128":
		return "number /* " + s + " */"
	}
	return s
}
