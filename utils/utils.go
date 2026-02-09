package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port       string `yaml:"port"`
	SECRET_KEY string `yaml:"SECRET_KEY"`
	AdminPWD   string `yaml:"AdminPWD"`
	FormMemory int64  `yaml:"FormMemory"` // 单位: MB
}

// GetCurrentPath Get Current Path
func GetCurrentPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(path)
	return dir, nil
}

// GenerateRandomKey 生成指定长度的随机字符串 (安全)
func GenerateRandomKey(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)[:n]
}

func CheckConfig(currentPath string) (*Config, error) {
	configFile := filepath.Join(currentPath, "config.yaml")
	conf := &Config{}

	// 1. 如果文件不存在，设置默认值并创建
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		conf.Port = "13008"
		conf.SECRET_KEY = GenerateRandomKey(32)
		conf.AdminPWD = GenerateRandomKey(12)
		conf.FormMemory = 32

		data, _ := yaml.Marshal(conf)
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			return nil, fmt.Errorf("创建配置文件失败: %v", err)
		}
		// 第一次创建，建议提示用户记录初始密码
		fmt.Printf("首次启动，已生成初始配置。管理员密码: %s\n", conf.AdminPWD)
	} else {
		// 2. 文件存在则读取
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, conf); err != nil {
			return nil, err
		}
	}
	return conf, nil
}

func IsExist(path string) bool {
	// 判断文件是否存在
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func MD5(a string) string {
	data := []byte(a)
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// --- 核心加密逻辑 (已增强安全性) ---

// AesEncrypt 使用随机 IV 进行加密
func AesEncrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	data = pkcs7Padding(data, blockSize)

	// 增强：使用随机产生的 IV
	ciphertext := make([]byte, blockSize+len(data))
	iv := ciphertext[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(ciphertext[blockSize:], data)
	return ciphertext, nil
}

// AesDecrypt 适配随机 IV 的解密
func AesDecrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(data) < blockSize {
		return nil, errors.New("密文太短")
	}

	// 提取密文开头的 IV
	iv := data[:blockSize]
	data = data[blockSize:]

	blockMode := cipher.NewCBCDecrypter(block, iv)
	crypted := make([]byte, len(data))
	blockMode.CryptBlocks(crypted, data)

	return pkcs7UnPadding(crypted)
}

// EncryptByAes 对外接口
func EncryptByAes(data, PwdKey []byte) (string, error) {
	res, err := AesEncrypt(data, PwdKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(res), nil
}

// DecryptByAes 对外接口
func DecryptByAes(data string, PwdKey []byte) ([]byte, error) {
	dataByte, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return AesDecrypt(dataByte, PwdKey)
}

// PKCS7 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("数据为空")
	}
	unPadding := int(data[length-1])
	if length < unPadding {
		return nil, errors.New("解密填充错误")
	}
	return data[:(length - unPadding)], nil
}
