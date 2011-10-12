package rubex

/*
#cgo LDFLAGS: -L/usr/local/lib -lonig
#cgo CFLAGS: -I/usr/local/include
#include <oniguruma.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

int GetOnigErrorInfo(char *buf, int error_code, OnigErrorInfo *errorInfo) {
    return onig_error_code_to_str((unsigned char*)buf, error_code, errorInfo);
}

int NewOnigRegex( char *pattern, int pattern_length, int option,
                  OnigRegex *regex, OnigEncoding *encoding, OnigErrorInfo **error_info, char **error_buffer) {
    int ret = ONIG_NORMAL;
    int error_msg_len = 0;

    OnigUChar *pattern_start = (OnigUChar *) pattern;
    OnigUChar *pattern_end = (OnigUChar *) (pattern + pattern_length);

    *error_info = (OnigErrorInfo *) malloc(sizeof(OnigErrorInfo));
    memset(*error_info, 0, sizeof(OnigErrorInfo));

    *encoding = (void*)ONIG_ENCODING_UTF8;

    *error_buffer = (char*) malloc(ONIG_MAX_ERROR_MESSAGE_LEN * sizeof(char));

    memset(*error_buffer, 0, ONIG_MAX_ERROR_MESSAGE_LEN * sizeof(char));

    ret = onig_new(regex, pattern_start, pattern_end, (OnigOptionType)(option), *encoding, OnigDefaultSyntax, *error_info);
    
    if (ret != ONIG_NORMAL) {
        error_msg_len = onig_error_code_to_str((unsigned char*)(*error_buffer), ret, *error_info);
        if (error_msg_len >= ONIG_MAX_ERROR_MESSAGE_LEN) {
            error_msg_len = ONIG_MAX_ERROR_MESSAGE_LEN - 1;
        }
        (*error_buffer)[error_msg_len] = '\0';
    }
    return ret;
}
*/
import "C"

import (
  "utf8"
  "unsafe"
  "fmt"
)

type Regex struct {
  regex C.OnigRegex
  encoding C.OnigEncoding
  errorInfo *C.OnigErrorInfo
  errorBuf *C.char
}

func NewRegex(pattern string, option int) (re *Regex) {
  re = &Regex{}
  error_code := C.NewOnigRegex(C.CString(pattern), C.int(len(pattern)), C.int(option), &re.regex, &re.encoding, &re.errorInfo, &re.errorBuf)
  fmt.Printf("error_code: %d\n", error_code)
  
  if error_code != C.ONIG_NORMAL {
    fmt.Printf("error: %q\n", C.GoString(re.errorBuf))
  }
  
  return re
}

/*
func NewRegex(pattern string, option int) (re *Regex) {
  var regexPtr C.OnigRegex = nil
  var encoding = C.GetOnigEncodingUTF8()
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
*/

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

func (re *Regex) Free() {
  if re.regex != nil {
    C.onig_free(re.regex)
    re.regex = nil
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


