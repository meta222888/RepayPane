package textencoding

import "testing"

func TestDecodeUTF8(t *testing.T) {
	text, info, err := Decode([]byte("hello 世界"))
	if err != nil {
		t.Fatal(err)
	}
	if text != "hello 世界" || info.Encoding != UTF8 {
		t.Fatalf("got %q %v", text, info)
	}
}

func TestDecodeUTF8BOM(t *testing.T) {
	data := append([]byte{0xEF, 0xBB, 0xBF}, []byte("abc")...)
	text, info, err := Decode(data)
	if err != nil || text != "abc" || !info.UTF8BOM {
		t.Fatalf("got %q %v err=%v", text, info, err)
	}
}

func TestDecodeGB18030(t *testing.T) {
	// "你好" in GBK/GB18030
	data := []byte{0xC4, 0xE3, 0xBA, 0xC3}
	text, info, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if text != "你好" || info.Encoding != GB18030 {
		t.Fatalf("got %q %v", text, info)
	}
}

func TestEncodeRoundTripGB18030(t *testing.T) {
	orig := "必看内容：中文测试"
	info := Info{Encoding: GB18030}
	enc, err := Encode(orig, info)
	if err != nil {
		t.Fatal(err)
	}
	text, outInfo, err := Decode(enc)
	if err != nil || text != orig || outInfo.Encoding != GB18030 {
		t.Fatalf("got %q %v", text, outInfo)
	}
}
