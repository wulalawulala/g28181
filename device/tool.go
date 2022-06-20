package device

import (
	"bytes"
	"encoding/xml"
	"math/rand"
	"time"

	"golang.org/x/net/html/charset"
)

// https://github.com/kpbird/golang_random_string
func randStringBySoure(src string, n int) string {
	randomness := make([]byte, n)

	rand.Seed(time.Now().UnixNano())
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}

	l := len(src)

	// fill output
	output := make([]byte, n)
	for pos := range output {
		random := randomness[pos]
		randomPos := random % uint8(l)
		output[pos] = src[randomPos]
	}

	return string(output)
}
func RandNumString(n int) string {
	numbers := "0123456789"
	return randStringBySoure(numbers, n)
}

// XMLDecode XMLDecode
func XMLDecode(data []byte, v interface{}) error {
	decoder := xml.NewDecoder(bytes.NewReader([]byte(data)))
	decoder.CharsetReader = charset.NewReaderLabel
	return decoder.Decode(v)
}

// XMLEncode XMLEncode
func XMLEncode(v interface{}) ([]byte, error) {
	return xml.MarshalIndent(v, "", "  ")
}
