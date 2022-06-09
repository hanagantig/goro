package util

import "unicode"

// Camelize an uppercased word
func Camelize(word string) (camelized string) {
	for pos, ru := range []rune(word) {
		if pos > 0 {
			camelized += string(unicode.ToLower(ru))
		} else {
			camelized += string(unicode.ToUpper(ru))
		}
	}
	return
}
