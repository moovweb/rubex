package main

import rubex "rubex"

import "testing"
import __os__ "os"
import __regexp__ "regexp"

var tests = []testing.InternalTest{
	{"rubex.TestGoodCompile", rubex.TestGoodCompile},	{"rubex.TestBadCompile", rubex.TestBadCompile},	{"rubex.TestMatch", rubex.TestMatch},	{"rubex.TestMatchFunction", rubex.TestMatchFunction},	{"rubex.TestReplaceAll", rubex.TestReplaceAll},	{"rubex.TestReplaceAllFunc", rubex.TestReplaceAllFunc},	{"rubex.TestQuoteMeta", rubex.TestQuoteMeta},	{"rubex.TestLiteralPrefix", rubex.TestLiteralPrefix},	{"rubex.TestNumSubexp", rubex.TestNumSubexp},
}

var benchmarks = []testing.InternalBenchmark{
	{"rubex.BenchmarkLiteral", rubex.BenchmarkLiteral},	{"rubex.BenchmarkNotLiteral", rubex.BenchmarkNotLiteral},	{"rubex.BenchmarkMatchClass", rubex.BenchmarkMatchClass},	{"rubex.BenchmarkMatchClass_InRange", rubex.BenchmarkMatchClass_InRange},	{"rubex.BenchmarkReplaceAll", rubex.BenchmarkReplaceAll},	{"rubex.BenchmarkAnchoredLiteralShortNonMatch", rubex.BenchmarkAnchoredLiteralShortNonMatch},	{"rubex.BenchmarkAnchoredLiteralLongNonMatch", rubex.BenchmarkAnchoredLiteralLongNonMatch},	{"rubex.BenchmarkAnchoredShortMatch", rubex.BenchmarkAnchoredShortMatch},	{"rubex.BenchmarkAnchoredLongMatch", rubex.BenchmarkAnchoredLongMatch},
}

var matchPat string
var matchRe *__regexp__.Regexp

func matchString(pat, str string) (result bool, err __os__.Error) {
	if matchRe == nil || matchPat != pat {
		matchPat = pat
		matchRe, err = __regexp__.Compile(matchPat)
		if err != nil {
			return
		}
	}
	return matchRe.MatchString(str), nil
}

func main() {
	testing.Main(matchString, tests, benchmarks)
}

