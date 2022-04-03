package binding

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/internal/fs"

	"github.com/leaanthony/slicer"
)

//go:embed assets/package.json
var packageJSON []byte

func (b *Bindings) GenerateGoBindings(baseDir string) error {
	store := b.db.store
	for packageName, structs := range store {
		packageDir := filepath.Join(baseDir, packageName)
		err := fs.Mkdir(packageDir)
		if err != nil {
			return err
		}
		for structName, methods := range structs {
			var jsoutput bytes.Buffer
			jsoutput.WriteString(`// @ts-check
// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
`)
			var tsBody bytes.Buffer
			var tsContent bytes.Buffer
			tsContent.WriteString(`// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
`)
			var importClasses slicer.StringSlicer
			for methodName, methodDetails := range methods {

				// Generate JS
				var args slicer.StringSlicer
				for count := range methodDetails.Inputs {
					arg := fmt.Sprintf("arg%d", count+1)
					args.Add(arg)
				}
				argsString := args.Join(", ")
				jsoutput.WriteString(fmt.Sprintf("\nexport function %s(%s) {", methodName, argsString))
				jsoutput.WriteString("\n")
				jsoutput.WriteString(fmt.Sprintf("  return window['go']['%s']['%s']['%s'](%s);", packageName, structName, methodName, argsString))
				jsoutput.WriteString("\n")
				jsoutput.WriteString(fmt.Sprintf("}"))
				jsoutput.WriteString("\n")

				// Generate TS

				if len(methodDetails.StructNames) > 0 {
					importClasses.AddSlice(methodDetails.StructNames)
				}
				tsBody.WriteString(fmt.Sprintf("\nexport function %s(", methodName))

				args.Clear()
				for count, input := range methodDetails.Inputs {
					arg := fmt.Sprintf("arg%d", count+1)
					args.Add(arg + ":" + goTypeToTypescriptType(input.TypeName, false))
				}
				tsBody.WriteString(args.Join(",") + "):")
				returnType := "Promise"
				if methodDetails.OutputCount() > 0 {
					firstType := goTypeToTypescriptType(methodDetails.Outputs[0].TypeName, false)
					returnType += "<" + firstType
					if methodDetails.OutputCount() == 2 {
						secondType := goTypeToTypescriptType(methodDetails.Outputs[1].TypeName, false)
						returnType += "|" + secondType
					}
					returnType += ">"
				} else {
					returnType = "Promise<void>"
				}
				tsBody.WriteString(returnType + ";\n")
			}

			importClasses.Deduplicate()
			importClasses.Each(func(class string) {
				tsContent.WriteString("import {" + class + "} from '../models';\n")
			})
			tsContent.WriteString(tsBody.String())

			jsfilename := filepath.Join(packageDir, structName+".js")
			err = os.WriteFile(jsfilename, jsoutput.Bytes(), 0755)
			if err != nil {
				return err
			}
			tsfilename := filepath.Join(packageDir, structName+".d.ts")
			err = os.WriteFile(tsfilename, tsContent.Bytes(), 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Bindings) GenerateBackendJS(targetfile string) error {

	store := b.db.store
	var output bytes.Buffer

	output.WriteString(`// @ts-check
// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT

// ************************************************
// This file is deprecated and will not be generated
// in the next version of Wails. Bindings are now 
// generated in their own files.
// ************************************************

`)

	output.WriteString(`const go = {`)
	output.WriteString("\n")

	var sortedPackageNames slicer.StringSlicer
	for packageName := range store {
		sortedPackageNames.Add(packageName)
	}
	sortedPackageNames.Sort()
	sortedPackageNames.Each(func(packageName string) {
		packages := store[packageName]
		output.WriteString(fmt.Sprintf("  \"%s\": {", packageName))
		output.WriteString("\n")
		var sortedStructNames slicer.StringSlicer
		for structName := range packages {
			sortedStructNames.Add(structName)
		}
		sortedStructNames.Sort()

		sortedStructNames.Each(func(structName string) {
			structs := packages[structName]
			output.WriteString(fmt.Sprintf("    \"%s\": {", structName))
			output.WriteString("\n")

			var sortedMethodNames slicer.StringSlicer
			for methodName := range structs {
				sortedMethodNames.Add(methodName)
			}
			sortedMethodNames.Sort()

			sortedMethodNames.Each(func(methodName string) {
				methodDetails := structs[methodName]
				output.WriteString("      /**\n")
				output.WriteString("       * " + methodName + "\n")
				var args slicer.StringSlicer
				for count, input := range methodDetails.Inputs {
					arg := fmt.Sprintf("arg%d", count+1)
					args.Add(arg)
					output.WriteString(fmt.Sprintf("       * @param {%s} %s - Go Type: %s\n", goTypeToJSDocType(input.TypeName, true), arg, input.TypeName))
				}
				returnType := "Promise"
				returnTypeDetails := ""
				if methodDetails.OutputCount() > 0 {
					firstType := goTypeToJSDocType(methodDetails.Outputs[0].TypeName, true)
					returnType += "<" + firstType
					if methodDetails.OutputCount() == 2 {
						secondType := goTypeToJSDocType(methodDetails.Outputs[1].TypeName, true)
						returnType += "|" + secondType
					}
					returnType += ">"
					returnTypeDetails = " - Go Type: " + methodDetails.Outputs[0].TypeName
				} else {
					returnType = "Promise<void>"
				}
				output.WriteString("       * @returns {" + returnType + "} " + returnTypeDetails + "\n")
				output.WriteString("       */\n")
				argsString := args.Join(", ")
				output.WriteString(fmt.Sprintf("      \"%s\": (%s) => {", methodName, argsString))
				output.WriteString("\n")
				output.WriteString(fmt.Sprintf("        return window.go.%s.%s.%s(%s);", packageName, structName, methodName, argsString))
				output.WriteString("\n")
				output.WriteString(fmt.Sprintf("      },"))
				output.WriteString("\n")

			})

			output.WriteString("    },\n")
		})

		output.WriteString("  },\n\n")
	})

	output.WriteString(`};
export default go;`)
	output.WriteString("\n")

	dir := filepath.Dir(targetfile)
	packageJsonFile := filepath.Join(dir, "package.json")
	if !fs.FileExists(packageJsonFile) {
		err := os.WriteFile(packageJsonFile, packageJSON, 0755)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(targetfile, output.Bytes(), 0755)
}

// GenerateBackendTS generates typescript bindings for
// the bound methods.
func (b *Bindings) GenerateBackendTS(targetfile string) error {

	store := b.db.store
	var output bytes.Buffer

	output.WriteString(`

// ************************************************
// This file is deprecated and will not be generated
// in the next version of Wails. Bindings are now 
// generated in their own files.
// ************************************************

`)

	output.WriteString("import * as models from './models';\n\n")
	output.WriteString("export interface go {\n")

	var sortedPackageNames slicer.StringSlicer
	for packageName := range store {
		sortedPackageNames.Add(packageName)
	}
	sortedPackageNames.Sort()
	sortedPackageNames.Each(func(packageName string) {
		packages := store[packageName]
		output.WriteString(fmt.Sprintf("  \"%s\": {", packageName))
		output.WriteString("\n")
		var sortedStructNames slicer.StringSlicer
		for structName := range packages {
			sortedStructNames.Add(structName)
		}
		sortedStructNames.Sort()

		sortedStructNames.Each(func(structName string) {
			structs := packages[structName]
			output.WriteString(fmt.Sprintf("    \"%s\": {", structName))
			output.WriteString("\n")

			var sortedMethodNames slicer.StringSlicer
			for methodName := range structs {
				sortedMethodNames.Add(methodName)
			}
			sortedMethodNames.Sort()

			sortedMethodNames.Each(func(methodName string) {
				methodDetails := structs[methodName]
				output.WriteString(fmt.Sprintf("\t\t%s(", methodName))

				var args slicer.StringSlicer
				for count, input := range methodDetails.Inputs {
					arg := fmt.Sprintf("arg%d", count+1)
					args.Add(arg + ":" + goTypeToTypescriptType(input.TypeName, true))
				}
				output.WriteString(args.Join(",") + "):")
				returnType := "Promise"
				if methodDetails.OutputCount() > 0 {
					firstType := goTypeToTypescriptType(methodDetails.Outputs[0].TypeName, true)
					returnType += "<" + firstType
					if methodDetails.OutputCount() == 2 {
						secondType := goTypeToTypescriptType(methodDetails.Outputs[1].TypeName, true)
						returnType += "|" + secondType
					}
					returnType += ">"
				} else {
					returnType = "Promise<void>"
				}
				output.WriteString(returnType + "\n")
			})

			output.WriteString("    },\n")
		})
		output.WriteString("  }\n\n")
	})
	output.WriteString("}\n")

	globals := `
declare global {
	interface Window {
		go: go;
	}
}
`
	output.WriteString(globals)
	return os.WriteFile(targetfile, output.Bytes(), 0755)
}

func goTypeToJSDocType(input string, useModelsNamespace bool) string {
	switch true {
	case input == "interface{}":
		return "any"
	case input == "string":
		return "string"
	case input == "error":
		return "Error"
	case
		strings.HasPrefix(input, "int"),
		strings.HasPrefix(input, "uint"),
		strings.HasPrefix(input, "float"):
		return "number"
	case input == "bool":
		return "boolean"
	case input == "[]byte":
		return "string"
	case strings.HasPrefix(input, "[]"):
		arrayType := goTypeToJSDocType(input[2:], useModelsNamespace)
		return "Array<" + arrayType + ">"
	default:
		if strings.ContainsRune(input, '.') {
			if useModelsNamespace {
				return "models." + strings.Split(input, ".")[1]
			}
			return strings.Split(input, ".")[1]
		}
		return "any"
	}
}

func goTypeToTypescriptType(input string, useModelsNamespace bool) string {
	if strings.HasPrefix(input, "[]") {
		arrayType := goTypeToJSDocType(input[2:], useModelsNamespace)
		return "Array<" + arrayType + ">"
	}
	return goTypeToJSDocType(input, useModelsNamespace)
}
