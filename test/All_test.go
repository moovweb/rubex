package test

import (
  "rubex"
  "testing"
  "fmt"
  "regexp"
)

func TestQuote(t *testing.T) {
  strAscii := "yeah"
  if rubex.Quote(strAscii) != strAscii {
      t.Error("Quote ascii failed")
  }
  strAsciiWithSpecialChars := "Yeah [ 	"
  if rubex.Quote(strAsciiWithSpecialChars) != "Yeah\\ \\[\\ \\t" {
  	fmt.Printf("%q %d\n", rubex.Quote(strAsciiWithSpecialChars), len(rubex.Quote(strAsciiWithSpecialChars))) 
    t.Error("Quote ascii with special chars failed")
  }
}

func TestNewRegexpGoodPatten(t *testing.T) {
  pattern := "yeah"
  re, err := rubex.NewRegexp(pattern, 0)
  defer re.Free()
  if err != nil {
    t.Error("good pattern failed")
  }
}

func TestNewRegexpBadPatten(t *testing.T) {
  pattern := "yeah(abc"
  re, err := rubex.NewRegexp(pattern, 0)
  defer re.Free()
  if err == nil {
    t.Error("bad pattern should fail")
  }
}

func TestSimpleSearch(t *testing.T) {
  pattern := "a(.*)b|[e-f]+"
  str :=  "zzzzaffffffffb"
  re, err := rubex.NewRegexp(pattern, rubex.ONIG_OPTION_DEFAULT)
  defer re.Free()
  if err != nil {
    t.Error("good pattern failed")
  }
  fmt.Printf("%v\n", re.FindAllString(str, len(str)))
  re1, err := regexp.Compile(pattern)
  fmt.Printf("sys %v\n", re1.FindAllString(str, len(str)))
  fmt.Printf("sys %v\n", re1.FindAllStringSubmatch(str, len(str)))
}

func TestSimpleSearchMultMatches(t *testing.T) {
  pattern := "a(b*)"
  str :=  "abbaab"
  re, err := rubex.NewRegexp(pattern, rubex.ONIG_OPTION_DEFAULT)
  defer re.Free()
  if err != nil {
    t.Error("good pattern failed")
  }
  a := re.FindAllString(str, len(str))
  fmt.Printf("%v %d\n", a, len(a))
  re1, err := regexp.Compile(pattern)
  fmt.Printf("sys %v\n", re1.FindAllString(str, len(str)))
  fmt.Printf("sys %v\n", re1.FindAllStringSubmatch(str, len(str)))
}



func TestSimpleSearchNoResult(t *testing.T) {
  pattern := "c(.*)b|[g-h]+"
  str :=  "zzzzaffffffffb"
  re, err := rubex.NewRegexp(pattern, 0)
  defer re.Free()
  if err != nil {
    t.Error("good pattern failed")
  }
  fmt.Printf("%v\n", re.FindAllString(str, len(str)))
  re1, err := regexp.Compile(pattern)
  fmt.Printf("%v\n", re1.FindAllString(str, len(str)))
}
