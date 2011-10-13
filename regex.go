package rubex

/*
#cgo LDFLAGS: -L/usr/local/lib -lonig
#cgo CFLAGS: -I/usr/local/include
#include <stdlib.h>
#include <oniguruma.h>
#include "chelper.h"
*/
import "C"

import (
  "unsafe"
  "fmt"
  "os"
)

type Regexp struct {
  regex C.OnigRegex
  region *C.OnigRegion
  encoding C.OnigEncoding
  errorInfo *C.OnigErrorInfo
  errorBuf *C.char
}

func NewRegexp(pattern string, option int) (re *Regexp, err os.Error) {
  re = &Regexp{}
  error_code := C.NewOnigRegex(C.CString(pattern), C.int(len(pattern)), C.int(option), &re.regex, &re.region, &re.encoding, &re.errorInfo, &re.errorBuf)
  if error_code != C.ONIG_NORMAL {
    fmt.Printf("error: %q\n", C.GoString(re.errorBuf))
    err = os.NewError(C.GoString(re.errorBuf))
  } else {
    err = nil
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

func (re *Regexp) find(b []byte, n int) (pos int, err os.Error) {
  ptr := unsafe.Pointer(&b[0])
  pos = int(C.SearchOnigRegex((ptr), C.int(n), C.int(ONIG_OPTION_DEFAULT), re.regex, re.region, re.encoding, re.errorInfo, re.errorBuf))
  if pos >= 0 {
    err = nil
  } else {
    err = os.NewError(C.GoString(re.errorBuf))
  }
  return pos, err
}

func (re *Regexp) findall(b []byte, n int, deliver func(int, int)) (err os.Error) {
  offset := 0
  bp := b[offset:]
  _, err = re.find(bp, n - offset)
  if err == nil {
    beg := int(C.IntAt(re.region.beg, 0))
    end := int(C.IntAt(re.region.end, 0))
    deliver(beg+offset, end+offset)
    offset = offset + end
  }
  for ;err == nil && offset < n; {
    bp = b[offset:]
    _, err = re.find(bp, n - offset)
    if err == nil {
      beg := int(C.IntAt(re.region.beg, 0))
      end := int(C.IntAt(re.region.end, 0))
      deliver(beg+offset, end+offset)
      offset = offset + end
    }
  }
  return err
}

const startSize = 10

func (re *Regexp) FindAll(b []byte, n int) [][]byte {
  results := make([][]byte, 0, startSize)
  re.findall(b, n, func(beg int, end int) {
    results = append(results, b[beg:end])
  })
  if len(results) == 0 {
    return nil  
  }
  return results
}

func (re *Regexp) FindAllString(s string, n int) []string {
  b := []byte(s)
  results := make([]string, 0, startSize)
  re.findall(b, n, func(beg int, end int) {
    results = append(results, string(b[beg:end]))
  })
  
  if len(results) == 0 {
    return nil  
  }
  return results
}

func (re *Regexp) FindAllIndex(b []byte, n int) [][]int { 
  results := make([][]int, 0, startSize)
  re.findall(b, n, func(beg int, end int) {
    m := make([]int,2)
    m[0] = beg
    m[1] = end
    results = append(results, m)
  })
  if len(results) == 0 {
    return nil  
  }
  return results
}

func (re *Regexp) FindAllStringIndex(s string, n int) [][]int {
  b := []byte(s)
  return re.FindAllIndex(b, n)
}

func (re *Regexp) FindAllStringSubmatch(s string, n int) [][]string {
  return nil
}

func (re *Regexp) Find(b []byte) []byte {
  _, err := re.find(b, len(b))
  if err == nil {
    num_matches := (int (re.region.num_regs))
    if num_matches > 0 {
      beg := int(C.IntAt(re.region.beg, 0))
      end := int(C.IntAt(re.region.end, 0))
      return b[beg:end]
    }
  }
  return nil  
 
}
