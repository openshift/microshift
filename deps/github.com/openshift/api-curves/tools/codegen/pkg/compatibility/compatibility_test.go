package compatibility

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/google/go-cmp/cmp"
)

func TestProcessFile(t *testing.T) {

	testCases := []struct {
		name        string
		src         string
		expected    string
		expectError bool
	}{
		{
			name: "NothingToDo",
			src: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofStruct("TestApiOne"),
					withComments("// TestApiOne does something"),
				),
			),
			expected: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofStruct("TestApiOne"),
					withComments("// TestApiOne does something"),
				),
			),
		},
		{
			name: "NothingToDoNoComment",
			src: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofStruct("TestApiOne")),
			),
			expected: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofStruct("TestApiOne")),
			),
		},
		{
			name: "GA/Level1",
			src: src(
				withPackage("v1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level1TagComment)),
			),
			expected: src(
				withPackage("v1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					level1CompatibilityComment,
					level1TagComment,
				)),
			),
		},
		{
			name: "GA/Level2",
			src: src(
				withPackage("v1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level2TagComment)),
			),
			expectError: true,
		},
		{
			name: "GA/Level4",
			src: src(
				withPackage("v1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level4TagComment)),
			),
			expectError: true,
		},
		{
			name: "Prerelease/Level1",
			src: src(
				withPackage("v1beta1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level1TagComment)),
			),
			expectError: true,
		},
		{
			name: "Prerelease/Level2",
			src: src(
				withPackage("v1beta1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level2TagComment)),
			),
			expected: src(
				withPackage("v1beta1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					level2CompatibilityComment,
					level2TagComment,
				)),
			),
		},
		{
			name: "Prerelease/Level4",
			src: src(
				withPackage("v1beta1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level4TagComment)),
			),
			expectError: true,
		},
		{
			name: "Experimental/Level1",
			src: src(
				withPackage("v1alpha1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level1TagComment)),
			),
			expectError: true,
		},
		{
			name: "Experimental/Level2",
			src: src(
				withPackage("v1alpha1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level2TagComment)),
			),
			expectError: true,
		},
		{
			name: "Experimental/Level4",
			src: src(
				withPackage("v1alpha1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level4TagComment)),
			),
			expected: src(
				withPackage("v1alpha1"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					level4CompatibilityComment,
					level4TagComment,
				)),
			),
		},
		{
			name: "NonConforming/Level1",
			src: src(
				withPackage("docker10"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level1TagComment)),
			),
			expectError: true,
		},
		{
			name: "NonConforming/Level2",
			src: src(
				withPackage("docker10"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level2TagComment)),
			),
			expectError: true,
		},
		{
			name: "NonConforming/Level4",
			src: src(
				withPackage("docker10"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(level4TagComment)),
			),
			expectError: true,
		},
		{
			name: "NonConforming/Internal",
			src: src(
				withPackage("docker10"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(internalTagComment)),
			),
			expected: src(
				withPackage("docker10"),
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					level4CompatibilityComment,
					internalTagComment,
				)),
			),
		},
		{
			name: "CommentStyleDefault",
			src: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					"// TestApiOne does something",
					level1TagComment,
				)),
			),
			expected: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					"// TestApiOne does something",
					emptyComment,
					level1CompatibilityComment,
					level1TagComment,
				)),
			),
		},
		{
			name: "CommentStyleTagFirst",
			src: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					level1TagComment,
					"// TestApiOne does something",
				)),
			),
			expected: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					level1CompatibilityComment,
					level1TagComment,
					"// TestApiOne does something",
				)),
			),
		},
		{
			name: "CommentStyleTagSeparate",
			src: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					level1TagComment,
					emptyLineWithinComments,
					"// TestApiOne does something",
				)),
			),
			expected: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					level1TagComment,
					emptyLineWithinComments,
					"// TestApiOne does something",
					emptyComment,
					level1CompatibilityComment,
				)),
			),
		},
		{
			name: "CommentStyleTagFar",
			src: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					level1TagComment,
					emptyLineWithinComments,
					"// Another comment.",
					emptyLineWithinComments,
					"// TestApiOne does something",
				)),
			),
			expected: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					level1TagComment,
					emptyLineWithinComments,
					"// Another comment.",
					emptyLineWithinComments,
					"// TestApiOne does something",
					emptyComment,
					level1CompatibilityComment,
				)),
			),
		},
		{
			name: "CommentHasOtherTags",
			src: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					"// TestApiOne does something",
					"// +some-other-tag",
					level1TagComment,
				)),
			),
			expected: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					"// TestApiOne does something",
					emptyComment,
					level1CompatibilityComment,
					"// +some-other-tag",
					level1TagComment,
				)),
			),
		},
		{
			name: "LevelChanged",
			src: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					"// TestApiOne does something",
					level3CompatibilityComment,
					level1TagComment,
				)),
			),
			expected: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					"// TestApiOne does something",
					level1CompatibilityComment,
					level1TagComment,
				)),
			),
		},
		{
			name: "LevelCommentChanged",
			src: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					"// TestApiOne does something",
					"// Compatibility level 1: This is what is used to mean.",
					level1TagComment,
				)),
			),
			expected: src(
				withImport("metav1", "k8s.io/apimachinery/pkg/apis/meta/v1"),
				withDeclaration(ofAPIType("TestApiOne"), withComments(
					"// TestApiOne does something",
					level1CompatibilityComment,
					level1TagComment,
				)),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := decorator.ParseFile(fset, "", tc.src, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			_, err = processFile(f)
			if tc.expectError && err == nil {
				t.Fatal("Error expected.")
			}
			if tc.expectError && err != nil {
				t.Log(err)
				return
			}
			//dst.Print(f)
			buf := bytes.Buffer{}
			err = decorator.Fprint(&buf, f)
			if err != nil {
				t.Fatal(err)
			}
			actual := buf.String()
			if !cmp.Equal(tc.expected, actual) {
				t.Fatal(cmp.Diff(tc.expected, actual))
			}
			t.Log("\n" + actual)
		})
	}

}

func src(options ...func(file *dst.File)) string {
	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, file(options...)); err != nil {
		panic(err)
	}
	return buf.String()
}

func file(options ...func(file *dst.File)) *dst.File {
	f := &dst.File{Name: ident("v1")}
	for _, o := range options {
		o(f)
	}
	return f
}

func withPackage(name string) func(file *dst.File) {
	return func(file *dst.File) {
		file.Name = ident(name)
	}
}

func withImport(name, path string) func(file *dst.File) {
	return func(file *dst.File) {
		var importDecl *dst.GenDecl
		if len(file.Decls) > 0 {
			decl, ok := file.Decls[0].(*dst.GenDecl)
			if ok && decl.Tok == token.IMPORT {
				importDecl = decl
			}
		}
		if importDecl == nil {
			importDecl = &dst.GenDecl{Tok: token.IMPORT, Lparen: true, Rparen: true}
			file.Decls = append([]dst.Decl{importDecl}, file.Decls...)
		}
		importDecl.Specs = append(importDecl.Specs, &dst.ImportSpec{Name: ident(name), Path: stringLit(path)})
	}
}

func ofAPIType(name string, options ...func(*dst.StructType)) func(decl *dst.GenDecl) {
	return ofStruct(name, append(options, withEmbeddedTypeMeta())...)
}

func withDeclaration(options ...func(decl *dst.GenDecl)) func(*dst.File) {
	return func(file *dst.File) {
		decl := &dst.GenDecl{}
		for _, o := range options {
			o(decl)
		}
		file.Decls = append(file.Decls, decl)
	}
}

func ofStruct(name string, options ...func(*dst.StructType)) func(*dst.GenDecl) {
	return func(decl *dst.GenDecl) {
		decl.Tok = token.TYPE
		s := &dst.StructType{
			Fields: &dst.FieldList{},
		}
		decl.Specs = []dst.Spec{
			&dst.TypeSpec{
				Name: ident(name),
				Type: s,
			},
		}
		for _, o := range options {
			o(s)
		}
	}
}

func withEmbeddedTypeMeta() func(*dst.StructType) {
	return func(structType *dst.StructType) {
		structType.Fields.List = append(structType.Fields.List,
			&dst.Field{
				Names: nil,
				Type: &dst.SelectorExpr{
					X:   ident("metav1"),
					Sel: ident("TypeMeta"),
				},
				Tag:  nil,
				Decs: dst.FieldDecorations{},
			},
		)
	}
}

func withComments(comments ...string) func(node *dst.GenDecl) {
	return func(node *dst.GenDecl) {
		node.Decorations().Before = dst.EmptyLine
		node.Decorations().Start.Replace(comments...)
	}
}

func ident(name string) *dst.Ident {
	return &dst.Ident{Name: name}
}

func stringLit(s string) *dst.BasicLit {
	return &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", s)}
}

const (
	levelTagComment    = "// +openshift:compatibility-gen:level"
	level1TagComment   = levelTagComment + "=1"
	level2TagComment   = levelTagComment + "=2"
	level4TagComment   = levelTagComment + "=4"
	internalTagComment = "// +openshift:compatibility-gen:internal"

	level1CompatibilityComment = "// Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer)."
	level2CompatibilityComment = "// Compatibility level 2: Stable within a major release for a minimum of 9 months or 3 minor releases (whichever is longer)."
	level3CompatibilityComment = "// Compatibility level 3: Will attempt to be as compatible from version to version as possible, but version to version compatibility is not guaranteed."
	level4CompatibilityComment = "// Compatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support."

	emptyComment            = "//"
	emptyLineWithinComments = "\n"
)
