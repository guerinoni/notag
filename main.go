package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/guerinoni/notag/pkg/analyzer"
)

func main() {
	singlechecker.Main(analyzer.NewAnalyzer())
}
