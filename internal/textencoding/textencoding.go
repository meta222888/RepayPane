package textencoding

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Encoding string

const (
	UTF8     Encoding = "UTF-8"
	GB18030  Encoding = "GB18030"
	UTF16LE  Encoding = "UTF-16 LE"
)

type Info struct {
	Encoding Encoding
	UTF8BOM  bool
}

func (i Info) Label() string {
	if i.Encoding == UTF8 && i.UTF8BOM {
		return "UTF-8 BOM"
	}
	return string(i.Encoding)
}

func Decode(data []byte) (string, Info, error) {
	if len(data) == 0 {
		return "", Info{Encoding: UTF8}, nil
	}
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		body := data[3:]
		if !utf8.Valid(body) {
			return "", Info{}, errors.New("invalid UTF-8 after BOM")
		}
		return string(body), Info{Encoding: UTF8, UTF8BOM: true}, nil
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		text, err := decodeUTF16LE(data[2:])
		return text, Info{Encoding: UTF16LE}, err
	}
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		text, err := decodeUTF16BE(data[2:])
		return text, Info{Encoding: UTF16LE}, err
	}
	if utf8.Valid(data) {
		return string(data), Info{Encoding: UTF8}, nil
	}
	text, err := decodeWith(simplifiedchinese.GB18030.NewDecoder(), data)
	if err != nil {
		return "", Info{}, err
	}
	return text, Info{Encoding: GB18030}, nil
}

func Encode(text string, info Info) ([]byte, error) {
	switch info.Encoding {
	case UTF16LE:
		return encodeUTF16LE(text), nil
	case GB18030:
		return encodeWith(simplifiedchinese.GB18030.NewEncoder(), text)
	case UTF8:
		fallthrough
	default:
		out := []byte(text)
		if info.UTF8BOM {
			out = append([]byte{0xEF, 0xBB, 0xBF}, out...)
		}
		return out, nil
	}
}

func decodeWith(tr transform.Transformer, data []byte) (string, error) {
	r := transform.NewReader(bytes.NewReader(data), tr)
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func encodeWith(tr transform.Transformer, text string) ([]byte, error) {
	r := transform.NewReader(bytes.NewReader([]byte(text)), tr)
	return io.ReadAll(r)
}

func decodeUTF16LE(data []byte) (string, error) {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}
	u16 := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		u16 = append(u16, uint16(data[i])|uint16(data[i+1])<<8)
	}
	return string(utf16.Decode(u16)), nil
}

func decodeUTF16BE(data []byte) (string, error) {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}
	u16 := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		u16 = append(u16, uint16(data[i])<<8|uint16(data[i+1]))
	}
	return string(utf16.Decode(u16)), nil
}

func encodeUTF16LE(text string) []byte {
	u16 := utf16.Encode([]rune(text))
	out := []byte{0xFF, 0xFE}
	for _, c := range u16 {
		out = append(out, byte(c), byte(c>>8))
	}
	return out
}
