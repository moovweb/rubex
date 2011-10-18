package rubex

/*
#include <stdlib.h>
#include <oniguruma.h>
#include "chelper.h"
*/
import "C"

import (
  "unsafe"
  "fmt"
  "os"
  "io"
  "utf8"
)

type strRange []int
const numMatchStartSize = 4
const numReadBufferStartSize = 256

type MatchData struct {
  //captures[i-1] is the i-th match -- there could be multiple non-overlapping matches for a given pattern
  //captures[i-1][0] gives the beginning and ending index of the i-th match
  //captures[i-1][j] (j >=1) gives the beginning and ending index of the j-th capture for the i-th match
  captures [][]strRange
  //namedCaptures["foo"] gives the j index of named capture "foo", then j can be used to get the beginning and ending index of the capture for this match
  namedCaptures map[string]int
}

type Regexp struct {
  pattern string
  regex C.OnigRegex
  region *C.OnigRegion
  encoding C.OnigEncoding
  errorInfo *C.OnigErrorInfo
  errorBuf *C.char
  //matchData *MatchData
}

func NewRegexp(pattern string, option int) (re *Regexp, err os.Error) {
  re = &Regexp{pattern: pattern}
  error_code := C.NewOnigRegex(C.CString(pattern), C.int(len(pattern)), C.int(option), &re.regex, &re.region, &re.encoding, &re.errorInfo, &re.errorBuf)
  if error_code != C.ONIG_NORMAL {
    err = os.NewError(C.GoString(re.errorBuf))
  } else {
    err = nil
    //re.matchData = &MatchData{}
    //re.matchData.captures = make([][]strRange, 0, numMatchStartSize)
    //re.matchData.namedCaptures = make(map[string]int)
  }
  return re, err
}

func Compile(str string) (*Regexp, os.Error) {
  return NewRegexp(str, ONIG_OPTION_DEFAULT)
}

func MustCompile(str string) *Regexp {
  regexp, error := NewRegexp(str, ONIG_OPTION_DEFAULT) 
  if error != nil {
    panic("regexp: compiling " + str + ": " + error.String())
  }
  return regexp
}

func (re *Regexp) Free() {
  if re.regex != nil {
    C.onig_free(re.regex)
    re.regex = nil
  }
  if re.region != nil {
    C.onig_region_free(re.region, 1)
  }
  if re.errorInfo != nil {
    C.free(unsafe.Pointer(re.errorInfo))
    re.errorInfo = nil
  }
  if re.errorBuf != nil {
    C.free(unsafe.Pointer(re.errorBuf))
    re.errorBuf = nil
  }
}

/*
func (re *Regexp) GetCaptureAt(at int) (sr strRange) {
  sr = nil
  if len(re.matchData.captures) > 0 && at < len(re.matchData.captures[0]) {
    sr = re.matchData.captures[0][at]
  }
  return
}

func (re *Regexp) GetCaptures()(srs []strRange) {
  srs = nil
  if len(re.matchData.captures) > 0 {
    srs = re.matchData.captures[0]
  }
  return
}

func (re *Regexp) GetAllCaptures()(srs [][]strRange) {
  return re.matchData.captures
}
*/
func (re *Regexp) getStrRange(ref int) (sr strRange) {
  sr = make([]int, 2)
  sr[0] = int(C.IntAt(re.region.beg, C.int(ref)))
  sr[1] = int(C.IntAt(re.region.end, C.int(ref)))
  return 
}

func (re *Regexp) processMatch() (captures []strRange) {
  //matchData := re.matchData
  num := (int (re.region.num_regs))
  if num <= 0 {
    panic("cannot have 0 captures when processing a match")
  }
  captures = make([]strRange, 0, num)
  //the first element indicates the beginning and ending indexes of the match
  //the rests are the beginning and ending indexes of the captures
  for i := 0; i < num; i ++ {
    sr := re.getStrRange(i)
    //sometimes we may encounter negative index, e.g. string: aacc, and pattern: a*(|(b))c*. A bug in onig? Ruby shows the same.
    //we should skip such 
    if sr[0] >= 0 && sr[1] >= 0 {
      captures = append(captures, sr)
    }
    
    fmt.Printf("in processMatch sr = %v num = %d\n", sr, num)
    fmt.Printf("in processMatch captures = %v\n", captures)
  }
  //matchData.captures = append(matchData.captures, captures)
  fmt.Printf("in processMatch captures = %v\n", captures)
  return
}

func (re *Regexp) find(b []byte, n int, offset int, deliver func([]strRange)) (err os.Error) {
  if n == 0 {
    b = []byte{0}
  }
  ptr := unsafe.Pointer(&b[0])
  pos := int(C.SearchOnigRegex((ptr), C.int(n), C.int(offset), C.int(ONIG_OPTION_DEFAULT), re.regex, re.region, re.encoding, re.errorInfo, re.errorBuf))
  if pos >= 0 {
    err = nil
    captures := re.processMatch()
    fmt.Printf("in find captures = %v\n", captures)
    deliver(captures)
  } else {
    err = os.NewError(C.GoString(re.errorBuf))
  }
  return
}

func (re *Regexp) findAll(b []byte, n int, deliver func([]strRange)) (err os.Error) {
  if n < 0 {
    n = len(b)
  }
  var captures []strRange
  err = nil; offset := 0
  hasMatched := false
  for ;err == nil && offset <= n; {
    err = re.find(b, n, offset, func(kaps []strRange) {
      captures = kaps
    })
    fmt.Printf("captures = %v err = %v\n", captures, err)
    if err == nil {
      hasMatched = true
      //we need to adjust the captures' indexes by offset because the search starts at offset
      //captures = adjustStrRangeByOffset(captures, offset)
      deliver(captures)
      //remember the first capture is in fact the current match
      match := captures[0]
      //move offset to the ending index of the current match and prepare to find the next non-overlapping match
      offset = match[1]
      //if match[0] == match[1], it means the current match does not advance the search. we need to exit the loop to avoid getting stuck here.
      if match[0] == match[1] {
        offset += 1 //TODO handle encoding
      }
      fmt.Printf("offset = %d\n", offset)
    } else {
      break
    }
  }
  //if there has been a match, we should not return error
  if hasMatched {
    err = nil
  }
  if err != nil {
    fmt.Printf("find Error: %q pattern: %v str: %v\n", err, re, b)
  }
  
  fmt.Printf("find Error: %q pattern: %v str: %v\n", err, re, b)
  return
}

func (re *Regexp) FindIndex(b []byte) (loc []int) {
  var captures []strRange
  err := re.find(b, len(b), 0, func(caps []strRange) {
    captures = caps
  })
  if err == nil {
    loc = captures[0]
  } else {
    loc = nil
  }
  return
}

func (re *Regexp) Find(b []byte) []byte {
  match := re.FindIndex(b)
  if match == nil {
    return nil
  }
  return b[match[0]:match[1]]
}

func (re *Regexp) FindString(s string) string {
  b := []byte(s)
  mb := re.Find(b)
  if mb == nil {
    return ""
  }
  return string(mb)
}

func (re *Regexp) FindStringIndex(s string) []int {
  b := []byte(s)
  return re.FindIndex(b)
}

func (re *Regexp) FindAllIndex(b []byte, n int) [][]int { 
  matches := make([][]int, 0, numMatchStartSize)
  re.findAll(b, n, func(captures []strRange) {
    match := captures[0]
    matches = append(matches, match)
  })
  if len(matches) == 0 {
    return nil  
  }
  return matches
}

func (re *Regexp) FindAll(b []byte, n int) [][]byte {
  matches := re.FindAllIndex(b, n)
  if matches == nil {
    return nil
  }
  matchBytes := make([][]byte, 0, len(matches))
  for _, match := range matches {
    matchBytes = append(matchBytes, b[match[0]:match[1]])
  }
  return matchBytes
}

func (re *Regexp) FindAllString(s string, n int) []string {
  b := []byte(s)
  matches := re.FindAllIndex(b, n)
  if matches == nil {
    return nil
  }
  matchStrings := make([]string, 0, len(matches))
  for _, match := range matches {
    matchStrings = append(matchStrings, string(b[match[0]:match[1]]))
  }
  return matchStrings

}

func (re *Regexp) FindAllStringIndex(s string, n int) [][]int {
  b := []byte(s)
  return re.FindAllIndex(b, n)
}

func (re *Regexp) findSubmatchIndex(b []byte) (captures []strRange) {
  fmt.Printf("findSubmatchIndex %q %q\n", string(b), re)
  err := re.find(b, len(b), 0, func(caps []strRange) {
    captures = caps
    fmt.Printf("caps: %v\n", caps)
  })
  fmt.Printf("captures: %v\n", captures)
  if err != nil {
    captures = nil
  }
  return 
}

func flattenCaptures(captures []strRange) []int {
  if captures == nil {
    return nil
  }
  flatCaptures := make([]int, 0, len(captures)*2)
  for _, cap := range captures {
    flatCaptures = append(flatCaptures, cap[0])
    flatCaptures = append(flatCaptures, cap[1])
  }
  if len(flatCaptures) == 0 {
    return nil
  }
  return flatCaptures
}

func (re *Regexp) FindSubmatchIndex(b []byte) []int {
  return flattenCaptures(re.findSubmatchIndex(b))
}


func (re *Regexp) FindSubmatch(b []byte) [][]byte {
  captures := re.findSubmatchIndex(b)
  if captures == nil {
    return nil
  }
  fmt.Printf("findSubmatchIndex got captures = %v\n", captures)
  length := len(captures)
  results := make([][]byte, 0, length)
  for i:= 0; i < length; i ++ {
    cap := captures[i]
    results = append(results, b[cap[0]:cap[1]])
  }
  if len(results) == 0 {
    return nil
  }
  fmt.Printf("findSubmatchIndex got results = %v\n", results)
  return results
}

func (re *Regexp) FindStringSubmatch(s string) []string {
  b := []byte(s)
  captures := re.findSubmatchIndex(b)
  if captures == nil {
    return nil
  }
  length := len(captures)
  results := make([]string, 0, length)
  for i:= 0; i < length; i ++ {
    cap := captures[i]
    results = append(results, string(b[cap[0]:cap[1]]))
  }
  if len(results) == 0 {
    return nil
  }
  return results
}

func (re *Regexp) FindStringSubmatchIndex(s string) []int {
  b := []byte(s)
  return re.FindSubmatchIndex(b)  
}

func (re *Regexp) findAllSubmatchIndex(b []byte, n int) [][]strRange {
  allCaptures := make([][]strRange, 0, numMatchStartSize) 
  re.findAll(b, n, func(caps []strRange) {
    fmt.Printf("captures: %v\n", caps)
    allCaptures = append(allCaptures, caps)
  })

  fmt.Printf("allCaptures: %v\n", allCaptures)
  if len(allCaptures) == 0 { 
    return nil
  }
  return allCaptures
}

func (re *Regexp) FindAllSubmatchIndex(b []byte, n int) [][]int {
  allCaptures := re.findAllSubmatchIndex(b, n)
  if len(allCaptures) == 0 {
    return nil
  }
  allFlatCaptures := make([][]int, 0, len(allCaptures))
  for _, captures := range allCaptures {
    flatCaptures := flattenCaptures(captures)
    allFlatCaptures = append(allFlatCaptures, flatCaptures)
  }
  if len(allFlatCaptures) == 0 {
    return nil
  }
  return allFlatCaptures
}

func (re *Regexp) FindAllSubmatch(b []byte, n int) [][][]byte {
  allCaptures := re.findAllSubmatchIndex(b, n)
  if len(allCaptures) == 0 {
    return nil
  }
  allCapturedBytes := make([][][]byte, 0, len(allCaptures))
  for _, captures := range allCaptures {
    capturedBytes := make([][]byte, 0, len(captures))
    for _, cap := range captures {
      capturedBytes = append(capturedBytes, b[cap[0]:cap[1]])
    }
    allCapturedBytes = append(allCapturedBytes, capturedBytes)
  }
  
  if len(allCapturedBytes) == 0 {
    return nil
  }
  return allCapturedBytes
}

func (re *Regexp) FindAllStringSubmatch(s string, n int) [][]string {
  b := []byte(s)
  allCaptures := re.findAllSubmatchIndex(b, n)
  if len(allCaptures) == 0 {
    return nil
  }
  allCapturedStrings := make([][]string, 0, len(allCaptures))
  for _, captures := range allCaptures {
    capturedStrings := make([]string, 0, len(captures))
    for _, cap := range captures {
      capturedStrings = append(capturedStrings, string(b[cap[0]:cap[1]]))
    }
    allCapturedStrings = append(allCapturedStrings, capturedStrings)
  }
  
  if len(allCapturedStrings) == 0 {
    return nil
  }
  return allCapturedStrings
}

func (re *Regexp) FindAllStringSubmatchIndex(s string, n int) [][]int {
  b := []byte(s)
  return re.FindAllSubmatchIndex(b, n)
}

func (re *Regexp) Match(b []byte) bool {
  err := re.find(b, len(b), 0, func(caps []strRange) {})
  return err == nil
}

func (re *Regexp) MatchString(s string) bool {
  b := []byte(s)
  return re.Match(b)
}

func (re *Regexp) NumSubexp() int {
  return (int)(C.onig_number_of_captures(re.regex))
}

func fillCapturedValues(repl []byte, capturedBytes [][]byte) []byte {
  newRepl := make([]byte, 0, len(repl) * 3)
  inEscapeMode := false
  for _, ch := range repl {
    if inEscapeMode && ch <= byte('9') && byte('1') <= ch {
      capNum := int(ch - byte('0'))
      if capNum > len(capturedBytes) - 1  {
        panic(fmt.Sprintf("invalid capture number: %d", capNum))
      }
      capBytes := capturedBytes[capNum]
      for _, c := range capBytes {
        newRepl = append(newRepl, c)
      }
    } else if inEscapeMode {
      newRepl = append(newRepl, '\\')
      newRepl = append(newRepl, ch)
    } else if ch != '\\' {
      newRepl = append(newRepl, ch)
    }
    if ch == byte('\\') || inEscapeMode {
      inEscapeMode = !inEscapeMode
    }
  }
  //fmt.Printf("newRepl: %v\n", string(newRepl))
  return newRepl
}

func (re *Regexp) replaceAll(src, repl []byte, replFunc func([]byte, [][]byte) []byte) []byte {
  allCaptures := re.findAllSubmatchIndex(src, len(src))
  fmt.Printf("allCaptures: %v\n", allCaptures)
  if allCaptures == nil {
    return src
  }
  dest := make([]byte, 0, len(src))
  for i, captures := range allCaptures {
    capturedBytes := make([][]byte, 0, len(captures))
    for _, cap := range captures {
      capturedBytes = append(capturedBytes, src[cap[0]:cap[1]])
    }
    newRepl := replFunc(repl, capturedBytes)
    match := captures[0]
    prevEnd := 0
    if i > 0 {
      prevMatch := allCaptures[i-1][0]
      prevEnd = prevMatch[1]
    }
    if match[0] > prevEnd {
      for _,c := range src[prevEnd:match[0]] {
        dest = append(dest, c)
      }
    }
    for _,c := range newRepl {
      dest = append(dest, c)
    }
  }
  lastEnd := allCaptures[len(allCaptures)-1][0][1]
  if lastEnd < len(src) { 
    if lastEnd < len(src) {
      for _, c:= range src[lastEnd:] {
        dest = append(dest, c)
      }
    }
  }
  return dest
}

func (re *Regexp) ReplaceAll(src, repl []byte) []byte {
  return re.replaceAll(src, repl, fillCapturedValues)
}

func (re *Regexp) ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte {
  return re.replaceAll(src, []byte(""), func(_ []byte, capturedBytes [][]byte) []byte {
    return repl(capturedBytes[0])
  })
}

func (re *Regexp) ReplaceAllString(src, repl string) string {
  return string(re.ReplaceAll([]byte(src), []byte(repl)))
}

func (re *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string {
  srcB := []byte(src)
  destB := re.replaceAll(srcB, []byte(""), func(_ []byte, capturedBytes [][]byte) []byte {
    return []byte(repl(string(capturedBytes[0])))
  })
  return string(destB)
}

func (re *Regexp) String() string {
  return re.pattern
}

func grow_buffer(b []byte, offset int, n int) []byte {
  fmt.Printf("offset = %d cap %d\n", offset, cap(b))
  if offset+n > cap(b) {
    buf := make([]byte, 2*cap(b)+n)
    copy(buf, b[:offset])
    return buf
  }
  return b
}

func fromReader(r io.RuneReader) []byte {
  b := make([]byte, numReadBufferStartSize)
  offset := 0
  var err os.Error = nil
  for ;err == nil; {
    rune, runeWidth, err := r.ReadRune()
    fmt.Printf("runeWidth = %d err = %v\n", runeWidth, err)
    if err == nil {
      fmt.Printf("runeWidth = %d\n", runeWidth)
      b = grow_buffer(b, offset, runeWidth)
      writeWidth := utf8.EncodeRune(b[offset:], rune)
      if runeWidth != writeWidth {
        panic("reading rune width not equal to the written rune width")
      }
      offset += writeWidth
    } else {
      break
    }
  }
  return b[:offset]
}

func (re *Regexp) FindReaderIndex(r io.RuneReader) []int {
  b := fromReader(r)
  return re.FindIndex(b)
}

func (re *Regexp) FindReaderSubmatchIndex(r io.RuneReader) []int {
  b := fromReader(r)
  return re.FindSubmatchIndex(b)
}

func (re *Regexp) MatchReader(r io.RuneReader) bool {
  b := fromReader(r)
  return re.Match(b)  
}

func (re *Regexp) LiteralPrefix() (prefix string, complete bool) {
  //no easy way to implement this
  return "", false
}
