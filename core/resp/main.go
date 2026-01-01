package resp

// RESP (REdis Serialization Protocol) is a fast, simple, text-based protocol used for 
// client-server communication. It is a request-response protocol i.e. clients sends request
// in RESP and server responds back in RESP. Every data starts with a special character and
// the data ends with \r\n (CRLF) as a terminator for easy parsing and human readability.
// Carriage Return + Line Feed)

import (
	"errors"
	"fmt"
	"log"
)

// reads the length typically the first integer of the string
// until hit by an non-digit byte and returns
// the integer and the delta = length + 2 (CRLF)
// TODO: Make it simpler and read until we get `\r` just like other functions
func readLength(data []byte) (int, int) {
	pos, length := 0, 0
	for pos = range data {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}
		length = length*10 + int(b-'0')
	}
	return 0, 0
}

// readSimpleString decodes a RESP simple string from the given byte slice.
// RESP simple strings are encoded as: +<string>\r\n
//
// Parameters:
//   - data: byte slice containing the RESP-encoded simple string
//
// Returns:
//   - string: the decoded string content (without +, \r, \n)
//   - int: number of bytes consumed from data (delta/offset)
//   - error: any error encountered during parsing (currently always nil)
//
// Example:
//   Input:  []byte("+OK\r\n")
//   Output: "OK", 5, nil
func readSimpleString(data []byte) (string, int, error) {
	// Skip the '+' prefix
	pos := 1 

	// Find the '\r' that marks the end of the string
	for ; data[pos] != '\r'; pos++ {
	}

	// Extract string between '+' and '\r', skip '\r\n' (2 bytes)
	return string(data[1:pos]), pos + 2, nil
}

// reads a RESP encoded error from data and returns
// the error string, the delta, and the error
func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

// reads a RESP encoded integer from data and returns
// the intger value, the delta, and the error
func readInt64(data []byte) (int64, int, error) {
	// first character ':'
	pos := 1
	var value int64 = 0

	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}

	return value, pos + 2, nil
}

// reads a RESP encoded string from data and returns
// the string, the delta, and the error
func readBulkString(data []byte) (string, int, error) {
	// first character '$'
	pos := 1
	// reading the length and forwarding the pos by
	// the lenth of the integer + the first special character
	len, delta := readLength(data[pos:])
	pos += delta

	// reading `len` bytes as string
	return string(data[pos:(pos + len)]), pos + len + 2, nil
}

// reads a RESP encoded array from data and returns
// the array, the delta, and the error
func readArray(data []byte) (interface{}, int, error) {
	// first character '*'
	pos := 1
	// reading the length
	count, delta := readLength(data[pos:])
	pos += delta
	log.Println("read length done")
	log.Println("count= ",count, "pos = ", pos)
	log.Println(data[pos:])
	log.Println(data[pos])
	log.Println(data)
	var elems []interface{} = make([]interface{}, count)
	for i := range elems {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elems[i] = elem
		pos += delta
	}
	return elems, pos, nil
}

func DecodeOne(data []byte) (interface{}, int, error) {

	if len(data) == 0 {
		log.Println("NO DATA")
		return nil, 0, errors.New("No data.")
	}
	switch data[0] {
		case '+':
			log.Println("READ SIMPLE STRING")
			return readSimpleString(data)
		case '-':
			return readError(data)
		case ':':
			return readInt64(data)
		case '$':
			log.Println("READ BUCK STRING")
			return readBulkString(data)
		case '*':
			log.Println("READ ARRAY")
			return readArray(data)
		}
	return nil, 0, nil
}

func Decode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("No data")
	}
	value, _, err := DecodeOne(data)
	return value, err
}

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	}
	return []byte{}
}

func DecodeArrayString(data []byte) ([]string, error) {
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}

	ts := value.([]interface{}) // temporary slice 
	// Create an slice(array) of strings, with a fixed length, but empty values.
	// Does not copy elements yet.
	tokens := make([]string, len(ts))

	// convert each element
	for i := range tokens {
		// type assertion
		tokens[i] = ts[i].(string)
	}

	return tokens, nil
}