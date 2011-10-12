package rubex

/*
#cgo LDFLAGS: -L/usr/local/lib -lonig
#cgo CFLAGS: -I/usr/local/include
#include <oniguruma.h>
#include <stdlib.h>
#include <stdio.h>

OnigErrorInfo* NewOnigErrorInfo() {
    OnigErrorInfo *error_info = malloc(sizeof(OnigErrorInfo));
    return error_info;
}

void GetPatternStartAndEnd(const char* pattern, int length, OnigUChar **pattern_start, OnigUChar **pattern_end) {
    *pattern_start = (OnigUChar *) pattern;
    *pattern_end = (OnigUChar *) (pattern + length);
}

int GetOnigErrorInfo(char *buf, int error_code, OnigErrorInfo *errorInfo) {
    return onig_error_code_to_str((unsigned char*)buf, error_code, errorInfo);
}
*/
import "C"

import (
  "utf8"
  "unsafe"
  "fmt"
)

type Regex struct {
  regexPtr C.OnigRegex
  encoding C.OnigEncoding
  error *C.OnigErrorInfo
}

func NewRegex(pattern string, option int) (re *Regex) {
  var regexPtr C.OnigRegex = nil
  var encoding = C.onigenc_get_default_encoding()
  var error = C.NewOnigErrorInfo()
  re = &Regex{regexPtr: regexPtr, encoding: encoding, error: error}
  var patternStartPtr *C.OnigUChar = nil
  var patternEndPtr *C.OnigUChar = nil
  C.GetPatternStartAndEnd(C.CString(pattern), (C.int)(len(pattern)), &patternStartPtr, &patternEndPtr)
  //error_code := C.onig_new(&re.regexPtr, patternStartPtr, patternEndPtr, (C.OnigOptionType)(option), re.encoding, C.OnigDefaultSyntax, re.error)
  error_code := C.onig_new(&re.regexPtr, patternStartPtr, patternEndPtr, C.ONIG_OPTION_MULTILINE, re.encoding, C.OnigDefaultSyntax, re.error)
  if error_code != C.ONIG_NORMAL {
      error_buf := make([]byte, C.ONIG_MAX_ERROR_MESSAGE_LEN)
      error_ptr := (*C.char)(unsafe.Pointer(&error_buf[0]))
      error_len := C.GetOnigErrorInfo(error_ptr, error_code, re.error)
      error_info := string(error_buf[0:error_len])
      
      fmt.Printf("error: %q %v\n", error_info, re.error)
  }
  return re
}

func Quote(str string) (string) {
  uStr := utf8.NewString(str) //convert it to utf8
  newStr := make([]byte, len(str)*2)
  newStrOffset := 0
  
  for i := 0; i < uStr.RuneCount(); i ++ {
    v := uStr.At(i)
    if  v == int('[') || v == int(']') || v == int('{') || v == int('}') ||
        v == int('(') || v == int(')') || v == int('|') || v == int('-') || 
        v == int('*') || v == int('.') || v == int('\\') ||
        v == int('?') || v == int('+') || v == int('^') || v == int('$') ||
        v == int(' ') || v == int('#') {
      newStr[newStrOffset] = byte('\\'); newStrOffset += 1
      newStr[newStrOffset] = byte(v); newStrOffset += 1
    } else if v == int('\t') {
      newStr[newStrOffset] = byte('\\'); newStrOffset += 1
      newStr[newStrOffset] = byte('t'); newStrOffset += 1
    } else if v == int('\f') {
      newStr[newStrOffset] = byte('\\'); newStrOffset += 1
      newStr[newStrOffset] = byte('f'); newStrOffset += 1
    } else if v == int('\v') {
      newStr[newStrOffset] = byte('\\'); newStrOffset += 1
      newStr[newStrOffset] = byte('v'); newStrOffset += 1
    } else if v == int('\n') {
      newStr[newStrOffset] = byte('\\'); newStrOffset += 1
      newStr[newStrOffset] = byte('n'); newStrOffset += 1
    } else if v == int('\r') {
      newStr[newStrOffset] = byte('\\'); newStrOffset += 1
      newStr[newStrOffset] = byte('r'); newStrOffset += 1
    } else {
      newStrOffset += utf8.EncodeRune(newStr[newStrOffset:], v)
    }
  }
  return string(newStr[0:newStrOffset])
}

