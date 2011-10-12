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

func TestNewRegex(t *testing.T) {
  pattern := "yeah"
  re := rubex.NewRegex(pattern, 0)
  re.Free()
  pattern = "yeah(abc"
  re = rubex.NewRegex(pattern, 0)
  re.Free()
}
