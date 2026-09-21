package i18n

import (
	"os/exec"
	"strings"
	"testing"
)

// TestRootPackageIsStandardLibraryOnly guards the two splits this package
// rests on.
//
// yamlloader exists because gopkg.in/yaml.v3 was reached from bundle.go, so
// every program importing this package carried a YAML parser even when all of
// its translation files were JSON -- which the standard library already reads.
//
// httpadapter exists because net/http was reached from four files, and it
// alone was 122 of the 203 packages the v3 root cost every importer. A CLI
// printing localized help or error messages has no web server in it and should
// not link one.
//
// Neither is self-enforcing. Adding `import "net/http"` back to a root-package
// file -- for an http.SameSite constant, say, or a convenience helper taking
// *http.Request -- compiles, passes every other test, and quietly puts those
// packages back into every importer's binary. This test is what fails instead.
//
// Test files are exempt by construction: `go list -deps .` reports the
// package's own import graph, not its tests'.
func TestRootPackageIsStandardLibraryOnly(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH; cannot inspect the import graph")
	}

	out, err := exec.Command("go", "list", "-deps", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps .: %v\n%s", err, out)
	}

	banned := map[string]string{
		"net/http":                    "use the httpadapter subpackage",
		"gopkg.in/yaml.v3":            "use the yamlloader subpackage",
		"github.com/gofiber/":         "use the fiberadapter subpackage",
		"github.com/valyala/fasthttp": "use the fiberadapter subpackage",
	}

	var linked []string
	for _, pkg := range strings.Fields(string(out)) {
		for name, remedy := range banned {
			hit := pkg == name
			if !hit && strings.HasSuffix(name, "/") {
				hit = strings.HasPrefix(pkg, name)
			}
			if hit {
				linked = append(linked, pkg+" ("+remedy+")")
			}
		}
	}

	if len(linked) > 0 {
		t.Errorf("the root package links %d package(s) it is supposed to stay clear of, want none:\n\t%s",
			len(linked), strings.Join(linked, "\n\t"))
	}
}
