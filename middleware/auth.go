package middleware

import (
	BadgerDB "LiteNAS/badger" // 请确保路径正确

	// 假设 Message 结构体在此
	"LiteNAS/utils"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

// CacheToken 局部定义，确保 JSON 解析安全
type CacheToken struct {
	UserID uint   `json:"user_id"`
	Token  string `json:"token"`
}

// 这里的 key 常量用于 Context 传递
const AuthUserKey = "current_admin_user"

// AdminVerifyMiddleware 验证中间件
func AdminVerifyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// 1. 健壮检查 Header 格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatus(403)
			return
		}
		tokenStr := authHeader[7:]

		// 2. 获取加密密钥 (增加保护)
		secretRaw, exists := c.Get("secret_key")
		if !exists {
			c.AbortWithStatus(500)
			return
		}
		secretKey := secretRaw.(string)

		// 3. 校验并解析 Token (核心逻辑)
		userInfo, err := validateAndGetInfo(secretKey, tokenStr)
		if err != nil {
			c.AbortWithStatus(403)
			return
		}

		// 4. 关键优化：存入 Context 供后续使用，避免重复查库
		c.Set(AuthUserKey, userInfo)
		c.Next()
	}
}

// GetCurrentAdminID 从 Context 直接获取 ID，不再重复解密
func GetCurrentAdminID(c *gin.Context) uint {
	if val, exists := c.Get(AuthUserKey); exists {
		if info, ok := val.(*CacheToken); ok {
			return info.UserID
		}
	}
	return 0
}

// validateAndGetInfo 校验逻辑封装 (线程安全)
func validateAndGetInfo(s, a string) (*CacheToken, error) {
	// AES 解密
	aesToken, err := utils.DecryptByAes(a, []byte(s))
	if err != nil {
		return nil, err
	}

	// 从 BadgerDB 获取数据
	dbData, err := BadgerDB.GetToken(aesToken)
	if err != nil {
		return nil, err
	}

	// 解析到局部变量 (并发安全)
	var info CacheToken
	if err := json.Unmarshal(dbData, &info); err != nil {
		return nil, err
	}

	// 校验 Token 字符串是否匹配
	if info.Token != string(aesToken) {
		return nil, errors.New("token mismatch")
	}

	return &info, nil
}
