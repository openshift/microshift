// Following program performs a static analysis of tests in OpenShift's origin repository to get K8s/OpenShift APIs particular tests are using.
// It's currently limited to test/extended/ directory.
// Program leverages Go tools to build following code representations: Abstract Syntax Trees (AST), Single Static Assignment (SSA) form,
// and call graphs using Rapid-Type Analysis (RTA) algorithm.
//
// In its current state it's only concerned with typed client-go's from Kubernetes and Openshift.
// Dynamic client-go (untyped, using unstructured.Unstructured) and OpenShift's CLI (`oc`-like interface) are next to be implemented.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	"k8s.io/apimachinery/pkg/util/sets"
	klog "k8s.io/klog/v2"
)

// TODO: Adjust report to requirements
// TODO: Handle k8s.io/client-go/dynamic & github.com/openshift/origin/test/extended/util
// TODO: Confirm correctness of the report

func main() {
	var originPathArg = flag.String("origin", "", "path to origin repository")
	var testdirRegexp = flag.String("filter", "", "regexp to filter test dirs")

	flag.Parse()
	defer klog.Flush()

	if *originPathArg == "" {
		klog.Fatalf("Provide path to origin repository using -origin")
	}

	if exists, err := checkIfPathExists(*originPathArg); !exists {
		klog.Exitf("Path %s does not exist", *originPathArg)
	} else if err != nil {
		klog.Exitf("Error occurred when checking if path %s exists: %v", *originPathArg, err)
	}

	var rx *regexp.Regexp
	if *testdirRegexp != "" {
		rx = regexp.MustCompile(*testdirRegexp)
	}

	report, err := buildReportUsingRTA(*originPathArg, rx)
	if err != nil {
		klog.Fatalf("Failed to build report: %v", err)
	}

	b, err := json.MarshalIndent(report, "", "   ")
	if err != nil {
		klog.Fatalf("Failed to marshal report: %v", err)
	}
	klog.Infof("Report: %v", string(b))
}

// Report is final result of the program. It's subject to changes to align with requirements coming from integration with origin repository.
type Report struct {
	TestAPICalls map[string][]string
	Ignored      []string
}

// APICallInTest represents a single call to the API (described as Group-Version-Kind) within some test tree.
type APICallInTest struct {
	Test []string // [Describe-Description, [Describe/Context-Description...], It-Description]
	GVK  string
}

// IgnoredCallInTest is a record of function call that was not considered an API call but was still persisted for debug purposes.
type IgnoredCallInTest struct {
	Test []string
	Pkg  string
	Recv string
	Func string
	Args []string
}

func buildReportUsingRTA(originPath string, rx *regexp.Regexp) (*Report, error) {
	cfg := packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes |
			packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps,
		Dir:   originPath,
		Tests: true,
	}

	originPkgs := func() []string {
		res := []string{}
		_ = filepath.WalkDir(originPath+"/test/extended/", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				klog.Fatalf("Error when walking origin's dir tree: %v", err)
			}
			if d.IsDir() &&
				!strings.Contains(path, "testdata") &&
				(rx == nil || (rx != nil && rx.MatchString(path))) {
				res = append(res, path)
			}
			return nil
		})
		return res
	}()
	klog.Infof("%d pkgs to scan: %#v", len(originPkgs), originPkgs)

	// Load packages into a memory with their Abstract Syntax Trees (stored in Syntax member)
	klog.Infof("Building Abstract Syntax Trees")
	initial, err := packages.Load(&cfg, originPkgs...)
	if err != nil {
		return nil, fmt.Errorf("packages.Load failed: %w", err)
	}

	// Build a Single Static Assignment (SSA) form of packages
	klog.Infof("Building Single Static Assignment forms")
	prog, _ := ssautil.AllPackages(initial, ssa.InstantiateGenerics)
	prog.Build()

	// Go through packages from github.com/openshift/origin/test/extended/ and look for init.start block which contains
	// variables created during init such as `var _ = ginkgo.Describe("DESC", func() {...})`.
	// Store their description and pointer to the `func() {...}`.
	describeFuncs := map[string]*ssa.Function{}
	for _, pkg := range prog.AllPackages() {
		if pkg != nil && strings.Contains(pkg.Pkg.Path(), "github.com/openshift/origin/test/extended/") {
			init := pkg.Members["init"].(*ssa.Function)
			initStart := init.Blocks[1] // "init.start"

			for _, instr := range initStart.Instrs {
				if c, ok := instr.(*ssa.Call); ok {
					if f, ok := c.Call.Value.(*ssa.Function); ok {
						if f.Name() == "Describe" {
							describeFuncs[strings.ReplaceAll(getStringFromValue(c.Call.Args[0]), "\"", "")] = c.Call.Args[1].(*ssa.Function)
						}
					}
				}
			}
		}
	}

	describeFuncsSlice := make([]*ssa.Function, 0, len(describeFuncs))
	for _, fun := range describeFuncs {
		describeFuncsSlice = append(describeFuncsSlice, fun)
	}
	klog.Infof("Found %d top level ginkgo Describes to be traversed. Note: It might not align with list of packages to scan.", len(describeFuncsSlice))

	// Perform a Rapid Type Analysis (RTA) for given functions. This will build a call graph as well.
	// This can even last about 5 minutes or more (depends on how many packages we want to analyze).
	klog.Infof("Building call graphs using Rapid Type Analysis (RTA) - it can take several minutes")
	rtaAnalysis := rta.Analyze(describeFuncsSlice, true)

	// Go routine receiving results (instead of returning them via return and merging them)
	apiChan := make(chan APICallInTest)
	apiCalls := []APICallInTest{}
	ignoredChan := make(chan IgnoredCallInTest) // Ignored calls are persisted for debugging
	ignoredCalls := sets.String{}
	go func() {
		for {
			select {
			case c, ok := <-apiChan:
				if !ok {
					return
				}
				apiCalls = append(apiCalls, c)
			case c, ok := <-ignoredChan:
				if !ok {
					return
				}
				// Simple string set for ignored otherwise we'd use a lot of memory just for these
				ignoredCalls.Insert(fmt.Sprintf("(%s.%s).%s()", c.Pkg, c.Recv, c.Func))
			}
		}
	}()

	wg := &sync.WaitGroup{}
	// Go through previously stored list of functions given to ginkgo.Describe and traverse their call graph.
	klog.Infof("Traversing %d tests - start", len(describeFuncs))
	for desc, fun := range describeFuncs {
		node, found := rtaAnalysis.CallGraph.Nodes[fun]
		if !found {
			return nil, fmt.Errorf("couldn't find %s (%s) in rtaAnalysis.CallGraph.Nodes[]", fun.Name(), desc)
		}
		wg.Add(1)
		go func(m map[*ssa.Function]*callgraph.Node, node *callgraph.Node, desc string,
			apiChan chan<- APICallInTest, ignoreChan chan<- IgnoredCallInTest) {
			klog.Infof("Inspecting '%s'", desc)
			res := traverseNodes(m, node, []string{desc}, apiChan, ignoreChan, 1)
			klog.Infof("Done inspecting '%s'. Potential call loop detected (quick unwind): %v", desc, !res)
			wg.Done()
		}(rtaAnalysis.CallGraph.Nodes, node, desc, apiChan, ignoredChan)
	}
	wg.Wait()
	close(apiChan)
	close(ignoredChan)

	return &Report{
		TestAPICalls: func(acits []APICallInTest) map[string][]string {
			m := make(map[string]sets.String)
			for _, acit := range acits {
				testName := strings.Join(acit.Test, " ")
				if _, found := m[testName]; found {
					m[testName] = m[testName].Insert(acit.GVK)
				} else {
					m[testName] = sets.NewString(acit.GVK)
				}
			}
			res := make(map[string][]string)
			for k, v := range m {
				res[k] = v.List()
			}
			return res
		}(apiCalls),
		Ignored: ignoredCalls.List(),
	}, nil
}

var potentialCallLoop = sync.Map{}

// traverseNodes visits callee's executed by given node.Caller in depth first manner.
// Callees are inspected in terms of pkg and function name to determine if they're an API call.
// Each ginkgo node (e.g. Describe, Context, It) appends a new string to parentTestTree.
// If a function call is considered final (API or ignored), it's reported via either apiChan or ignoreChan.
// Return value means if traversal should proceed. Currently primary case when traversal should stop is
// traversing gets too deep - that might mean a infinite call loop was detected.
func traverseNodes(m map[*ssa.Function]*callgraph.Node, node *callgraph.Node, parentTestTree []string,
	apiChan chan<- APICallInTest, ignoreChan chan<- IgnoredCallInTest, depth int) bool {
	// In some tests there appears to be a call loop going through couple of the same tests over and over.
	// Following block is just reports them.
	// TODO: Investigate and figure out how to break that loop
	if depth > 50 {
		if _, found := potentialCallLoop.Load(parentTestTree[0]); !found {
			potentialCallLoop.Store(parentTestTree[0], true)
			klog.Infof("Potential infinite call loop detected: '%s'\n", parentTestTree[0])
		}
		return false
	}

	// Node represents specific function within a call graph
	// This function makes a function calls represented as edges
	for _, edge := range node.Out {
		callee := edge.Callee
		funcName := callee.Func.Name()
		pkg := getCalleesPkgPath(callee)

		testTree := make([]string, len(parentTestTree))
		copy(testTree, parentTestTree)

		// Let's go into Defers only if they're not GinkgoRecover
		if _, ok := edge.Site.(*ssa.Defer); ok {
			if funcName != "GinkgoRecover" {
				if !traverseNodes(m, edge.Callee, testTree, apiChan, ignoreChan, depth+1) {
					return false
				}
			}
			continue
		}

		// If function call is ginkgo.Context/Describe/It then
		// we want to get its description
		// Context/It("desc", func(){))
		//          and visit ^^^^^^^^
		if pkg == "github.com/onsi/ginkgo" {
			if funcName == "Context" || funcName == "Describe" || funcName == "It" {
				args := edge.Site.Value().Call.Args

				f := getFuncFromValue(args[1])
				nodeToVisit, found := m[f]
				if !found {
					panic(fmt.Sprintf("f3(%v) not found in m", f.Name()))
				}

				ginkgoDesc := getStringFromValue(args[0])
				testTree = append(testTree, strings.ReplaceAll(ginkgoDesc, "\"", ""))
				if !traverseNodes(m, nodeToVisit, testTree, apiChan, ignoreChan, depth+1) {
					return false
				}
			}

		} else if (strings.Contains(pkg, "github.com/openshift/client-go") || strings.Contains(pkg, "k8s.io/client-go")) &&
			!strings.Contains(pkg, "k8s.io/client-go/dynamic") {
			recv, _ := getRecvFromFunc(callee.Func)
			if recv == nil {
				continue
			}

			if checkIfClientGoInterface(recv) {
				for i := 0; i < recv.NumMethods(); i++ {
					if recv.Method(i).Name() == "Get" {
						obj := recv.Method(i).Type().(*types.Signature).Results().At(0).Type().(*types.Pointer).Elem().(*types.Named).Obj()
						gvk := fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name())
						apiChan <- APICallInTest{
							Test: testTree,
							GVK:  gvk,
						}
					}
				}
			}

		} else if strings.Contains(pkg, "k8s.io/kubernetes/test/e2e") || strings.Contains(pkg, "github.com/openshift/origin/test/") {
			// Go into the helper functions but avoid recursion
			if edge.Callee != edge.Caller {
				if !traverseNodes(m, callee, testTree, apiChan, ignoreChan, depth+1) {
					return false
				}
			}

		} else if !strings.Contains(pkg, "gomega") && !strings.Contains(pkg, "ginkgo") && !strings.Contains(pkg, "golang.org") {
			if _, found := standardPackages[pkg]; !found {
				// Store ignored calls for debug purposes
				_, recvName := getRecvFromFunc(callee.Func)
				ignoreChan <- IgnoredCallInTest{
					Test: testTree,
					Pkg:  pkg,
					Recv: recvName,
					Func: funcName,
					Args: argsToStrings(edge.Site.Common().Args),
				}
			}
		}
	}

	return true
}

// standardPackages is a naive list of Go's standard packages from which function calls
// are completely ignored (they don't even make it to the report's ignored calls)
var standardPackages = func() map[string]struct{} {
	pkgs, err := packages.Load(nil, "std")
	if err != nil {
		panic(err)
	}

	result := make(map[string]struct{})
	for _, p := range pkgs {
		result[p.PkgPath] = struct{}{}
	}

	return result
}()

// argsToStrings transforms []ssa.Value
func argsToStrings(args []ssa.Value) []string {
	as := []string{}
	for _, arg := range args {
		as = append(as, arg.String())
	}
	return as
}

func getCalleesPkgPath(callee *callgraph.Node) string {
	if callee.Func.Pkg != nil {
		return callee.Func.Pkg.Pkg.Path()
	} else if callee.Func.Signature.Recv() != nil {
		// if Recv != nil, then it's stored as a first parameter
		if callee.Func.Params[0].Object().Pkg() != nil {
			return callee.Func.Params[0].Object().Pkg().Path()
		} else {
			// callee can have a receiver, but not be a part of a Pkg like error
			return ""
		}
	} else if callee.Func.Object() != nil && callee.Func.Object().Pkg() != nil {
		return callee.Func.Object().Pkg().Path()
	}

	panic("unknown calleePkg")
}

func getFuncFromValue(v ssa.Value) *ssa.Function {
	getFuncFromClosure := func(mc *ssa.MakeClosure) *ssa.Function {
		if ssaFunc, ok := mc.Fn.(*ssa.Function); ok {
			return ssaFunc
		}
		klog.Fatalf("getFuncFromClosure: mc.Fn is not *ssa.Function")
		return nil
	}

	switch v := v.(type) {
	case *ssa.MakeInterface:
		switch x := v.X.(type) {
		case *ssa.MakeClosure:
			return getFuncFromClosure(x)
		case *ssa.MakeInterface:
			if c, ok := x.X.(*ssa.Call); ok {
				if mc, ok := c.Call.Value.(*ssa.MakeClosure); ok {
					return getFuncFromClosure(mc)
				}
			}
		case *ssa.Function:
			return x
		case *ssa.Call:
			if mc, ok := x.Call.Value.(*ssa.MakeClosure); ok {
				return getFuncFromClosure(mc)
			}
		default:
			klog.Fatalf("getFuncFromValue: mi.X.(type) not handled: %T", x)
		}

	case *ssa.MakeClosure:
		return getFuncFromClosure(v)
	case *ssa.Function:
		return v
	case *ssa.Call:
		if f, ok := v.Call.Value.(*ssa.Function); ok {
			return f
		} else {
			klog.Fatalf("getFuncFromValue: v.Call.Value's type is unexpected: %T", v.Call.Value)
		}
	}

	klog.Fatalf("getFuncFromValue: v's type is unexpected: %T", v)
	return nil
}

// getStringFromValue converts ssa.Value to string, panics if fails
// It expects to receive an arg0 of ginkgo's Describe/Context/It
func getStringFromValue(v ssa.Value) string {
	switch v := v.(type) {
	case *ssa.Const:
		// Describe(STRING, ...)
		return v.Value.ExactString()

	case *ssa.Call:
		if v.Call.Value.Name() == "Sprintf" {
			if c, ok := v.Call.Args[0].(*ssa.Const); ok {
				// Describe(fmt.Sprintf(STRING, ...), ...)
				return c.Value.String()
			} else if bo, ok := v.Call.Args[0].(*ssa.BinOp); ok {
				return bo.String()
			}
		} else {
			klog.Fatalf("getStringFromValue: unexpected call: %s", v.Call.Value.Name())
		}

	case *ssa.BinOp:
		// It("test/cmd/"+currFilename, ...)
		// TODO: Better string
		return v.String()
	}

	panic(fmt.Sprintf("getStringFromValue: v's type in unexpected: %T", v))
}

var expectedMethods = []string{
	"Create",
	"Update",
	"Delete",
	"DeleteCollection",
	"Get",
	"List",
	"Watch",
	"Patch",
	// Not every interface we're interested in implements "UpdateStatus", "Apply", "ApplyStatus"
}

// checkIfClientGoInterface checks given types.Named if its methods are intersecting with a set of methods that client-go interface should have
func checkIfClientGoInterface(n *types.Named) bool {
	methods := sets.String{}
	for i := 0; i < n.NumMethods(); i++ {
		methods.Insert(n.Method(i).Name())
	}
	return methods.HasAll(expectedMethods...)
}

// getRecvFromFunc tries to get the receiver object and/or its name from given ssa.Function
func getRecvFromFunc(f *ssa.Function) (*types.Named, string) {
	if f.Signature.Recv() == nil {
		return nil, ""
	}

	arg0type := f.Params[0].Type()

	if ptr, ok := arg0type.(*types.Pointer); ok {
		return ptr.Elem().(*types.Named),
			ptr.Elem().(*types.Named).Obj().Name()

	} else if named, ok := arg0type.(*types.Named); ok {
		return named,
			named.Obj().Name()

	} else if strct, ok := arg0type.(*types.Struct); ok {
		// Some structure embedded some `error` and Error() was called
		return nil,
			strct.Field(0).Name()

	} else {
		panic("investigate arg0type type")
	}
}

func checkIfPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
