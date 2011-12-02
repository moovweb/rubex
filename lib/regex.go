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
	"bytes"
	"log"
	"utf8"
	"sync"
)

type strRange []int

const numMatchStartSize = 4
const numReadBufferStartSize = 256

var mutex sync.Mutex

type namedCapture struct {
	id int
	capture [][]byte
}

type Regexp struct {
	pattern       string
	regex         C.OnigRegex
	region        *C.OnigRegion
	encoding      C.OnigEncoding
	errorInfo     *C.OnigErrorInfo
	errorBuf      *C.char
	matches       [][]int
	matchIndex    int
	numCapturesInPattern int
	//TODO change this to an array
	namedCaptures map[string]*namedCapture
}

func NewRegexp(pattern string, option int) (re *Regexp, err os.Error) {
	re = &Regexp{pattern: pattern}
	patternCharPtr := C.CString(pattern)
	defer C.free(unsafe.Pointer(patternCharPtr))

	mutex.Lock()
	defer mutex.Unlock()
	error_code := C.NewOnigRegex(patternCharPtr, C.int(len(pattern)), C.int(option), &re.regex, &re.region, &re.encoding, &re.errorInfo, &re.errorBuf)
	if error_code != C.ONIG_NORMAL {
		err = os.NewError(C.GoString(re.errorBuf))
	} else {
		err = nil
		numNamedGroups := int(C.onig_number_of_names(re.regex))
		if numNamedGroups > 0 {
			re.namedCaptures = make(map[string]*namedCapture)
			bufferSize := len(pattern)*2
			nameBuffer := make([]byte, bufferSize)
			groupNumbers := make([]int, numNamedGroups)
			bufferPtr := unsafe.Pointer(&nameBuffer[0])
			numbersPtr := unsafe.Pointer(&groupNumbers[0])
			length := int(C.GetCaptureNames(re.regex, bufferPtr, (C.int)(bufferSize), (*C.int)(numbersPtr)))
			if length > 0  {
				namesAsBytes := bytes.Split(nameBuffer[:length], ([]byte)(";"))
				if len(namesAsBytes) != numNamedGroups {
					log.Fatalf("the number of named groups (%d) does not match the number names found (%d)\n", numNamedGroups, len(namesAsBytes))
				}
				for i, nameAsBytes := range(namesAsBytes) {
					name := string(nameAsBytes)
					nc := &namedCapture{id: groupNumbers[i]} 
					nc.capture = make([][]byte, numMatchStartSize)
					re.namedCaptures[name] = nc
				}
			}
		}
		numCapturesInPattern := int(C.onig_number_of_captures(re.regex)) + 1
		re.matches = make([][]int, numMatchStartSize)
		for i := 0; i < numMatchStartSize; i ++ {
			re.matches[i] = make([]int, numCapturesInPattern*2)
		}
		re.numCapturesInPattern = numCapturesInPattern
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

func CompileWithOption(str string, option int) (*Regexp, os.Error) {
	return NewRegexp(str, option)
}

func MustCompileWithOption(str string, option int) *Regexp {
	regexp, error := NewRegexp(str, option)
	if error != nil {
		panic("regexp: compiling " + str + ": " + error.String())
	}
	return regexp
}

func (re *Regexp) Free() {
	mutex.Lock()
	if re.regex != nil {
		C.onig_free(re.regex)
		re.regex = nil
	}
	if re.region != nil {
		C.onig_region_free(re.region, 1)
	}
	mutex.Unlock()
	if re.errorInfo != nil {
		C.free(unsafe.Pointer(re.errorInfo))
		re.errorInfo = nil
	}
	if re.errorBuf != nil {
		C.free(unsafe.Pointer(re.errorBuf))
		re.errorBuf = nil
	}
}

func (re *Regexp) groupNameToId(name string) (id int) {
	if re.namedCaptures == nil {
		return ONIGERR_UNDEFINED_NAME_REFERENCE
	}

	//note that the Id (or Reference number) of a named capture is never 0
	if re.namedCaptures[name] == nil {
		nameCharPtr := C.CString(name)
		defer C.free(unsafe.Pointer(nameCharPtr))
		id = int(C.LookupOnigCaptureByName(nameCharPtr, C.int(len(name)), re.regex, re.region))
		re.namedCaptures[name] = &namedCapture{id:id}
		return
	}
	id = re.namedCaptures[name].id
	return
}


func (re *Regexp) processMatch(numCaptures int) (match []int) {
	if numCaptures <= 0 {
		panic("cannot have 0 captures when processing a match")
	}
	return re.matches[re.matchIndex][:numCaptures*2]
}

func (re *Regexp) find(b []byte, n int, offset int) (match []int) {
	if n == 0 {
		b = []byte{0}
	}
	ptr := unsafe.Pointer(&b[0])

	capturesPtr := unsafe.Pointer(&(re.matches[re.matchIndex][0]))
	numCaptures := 0
	numCapturesPtr := unsafe.Pointer(&numCaptures)
	pos := int(C.SearchOnigRegex((ptr), C.int(n), C.int(offset), C.int(ONIG_OPTION_DEFAULT), re.regex, re.region, re.errorInfo, (*C.char)(nil), (*C.int)(capturesPtr), (*C.int)(numCapturesPtr)))
	if pos >= 0 {
		if numCaptures <= 0 {
			panic("cannot have 0 captures when processing a match")
		}
		match = re.matches[re.matchIndex][:numCaptures*2]
	}
	return
}

func (re *Regexp) match(b []byte, n int, offset int) bool {
	if n == 0 {
		b = []byte{0}
	}
	ptr := unsafe.Pointer(&b[0])
	pos := int(C.SearchOnigRegex((ptr), C.int(n), C.int(offset), C.int(ONIG_OPTION_DEFAULT), re.regex, re.region, re.errorInfo, (*C.char)(nil), (*C.int)(nil), (*C.int)(nil)))
	return pos >= 0
}

func (re *Regexp) findAll(b []byte, n int) (matches [][]int) {
	if n < 0 {
		n = len(b)
	}
	offset := 0
	re.matchIndex = 0
	for offset <= n {
		if re.matchIndex >= len(re.matches) {
			re.matches = append(re.matches, make([]int, re.numCapturesInPattern*2))
			for _, value := range(re.namedCaptures) {
				value.capture = append(value.capture, nil)
			}
		}
		if match := re.find(b, n, offset); len(match) > 0 {
			if re.namedCaptures != nil {
				for _, value := range(re.namedCaptures) {
					beg := match[value.id*2]
					end := match[value.id*2+1]
					value.capture[re.matchIndex] = b[beg:end]
				}
			}
			re.matchIndex += 1
			//move offset to the ending index of the current match and prepare to find the next non-overlapping match
			offset = match[1]
			//if match[0] == match[1], it means the current match does not advance the search. we need to exit the loop to avoid getting stuck here.
			if match[0] == match[1] {
				if offset < n {
					//there are more bytes, so move offset by a word
					_, width := utf8.DecodeRune(b[offset:])
					offset += width
				} else {
					//search is over, exit loop
					break
				}
			}
		} else {
			break
		}
	}
	matches = re.matches[:re.matchIndex]
	return
}

func (re *Regexp) FindIndex(b []byte) []int {
	re.matchIndex = 0
	match := re.find(b, len(b), 0)
	if len(match) == 0 {
		return nil
	}
	return match[:2]
}

func (re *Regexp) Find(b []byte) []byte {
	loc := re.FindIndex(b)
	if loc == nil {
		return nil
	}
	return b[loc[0]:loc[1]]
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
	matches := re.findAll(b, n)
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

func (re *Regexp) findSubmatchIndex(b []byte) (match []int) {
	re.matchIndex = 0
	match = re.find(b, len(b), 0)
	return
}

func (re *Regexp) FindSubmatchIndex(b []byte) []int {
	match := re.findSubmatchIndex(b)
	if len(match) == 0 {
		return nil
	}
	return match
}

func (re *Regexp) FindSubmatch(b []byte) [][]byte {
	match := re.findSubmatchIndex(b)
	if match == nil {
		return nil
	}
	length := len(match)/2
	if length == 0 {
		return nil
	}
	results := make([][]byte, 0, length)
	for i := 0; i < length; i ++ {
		results = append(results, b[match[2*i]:match[2*i+1]])
	}
	return results
}

func (re *Regexp) FindStringSubmatch(s string) []string {
	b := []byte(s)
	match := re.findSubmatchIndex(b)
	if match == nil {
		return nil
	}
	length := len(match)/2
	if length == 0 {
		return nil
	}

	results := make([]string, 0, length)
	for i := 0; i < length; i ++ {
		results = append(results, string(b[match[2*i]:match[2*i+1]]))
	}
	return results
}

func (re *Regexp) FindStringSubmatchIndex(s string) []int {
	b := []byte(s)
	return re.FindSubmatchIndex(b)
}

func (re *Regexp) FindAllSubmatchIndex(b []byte, n int) [][]int {
	matches := re.findAll(b, n)
	if len(matches) == 0 {
		return nil
	}
	return matches
}

func (re *Regexp) FindAllSubmatch(b []byte, n int) [][][]byte {
	matches := re.findAll(b, n)
	if len(matches) == 0 {
		return nil
	}
	allCapturedBytes := make([][][]byte, 0, len(matches))
	for _, match := range matches {
		length := len(match)/2
		capturedBytes := make([][]byte, 0, length)
		for i := 0; i < length; i++ {
			capturedBytes = append(capturedBytes, b[match[2*i]:match[2*i+1]])
		}
		allCapturedBytes = append(allCapturedBytes, capturedBytes)
	}

	return allCapturedBytes
}

func (re *Regexp) FindAllStringSubmatch(s string, n int) [][]string {
	b := []byte(s)
	matches := re.findAll(b, n)
	if len(matches) == 0 {
		return nil
	}
	allCapturedStrings := make([][]string, 0, len(matches))
	for _, match := range matches {
		length := len(match)/2
		capturedStrings := make([]string, 0, length)
		for i := 0; i < length; i++ {
			capturedStrings = append(capturedStrings, string(b[match[2*i]:match[2*i+1]]))
		}
		allCapturedStrings = append(allCapturedStrings, capturedStrings)
	}
	return allCapturedStrings
}

func (re *Regexp) FindAllStringSubmatchIndex(s string, n int) [][]int {
	b := []byte(s)
	return re.FindAllSubmatchIndex(b, n)
}

func (re *Regexp) Match(b []byte) bool {
	return re.match(b, len(b), 0)
}

func (re *Regexp) MatchString(s string) bool {
	b := []byte(s)
	return re.Match(b)
}

func (re *Regexp) NumSubexp() int {
	return (int)(C.onig_number_of_captures(re.regex))
}

func (re *Regexp) getNamedCapture(name []byte, capturedBytes [][]byte) []byte {
	nameStr := string(name)
	capNum := re.groupNameToId(nameStr)
	if capNum < 0 || capNum >= len(capturedBytes) {
		panic(fmt.Sprintf("capture group name (%q) has error\n", nameStr))
	}
	return capturedBytes[capNum]
}

func (re *Regexp) getNumberedCapture(num int, capturedBytes [][]byte) []byte {
	//when named capture groups exist, numbered capture groups returns ""
	if re.namedCaptures == nil && num <= (len(capturedBytes)-1) && num >= 0 {
		return capturedBytes[num]
	}
	return ([]byte)("")
}

func fillCapturedValues(re *Regexp, repl []byte, capturedBytes [][]byte) []byte {
	replLen := len(repl)
	newRepl := make([]byte, 0, replLen*3)
	inEscapeMode := false
	inGroupNameMode := false
	groupName := make([]byte, 0, replLen)
	for index := 0; index < replLen; index += 1 {
		ch := repl[index]
		if inGroupNameMode && ch == byte('<') {
		} else if inGroupNameMode && ch == byte('>') {
			inGroupNameMode = false
			capBytes := re.getNamedCapture(groupName, capturedBytes)
			newRepl = append(newRepl, capBytes...)
			groupName = groupName[:0] //reset the name
		} else if inGroupNameMode {
			groupName = append(groupName, ch)
		} else if inEscapeMode && ch <= byte('9') && byte('1') <= ch {
			capNum := int(ch - byte('0'))
			capBytes := re.getNumberedCapture(capNum, capturedBytes)
			newRepl = append(newRepl, capBytes...)
		} else if inEscapeMode && ch == byte('k') && (index+1) < replLen && repl[index+1] == byte('<') {
			inGroupNameMode = true
			inEscapeMode = false
			index += 1 //bypass the next char '<'
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
	return newRepl
}

func (re *Regexp) replaceAll(src, repl []byte, replFunc func(*Regexp, []byte, [][]byte) []byte) []byte {
	matches := re.findAll(src, len(src))
	if len(matches) == 0 {
		return src
	}
	dest := make([]byte, 0, len(src))
	for i, match := range matches {
		length := len(match)/2
		capturedBytes := make([][]byte, 0, length)
		for i := 0; i < length; i ++ {
			capturedBytes = append(capturedBytes, src[match[2*i]:match[2*i+1]])
		}
		newRepl := replFunc(re, repl, capturedBytes)
		prevEnd := 0
		if i > 0 {
			prevMatch := matches[i-1][:2]
			prevEnd = prevMatch[1]
		}
		if match[0] > prevEnd {
			dest = append(dest, src[prevEnd:match[0]]...)
		}
		dest = append(dest, newRepl...)
	}
	lastEnd := matches[len(matches)-1][1]
	if lastEnd < len(src) {
		if lastEnd < len(src) {
			dest = append(dest, src[lastEnd:]...)
		}
	}
	return dest
}

func (re *Regexp) ReplaceAll(src, repl []byte) []byte {
	return re.replaceAll(src, repl, fillCapturedValues)
}

func (re *Regexp) ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte {
	return re.replaceAll(src, []byte(""), func(_ *Regexp, _ []byte, capturedBytes [][]byte) []byte {
		return repl(capturedBytes[0])
	})
}

func (re *Regexp) ReplaceAllString(src, repl string) string {
	return string(re.ReplaceAll([]byte(src), []byte(repl)))
}

func (re *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string {
	srcB := []byte(src)
	destB := re.replaceAll(srcB, []byte(""), func(_ *Regexp, _ []byte, capturedBytes [][]byte) []byte {
		return []byte(repl(string(capturedBytes[0])))
	})
	return string(destB)
}

func (re *Regexp) String() string {
	return re.pattern
}

func grow_buffer(b []byte, offset int, n int) []byte {
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
	for err == nil {
		rune, runeWidth, err := r.ReadRune()
		if err == nil {
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

func MatchString(pattern string, s string) (matched bool, error os.Error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(s), nil
}

func (re *Regexp) Gsub(src, repl string) string {
	srcBytes := ([]byte)(src)
	replBytes := ([]byte)(repl)
	replaced := re.replaceAll(srcBytes, replBytes, fillCapturedValues)
	return string(replaced)
}

func (re *Regexp) GetAllNamedCaptures(index int) (namedCaptures map[string]string) {
	if re.namedCaptures == nil || index >= re.matchIndex {
		return nil
	}
	namedCaptures = make(map[string]string)
	for name, value := range(re.namedCaptures) {
		namedCaptures[name] = string(value.capture[index])
	}
	return
}

func (re *Regexp) GsubFunc(src string, replFunc func(*Regexp, []string) string) string {
	srcBytes := ([]byte)(src)
	replaced := re.replaceAll(srcBytes, nil, func(re *Regexp, _ []byte, capturedBytes [][]byte) []byte {
		numCaptures := len(capturedBytes)
		capturedStrings := make([]string, numCaptures)
		for index, capBytes := range capturedBytes {
			capturedStrings[index] = string(capBytes)
		}
		return ([]byte)(replFunc(re, capturedStrings))
	})
	return string(replaced)
}
