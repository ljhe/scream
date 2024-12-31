package encryption

import (
	"testing"
)

var privateKey = "-----BEGIN PRIVATE KEY-----\nMIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAKKHZi53pPok4K/M\ne0T3StjxAwTVRw7Mt3sZyrypmelPMQv4l1MtGbv8IyJu+xngN+r/8N80sYQGp8dt\ngRS2VIbOH4DOQ+E/6ZkOxlp+yqilfgHDNkvZZ4ME7wdEWhLMJ3CntbgmLPTWwPo9\nOq1I/ykWYbdIiJxnvJFJ3xi+hDAdAgMBAAECgYBmXsA2nneUcyvuru4px1Umhc+v\n+KI0KS+cYud2o6Dk+lNbRe4SnsKpzCtZIewZJUgzRZgzDt6M2SBOlaOLJzjfWuAk\nlZydwmtDbZYySTR3BDVOEFGkoKFZkIFJoy97/ZJoTQMRX0hYKeVCLF4jFH5I/IGZ\nJuDm0ZvJmMnf7EJVYQJBANG6/U4P0gXJFjYkJUqGgRHQI3WtaQogXru/E+76SYAI\nMMdrLlSFpwbWcvQOanEczSKOkPHFb6LWT2wELTF/YDUCQQDGYpgw8cLQ0Na80Kv4\nbwtPwZNgUkyHQ45SzpuVK5kQ/hoqzWRQqoaskFYyC3nkUpNz56earUjlmnvPzi1r\nFN1JAkEAh4ZKWuAUOhLX3IJ86myCCO2zjD5TSuzh6nYtvlZTmn0wcByNYqa+6Mc4\nnwaVt6QB1pvDg8euPM45ojYMshh6JQJAHNz1ZZGXFYh85aW6j3+gdq8kQQxYRAnJ\nKDUVH8PjFjzSE84kPTRCOdMaJ1fSGS0GdQOMOA3kIDu0rcxCgWTcuQJBAMg+gVWd\nKOqTT86O1D61AdYDhE1vllHD5r2VNc33+0WWDMfjBS4Gkfrp5ENXhJeI8NLzEUr8\nrTOGPuHVJu7LPu8=\n-----END PRIVATE KEY-----\n"
var publicKey = "-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCih2Yud6T6JOCvzHtE90rY8QME\n1UcOzLd7Gcq8qZnpTzEL+JdTLRm7/CMibvsZ4Dfq//DfNLGEBqfHbYEUtlSGzh+A\nzkPhP+mZDsZafsqopX4BwzZL2WeDBO8HRFoSzCdwp7W4Jiz01sD6PTqtSP8pFmG3\nSIicZ7yRSd8YvoQwHQIDAQAB\n-----END PUBLIC KEY-----"

func TestRSAGenerateKey(t *testing.T) {
	err := RSAGenerateKey("t", true)
	if err != nil {
		t.Errorf("TestRSAGenerateKey err:%s", err)
	}
}

func TestRSAEncrypt(t *testing.T) {
	str := "hello world"
	encryptStr, err := RSAEncrypt([]byte(str), []byte(publicKey))
	if err != nil {
		t.Errorf("TestRSADecrypt err:%s", err)
	}

	t.Logf("str:[%v] RSAEncrypt res:[%v]", str, encryptStr)

	decryptStr, err := RSADecrypt(encryptStr, []byte(privateKey))
	if err != nil {
		t.Errorf("TestRSADecrypt err:%s", err)
	}
	t.Logf("res:[%v] RSADecrypt str:[%v]", encryptStr, string(decryptStr))
}
