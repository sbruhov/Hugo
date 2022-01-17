// Copyright 2019 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package page_generate

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"

	"github.com/gohugoio/hugo/common/maps"

	"github.com/gohugoio/hugo/codegen"
	"github.com/gohugoio/hugo/resources/page"
	"github.com/gohugoio/hugo/resources/resource"
	"github.com/gohugoio/hugo/source"
)

const header = `// Copyright 2019 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file is autogenerated.
`

var (
	fileInterfaceDeprecated = reflect.TypeOf((*source.FileWithoutOverlap)(nil)).Elem()
	pageInterfaceDeprecated = reflect.TypeOf((*page.DeprecatedWarningPageMethods)(nil)).Elem()
	pageInterface           = reflect.TypeOf((*page.Page)(nil)).Elem()

	packageDir = filepath.FromSlash("resources/page")
)

func Generate(c *codegen.Inspector) error {
	if err := generateMarshalJSON(c); err != nil {
		return errors.Wrap(err, "failed to generate JSON marshaler")
	}

	if err := generateDeprecatedWrappers(c); err != nil {
		return errors.Wrap(err, "failed to generate deprecate wrappers")
	}

	if err := generateFileIsZeroWrappers(c); err != nil {
		return errors.Wrap(err, "failed to generate file wrappers")
	}

	return nil
}

func generateMarshalJSON(c *codegen.Inspector) error {
	filename := filepath.Join(c.ProjectRootDir, packageDir, "page_marshaljson.autogen.go")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	includes := []reflect.Type{pageInterface}

	// Exclude these methods
	excludes := []reflect.Type{
		// We need to evaluate the deprecated vs JSON in the future,
		// but leave them out for now.
		pageInterfaceDeprecated,

		// Leave this out for now. We need to revisit the author issue.
		reflect.TypeOf((*page.AuthorProvider)(nil)).Elem(),

		reflect.TypeOf((*resource.ErrProvider)(nil)).Elem(),

		// navigation.PageMenus

		// Prevent loops.
		reflect.TypeOf((*page.SitesProvider)(nil)).Elem(),
		reflect.TypeOf((*page.Positioner)(nil)).Elem(),

		reflect.TypeOf((*page.ChildCareProvider)(nil)).Elem(),
		reflect.TypeOf((*page.TreeProvider)(nil)).Elem(),
		reflect.TypeOf((*page.InSectionPositioner)(nil)).Elem(),
		reflect.TypeOf((*page.PaginatorProvider)(nil)).Elem(),
		reflect.TypeOf((*maps.Scratcher)(nil)).Elem(),
	}

	methods := c.MethodsFromTypes(
		includes,
		excludes)

	if len(methods) == 0 {
		return errors.New("no methods found")
	}

	marshalJSON, pkgImports := methods.ToMarshalJSON(
		"Page",
		"github.com/gohugoio/hugo/resources/page",
		// Exclusion regexps. Matches method names.
		`\bPage\b`,
	)

	fmt.Fprintf(f, `%s

package page

%s


%s


`, header, importsString(pkgImports), marshalJSON)

	return nil
}

func generateDeprecatedWrappers(c *codegen.Inspector) error {
	filename := filepath.Join(c.ProjectRootDir, packageDir, "page_wrappers.autogen.go")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Generate a wrapper for deprecated page methods

	reasons := map[string]string{
		"IsDraft":        "Use .Draft.",
		"Hugo":           "Use the global hugo function.",
		"LanguagePrefix": "Use .Site.LanguagePrefix.",
		"GetParam":       "Use .Param or .Params.myParam.",
		"RSSLink": `Use the Output Format's link, e.g. something like:
    {{ with .OutputFormats.Get "RSS" }}{{ .RelPermalink }}{{ end }}`,
		"URL": "Use .Permalink or .RelPermalink. If what you want is the front matter URL value, use .Params.url",
	}

	deprecated := func(name string, tp reflect.Type) string {
		var alternative string
		if tp == fileInterfaceDeprecated {
			alternative = "Use .File." + name
		} else {
			var found bool
			alternative, found = reasons[name]
			if !found {
				panic(fmt.Sprintf("no deprecated reason found for %q", name))
			}
		}

		return fmt.Sprintf("helpers.Deprecated(%q, %q, true)", "Page."+name, alternative)
	}

	var buff bytes.Buffer

	methods := c.MethodsFromTypes([]reflect.Type{fileInterfaceDeprecated, pageInterfaceDeprecated}, nil)

	for _, m := range methods {
		fmt.Fprint(&buff, m.Declaration("*pageDeprecated"))
		fmt.Fprintln(&buff, " {")
		fmt.Fprintf(&buff, "\t%s\n", deprecated(m.Name, m.Owner))
		fmt.Fprintf(&buff, "\t%s\n}\n", m.Delegate("p", "p"))

	}

	pkgImports := append(methods.Imports(), "github.com/gohugoio/hugo/helpers")

	fmt.Fprintf(f, `%s

package page

%s
// NewDeprecatedWarningPage adds deprecation warnings to the given implementation.
func NewDeprecatedWarningPage(p DeprecatedWarningPageMethods) DeprecatedWarningPageMethods {
	return &pageDeprecated{p: p}
}

type pageDeprecated struct {
	p DeprecatedWarningPageMethods
}

%s

`, header, importsString(pkgImports), buff.String())

	return nil
}

func generateFileIsZeroWrappers(c *codegen.Inspector) error {
	filename := filepath.Join(c.ProjectRootDir, packageDir, "zero_file.autogen.go")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Generate warnings for zero file access

	warning := func(name string, tp reflect.Type) string {
		msg := fmt.Sprintf(".File.%s on zero object. Wrap it in if or with: {{ with .File }}{{ .%s }}{{ end }}", name, name)

		// We made this a Warning in 0.92.0.
		// When we remove this construct in 0.93.0, people will get a nil pointer.
		return fmt.Sprintf("z.log.Warnln(%q)", msg)
	}

	var buff bytes.Buffer

	methods := c.MethodsFromTypes([]reflect.Type{reflect.TypeOf((*source.File)(nil)).Elem()}, nil)

	for _, m := range methods {
		if m.Name == "IsZero" {
			continue
		}
		fmt.Fprint(&buff, m.DeclarationNamed("zeroFile"))
		fmt.Fprintln(&buff, " {")
		fmt.Fprintf(&buff, "\t%s\n", warning(m.Name, m.Owner))
		if len(m.Out) > 0 {
			fmt.Fprintln(&buff, "\treturn")
		}
		fmt.Fprintln(&buff, "}")

	}

	pkgImports := append(methods.Imports(), "github.com/gohugoio/hugo/common/loggers", "github.com/gohugoio/hugo/source")

	fmt.Fprintf(f, `%s

package page

%s

// ZeroFile represents a zero value of source.File with warnings if invoked.
type zeroFile struct {
	log loggers.Logger
}

func NewZeroFile(log loggers.Logger) source.File {
	return zeroFile{log: log}
}

func (zeroFile) IsZero() bool {
	return true
}

%s

`, header, importsString(pkgImports), buff.String())

	return nil
}

func importsString(imps []string) string {
	if len(imps) == 0 {
		return ""
	}

	if len(imps) == 1 {
		return fmt.Sprintf("import %q", imps[0])
	}

	impsStr := "import (\n"
	for _, imp := range imps {
		impsStr += fmt.Sprintf("%q\n", imp)
	}

	return impsStr + ")"
}