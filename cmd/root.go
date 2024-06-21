/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "genTypes",
	Short: "Generate typescript file from go types",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		sourceFile, _ := cmd.Flags().GetString("source")
		targetFile, _ := cmd.Flags().GetString("target")

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, sourceFile, nil, 0)
		if err != nil {
			log.Fatal(err)
		}

		s := new(strings.Builder)

		// Collect the struct types in this slice.
		// var structTypes []*ast.TypeSpec

		// Use the Inspect function to walk AST looking for struct
		// type nodes.
		ast.Inspect(f, func(n ast.Node) bool {
			if n, ok := n.(*ast.GenDecl); ok {
				if n.Tok == token.IMPORT || n.Tok == token.VAR {
					return false
				}
				for _, dec := range n.Specs {
					generateTS(s, dec)
				}

			}
			return true
		})

		fmt.Println("generate called", sourceFile)
		str := s.String()

		fmt.Println(str, targetFile)

		err = os.MkdirAll(filepath.Dir(targetFile), os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}

		err = os.WriteFile(targetFile, []byte(str), 0664)
		if err != nil {
			fmt.Println(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringP("source", "s", "types.go", "Source file to read from")
	rootCmd.Flags().StringP("target", "t", "types.d.ts", "Target file to write to")
}
