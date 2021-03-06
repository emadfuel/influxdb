//lint:file-ignore U1000 ignore these flagger-related dead code issues until we can circle back
package testing_test

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/stdlib"
	"github.com/influxdata/influxdb/v2/kit/feature"
	"github.com/influxdata/influxdb/v2/kit/feature/override"

	platform "github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/cmd/influxd/launcher"
	influxdbcontext "github.com/influxdata/influxdb/v2/context"
	"github.com/influxdata/influxdb/v2/mock"
	"github.com/influxdata/influxdb/v2/query"
	_ "github.com/influxdata/influxdb/v2/query/stdlib"
	itesting "github.com/influxdata/influxdb/v2/query/stdlib/testing" // Import the stdlib
)

// Flagger for end-to-end test cases. This flagger contains a pointer to a
// single struct instance that all the test cases will consult. It will return flags
// based on the contents of FluxEndToEndFeatureFlags and the currently active
// test case. This works only because tests are serialized. We can set the
// current test case in the common flagger state, then run the test. If we were
// to run tests in parallel we would need to create multiple users and assign
// them different flags combinations, then run the tests under different users.

type Flagger struct {
	flaggerState *FlaggerState
}

type FlaggerState struct {
	Path           string
	Name           string
	FeatureFlags   itesting.PerTestFeatureFlagMap
	DefaultFlagger feature.Flagger
}

func newFlagger(featureFlagMap itesting.PerTestFeatureFlagMap) Flagger {
	flaggerState := &FlaggerState{}
	flaggerState.FeatureFlags = featureFlagMap
	flaggerState.DefaultFlagger = feature.DefaultFlagger()
	return Flagger{flaggerState}
}

func (f Flagger) SetActiveTestCase(path string, name string) {
	f.flaggerState.Path = path
	f.flaggerState.Name = name
}

func (f Flagger) Flags(ctx context.Context, _f ...feature.Flag) (map[string]interface{}, error) {
	// If an override is set for the test case, construct an override flagger
	// and use it's computed flags.
	overrides := f.flaggerState.FeatureFlags[f.flaggerState.Path][f.flaggerState.Name]
	if overrides != nil {
		f, err := override.Make(overrides, nil)
		if err != nil {
			panic("failed to construct override flagger, probably an invalid flag in FluxEndToEndFeatureFlags")
		}
		return f.Flags(ctx)
	}

	// Otherwise use flags from a default flagger.
	return f.flaggerState.DefaultFlagger.Flags(ctx)
}

// Default context.
var ctx = influxdbcontext.SetAuthorizer(context.Background(), mock.NewMockAuthorizer(true, nil))

func init() {
	flux.FinalizeBuiltIns()
}

func TestFluxEndToEnd(t *testing.T) {
	t.Skip("Skipping per https://github.com/influxdata/influxdb/issues/19299")
	runEndToEnd(t, stdlib.FluxTestPackages)
}
func BenchmarkFluxEndToEnd(b *testing.B) {
	benchEndToEnd(b, stdlib.FluxTestPackages)
}

func runEndToEnd(t *testing.T, pkgs []*ast.Package) {
	flagger := newFlagger(itesting.FluxEndToEndFeatureFlags)
	l := launcher.RunTestLauncherOrFail(t, ctx, flagger)
	l.SetupOrFail(t)
	defer l.ShutdownOrFail(t, ctx)
	for _, pkg := range pkgs {
		test := func(t *testing.T, f func(t *testing.T)) {
			t.Run(pkg.Path, f)
		}
		if pkg.Path == "universe" {
			test = func(t *testing.T, f func(t *testing.T)) {
				f(t)
			}
		}

		test(t, func(t *testing.T) {
			for _, file := range pkg.Files {
				name := strings.TrimSuffix(file.Name, "_test.flux")
				t.Run(name, func(t *testing.T) {
					if reason, ok := itesting.FluxEndToEndSkipList[pkg.Path][name]; ok {
						t.Skip(reason)
					}
					testFlux(t, l, file)
				})
			}
		})
	}
}

func benchEndToEnd(b *testing.B, pkgs []*ast.Package) {
	// TODO(jsternberg): These benchmarks don't run properly
	// and need to be fixed. Commenting out the code for now.
	b.Skip("https://github.com/influxdata/influxdb/issues/15391")
	// l := launcher.RunTestLauncherOrFail(b, ctx)
	// l.SetupOrFail(b)
	// defer l.ShutdownOrFail(b, ctx)
	// for _, pkg := range pkgs {
	// 	pkg := pkg.Copy().(*ast.Package)
	// 	name := pkg.Files[0].Name
	// 	b.Run(name, func(b *testing.B) {
	// 		if reason, ok := itesting.FluxEndToEndSkipList[strings.TrimSuffix(name, ".flux")]; ok {
	// 			b.Skip(reason)
	// 		}
	// 		b.ResetTimer()
	// 		b.ReportAllocs()
	// 		for i := 0; i < b.N; i++ {
	// 			testFlux(b, l, pkg)
	// 		}
	// 	})
	// }
}

func makeTestPackage(file *ast.File) *ast.Package {
	file = file.Copy().(*ast.File)
	file.Package.Name.Name = "main"
	pkg := &ast.Package{
		Package: "main",
		Files:   []*ast.File{file},
	}
	return pkg
}

var optionsSource = `
import "testing"
import c "csv"

// Options bucket and org are defined dynamically per test

option testing.loadStorage = (csv) => {
	c.from(csv: csv) |> to(bucket: bucket, org: org)
	return from(bucket: bucket)
}
`
var optionsAST *ast.File

func init() {
	pkg := parser.ParseSource(optionsSource)
	if ast.Check(pkg) > 0 {
		panic(ast.GetError(pkg))
	}
	optionsAST = pkg.Files[0]
}

func testFlux(t testing.TB, l *launcher.TestLauncher, file *ast.File) {

	// Query server to ensure write persists.

	b := &platform.Bucket{
		OrgID:           l.Org.ID,
		Name:            t.Name(),
		RetentionPeriod: 0,
	}

	s := l.BucketService(t)
	if err := s.CreateBucket(context.Background(), b); err != nil {
		t.Fatal(err)
	}

	// Define bucket and org options
	bucketOpt := &ast.OptionStatement{
		Assignment: &ast.VariableAssignment{
			ID:   &ast.Identifier{Name: "bucket"},
			Init: &ast.StringLiteral{Value: b.Name},
		},
	}
	orgOpt := &ast.OptionStatement{
		Assignment: &ast.VariableAssignment{
			ID:   &ast.Identifier{Name: "org"},
			Init: &ast.StringLiteral{Value: l.Org.Name},
		},
	}
	options := optionsAST.Copy().(*ast.File)
	options.Body = append([]ast.Statement{bucketOpt, orgOpt}, options.Body...)

	// Add options to pkg
	pkg := makeTestPackage(file)
	pkg.Files = append(pkg.Files, options)

	// Add testing.inspect call to ensure the data is loaded
	inspectCalls := stdlib.TestingInspectCalls(pkg)
	pkg.Files = append(pkg.Files, inspectCalls)

	req := &query.Request{
		OrganizationID: l.Org.ID,
		Compiler:       lang.ASTCompiler{AST: pkg},
	}
	if r, err := l.FluxQueryService().Query(ctx, req); err != nil {
		t.Fatal(err)
	} else {
		for r.More() {
			v := r.Next()
			if err := v.Tables().Do(func(tbl flux.Table) error {
				return tbl.Do(func(reader flux.ColReader) error {
					return nil
				})
			}); err != nil {
				t.Error(err)
			}
		}
	}

	// quirk: our execution engine doesn't guarantee the order of execution for disconnected DAGS
	// so that our function-with-side effects call to `to` may run _after_ the test instead of before.
	// running twice makes sure that `to` happens at least once before we run the test.
	// this time we use a call to `run` so that the assertion error is triggered
	runCalls := stdlib.TestingRunCalls(pkg)
	pkg.Files[len(pkg.Files)-1] = runCalls
	r, err := l.FluxQueryService().Query(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	for r.More() {
		v := r.Next()
		if err := v.Tables().Do(func(tbl flux.Table) error {
			return tbl.Do(func(reader flux.ColReader) error {
				return nil
			})
		}); err != nil {
			t.Error(err)
		}
	}
	if err := r.Err(); err != nil {
		t.Error(err)
		// Replace the testing.run calls with testing.inspect calls.
		pkg.Files[len(pkg.Files)-1] = inspectCalls
		r, err := l.FluxQueryService().Query(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		defer func() {
			if t.Failed() {
				scanner := bufio.NewScanner(&out)
				for scanner.Scan() {
					t.Log(scanner.Text())
				}
			}
		}()
		for r.More() {
			v := r.Next()
			err := execute.FormatResult(&out, v)
			if err != nil {
				t.Error(err)
			}
		}
		if err := r.Err(); err != nil {
			t.Error(err)
		}
	}
}
