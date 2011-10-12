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
  rubex.NewRegex(pattern, 0)
  pattern = "yeah(abc"
  rubex.NewRegex(pattern, 0)

}
