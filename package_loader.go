package halstead

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type packageMetadata struct {
	Dir        string
	ImportPath string
	GoFiles    []string
	Export     string
}

// AnalyzeASTFile analyzes one Go file using package-aware type information from go list.
func AnalyzeASTFile(path string) (AnalysisReport, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return AnalysisReport{}, err
	}
	return AnalyzeASTSource(path, src)
}

// AnalyzeASTSource analyzes one Go source file using the package context inferred from path.
func AnalyzeASTSource(path string, src []byte) (AnalysisReport, error) {
	ctx, err := loadPackageContext(path, src)
	if err != nil {
		return AnalysisReport{}, err
	}

	analyzer := astAnalyzer{
		fileSet: ctx.fileSet,
		info:    ctx.info,
	}
	fileMetrics := analyzer.analyzeNode(ctx.targetFile)

	return AnalysisReport{
		Analyzer:  fileMetrics.Name,
		File:      fileMetrics.Summary(),
		Functions: analyzer.functionReports(ctx.targetFile),
	}, nil
}

type analysisContext struct {
	fileSet    *token.FileSet
	info       *types.Info
	targetFile *ast.File
}

func loadPackageContext(path string, src []byte) (analysisContext, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return analysisContext{}, err
	}
	dir := filepath.Dir(absPath)
	base := filepath.Base(absPath)

	pkg, deps, err := goListPackage(dir, base)
	if err != nil {
		return analysisContext{}, err
	}

	fileSet := token.NewFileSet()
	files := make([]*ast.File, 0, len(pkg.GoFiles))
	var targetFile *ast.File

	for _, name := range pkg.GoFiles {
		filename := filepath.Join(pkg.Dir, name)
		var fileSrc interface{}
		if name == base {
			fileSrc = src
		}
		file, err := parser.ParseFile(fileSet, filename, fileSrc, parser.ParseComments)
		if err != nil {
			return analysisContext{}, err
		}
		files = append(files, file)
		if name == base {
			targetFile = file
		}
	}
	if targetFile == nil {
		return analysisContext{}, fmt.Errorf("file %s is not part of package %s", base, pkg.ImportPath)
	}

	info := &types.Info{
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Types:      map[ast.Expr]types.TypeAndValue{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}

	config := &types.Config{
		Importer: importer.ForCompiler(fileSet, "gc", func(importPath string) (io.ReadCloser, error) {
			exportPath, ok := deps[importPath]
			if !ok || exportPath == "" {
				return nil, fmt.Errorf("can't find import: %q", importPath)
			}
			return os.Open(exportPath)
		}),
	}
	if _, err := config.Check(pkg.ImportPath, fileSet, files, info); err != nil {
		return analysisContext{}, err
	}

	return analysisContext{
		fileSet:    fileSet,
		info:       info,
		targetFile: targetFile,
	}, nil
}

func goListPackage(dir, targetBase string) (packageMetadata, map[string]string, error) {
	cmd := exec.Command("go", "list", "-json", "-deps", "-export", ".")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return packageMetadata{}, nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(output))
	deps := map[string]string{}
	var target packageMetadata
	foundTarget := false

	for {
		var pkg packageMetadata
		if err := decoder.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			return packageMetadata{}, nil, err
		}
		if pkg.ImportPath != "" && pkg.Export != "" {
			deps[pkg.ImportPath] = pkg.Export
		}
		if pkg.Dir == dir && containsFile(pkg.GoFiles, targetBase) {
			target = pkg
			foundTarget = true
		}
	}

	if !foundTarget {
		return packageMetadata{}, nil, fmt.Errorf("could not resolve package metadata for %s", filepath.Join(dir, targetBase))
	}
	return target, deps, nil
}

func containsFile(files []string, target string) bool {
	for _, file := range files {
		if file == target {
			return true
		}
	}
	return false
}
