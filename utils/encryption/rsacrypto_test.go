package encryption

import (
	"testing"
)

func TestRSAGenerateKey(t *testing.T) {
	err := RSAGenerateKey("ws", true)
	if err != nil {
		t.Errorf("TestRSAGenerateKey err:%s", err)
	}
}

func TestRSAEncrypt(t *testing.T) {
	str := "hello world"
	encryptStr, err := RSAEncrypt([]byte(str), RSAWSPublicKey)
	if err != nil {
		t.Errorf("TestRSADecrypt err:%s", err)
	}

	t.Logf("str:[%v] RSAEncrypt res:[%v]", str, encryptStr)

	decryptStr, err := RSADecrypt(encryptStr, RSAWSPrivateKey)
	if err != nil {
		t.Errorf("TestRSADecrypt err:%s", err)
	}
	t.Logf("res:[%v] RSADecrypt str:[%v]", encryptStr, string(decryptStr))
}
