package rubex

import "utf8"

func Quote(str string) string {
	uStr := utf8.NewString(str) //convert it to utf8
	newStr := make([]byte, len(str)*2)
	newStrOffset := 0

	for i := 0; i < uStr.RuneCount(); i++ {
		v := uStr.At(i)
		if v == int('[') || v == int(']') || v == int('{') || v == int('}') ||
			v == int('(') || v == int(')') || v == int('|') || v == int('-') ||
			v == int('*') || v == int('.') || v == int('\\') ||
			v == int('?') || v == int('+') || v == int('^') || v == int('$') ||
			v == int(' ') || v == int('#') {
			newStr[newStrOffset] = byte('\\')
			newStrOffset += 1
			newStr[newStrOffset] = byte(v)
			newStrOffset += 1
		} else if v == int('\t') {
			newStr[newStrOffset] = byte('\\')
			newStrOffset += 1
			newStr[newStrOffset] = byte('t')
			newStrOffset += 1
		} else if v == int('\f') {
			newStr[newStrOffset] = byte('\\')
			newStrOffset += 1
			newStr[newStrOffset] = byte('f')
			newStrOffset += 1
		} else if v == int('\v') {
			newStr[newStrOffset] = byte('\\')
			newStrOffset += 1
			newStr[newStrOffset] = byte('v')
			newStrOffset += 1
		} else if v == int('\n') {
			newStr[newStrOffset] = byte('\\')
			newStrOffset += 1
			newStr[newStrOffset] = byte('n')
			newStrOffset += 1
		} else if v == int('\r') {
			newStr[newStrOffset] = byte('\\')
			newStrOffset += 1
			newStr[newStrOffset] = byte('r')
			newStrOffset += 1
		} else {
			newStrOffset += utf8.EncodeRune(newStr[newStrOffset:], v)
		}
	}
	return string(newStr[0:newStrOffset])
}
