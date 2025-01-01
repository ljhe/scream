package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

var RSAWSPrivateKey = []byte("-----BEGIN PRIVATE KEY-----\nMIICeAIBADANBgkqhkiG9w0BAQEFAASCAmIwggJeAgEAAoGBAMPX5ZSo+twqev+x\nHmCgNoS6AQT7LDFyAVVIzw9tseUxKXAU23n+HCdF3HMYCT74i+VvTcGcx8kLJfgc\n2eUxfhoYRQfK7gI6K8LpKdnTGH8ehMWrxDVPJmKUyWFCxPyrjKaFNx+NmeVK4uAQ\nOhgsXXDUvx3kjdmaDxgMghoDokxfAgMBAAECgYEAnaWmHgPkY+QiPL9D175ABJmC\nBpN1oJvH7PH+E8pWgEsRszJm9g2Sdh6rdU5s6u7CFj+BlQ/yVqiNuOroj7FGcqxP\nHavMyXWNBLWD9au8VXebTonT9K5OLupmUR0WCRBk+pjGGvGSbLZzERoIdIfArp7S\n7fAx1ifkTrvYroa0RwkCQQDYKedfIHiRrlCDP+DPuFkwuiqi2t6tQL+LHcBrzQHZ\njyFJZvwo/eUB4+ADcrZW6+wEb+FhzlRoyBa2FOsSeNP9AkEA5+9TtXfiA1amHp2y\nxKQQf/9D+meTrVcUklugH9q0A3tjRDldJUaqmSIipgvQ4U/ZeZSwL9jYKzRZRWz+\nB9WaiwJBAK7g18ph3qkdOQ219A6Yua9uLWgrYdMQeuX1X+LWrBRycx+LLZ2MKmVp\nEaY4e8O+geblDJWv8yICHj2YlsUO85ECQGav94foxBBmVLZJa9TULtn80sQTB7c/\nTsRd/M8drYW9I34ZR7wxRWb3Tg/mO10GVWsXAcqtX0gBrWSnlPEzCXECQQChnHZo\nRDBt7WlwRuI+1vcbS5k/piqUj8C0AC6Uc+R7fMvkYw197FNC6gI+1Mwr65KPvg0W\nOlIMwBC3dbT130GP\n-----END PRIVATE KEY-----\n")
var RSAWSPublicKey = []byte("-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDD1+WUqPrcKnr/sR5goDaEugEE\n+ywxcgFVSM8PbbHlMSlwFNt5/hwnRdxzGAk++Ivlb03BnMfJCyX4HNnlMX4aGEUH\nyu4COivC6SnZ0xh/HoTFq8Q1TyZilMlhQsT8q4ymhTcfjZnlSuLgEDoYLF1w1L8d\n5I3Zmg8YDIIaA6JMXwIDAQAB\n-----END PUBLIC KEY-----\n")

// RSAGenerateKey 生成私钥和公钥
func RSAGenerateKey(suffix string, save bool) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return fmt.Errorf("error generating RSA key: %s", err)
	}

	PKCS8Private, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("error marshalling RSA private key: %s", err)
	}

	pemBlockPrivate := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: PKCS8Private,
	}
	pemDataPrivate := pem.EncodeToMemory(pemBlockPrivate)
	fmt.Println("PKCS#8 格式的私钥: \n", string(pemDataPrivate))

	PKCS8Public, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("error marshalling RSA private key: %s", err)
	}

	pemBlockPublic := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: PKCS8Public,
	}
	pemDataPublic := pem.EncodeToMemory(pemBlockPublic)
	fmt.Println("PKCS#8 格式的公钥: \n", string(pemDataPublic))

	if save {
		privateFile, err := os.Create("./private_" + suffix + ".pem")
		if err != nil {
			return err
		}
		defer privateFile.Close()
		if err = pem.Encode(privateFile, pemBlockPrivate); err != nil {
			return err
		}

		publicFile, err := os.Create("./public_" + suffix + ".pem")
		if err != nil {
			return err
		}
		defer publicFile.Close()

		if err = pem.Encode(publicFile, pemBlockPublic); err != nil {
			return err
		}
	}

	return nil
}

// RSAEncrypt 加密:先RSA加密 然后base64转码
func RSAEncrypt(str, key []byte) ([]byte, error) {
	block, _ := pem.Decode(key)
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	tmpPublicKey := publicKeyInterface.(*rsa.PublicKey)
	tmpRetText, err := rsa.EncryptPKCS1v15(rand.Reader, tmpPublicKey, str)
	if err != nil {
		return nil, err
	}

	retText := base64.StdEncoding.EncodeToString(tmpRetText)
	return []byte(retText), nil
}

// RSADecrypt 解密:先base64转码 再RSA解密
func RSADecrypt(str, key []byte) ([]byte, error) {
	tmpStr, err := base64.StdEncoding.DecodeString(string(str))
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(key)
	tmpPrivateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	retText, err := rsa.DecryptPKCS1v15(rand.Reader, tmpPrivateKey.(*rsa.PrivateKey), tmpStr)
	if err != nil {
		return nil, err
	}
	return retText, err
}

// RSAReadKey 从文件中读取密钥
func RSAReadKey(path string) ([]byte, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return key, err
}
