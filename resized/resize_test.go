package resized

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestResize(t *testing.T) {
	file, err := os.Open("./calendar.jpg")
	if err != nil {
		t.Fatal("sample resize data not found")
	}
	body, _ := ioutil.ReadAll(file)
	data, _ := Resize(210, 210, 70, body)
	if data == nil {
		t.Error("resize failed, data is nil")
	}
}

func TestResizeAspectRatio(t *testing.T) {
	file, err := os.Open("./calendar.jpg")
	if err != nil {
		t.Fatal("sample resize data not found")
	}
	body, _ := ioutil.ReadAll(file)
	data, _ := Resize(0, 210, 70, body)
	if data == nil {
		t.Error("resize failed, data is nil")
	}
}

func TestWebpResize(t *testing.T) {
	file, err := os.Open("./calendar.jpg")
	if err != nil {
		t.Fatal("sample resize data not found")
	}
	body, _ := ioutil.ReadAll(file)
	data, _ := Resize(0, 210, 70, body)
	if data == nil {
		t.Error("resize failed, data is nil")
	}
	_, err = EncodeWebp(body, 70)
	if err != nil {
		t.Error("webp encode failed, error ", err.Error())
	}
}
