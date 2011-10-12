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

func (re *Regexp) Find(b []byte) []byte {
  ptr := unsafe.Pointer(&b[0])
  C.SearchOnigRegex((ptr), C.int(len(b)), C.int(ONIG_OPTION_DEFAULT), re.regex, re.region, re.encoding, re.errorInfo, re.errorBuf);
  return b  
}

