package goEncrypt_test

import (
	"github.com/open-source/game/chess.git/pkg/static/goEncrypt"
	"golang.org/x/sys/windows"
	"testing"
)

func TestAesCTR_Decrypt(t *testing.T) {
	data := "YWwwRzVUTDQ5ajVhL1k0ODY4N1Y3S0RSYS9XQnhQZGFRb0FDdHA5MGVKdz0="
	res, err := goEncrypt.AesCTR_Decrypt([]byte(data), []byte("facai20190110$#"))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(windows.ByteSliceToString(res))
}

func TestAesCTR_Encrypt(t *testing.T) {
	data := `{"uid":10001}`
	t.Log("原数据:", data)
	res, err := goEncrypt.AesCTR_Encrypt([]byte(data), []byte("facai20190110$#"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("加密后的数据：", string(res))
	res, err = goEncrypt.AesCTR_Decrypt(res, []byte("facai20190110$#"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("解密后的数据", string(res))
}
