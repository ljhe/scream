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
