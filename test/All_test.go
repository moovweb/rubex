package test

import (
  "rubex"
  "testing"
  "fmt"
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
  pattern := "yeah"
  str :=  "fine... yeah"
  re, err := rubex.NewRegexp(pattern, 0)
  defer re.Free()
  if err != nil {
    t.Error("good pattern failed")
  }
  re.Find([]byte(str))
}


