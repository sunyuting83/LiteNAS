package middleware

import (
	BadgerDB "LiteNAS/badger"
	"LiteNAS/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
		fmt.Println(tokenStr)

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

// AuthMiddleware 身份验证中间件
func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Token
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
		}
		token = strings.TrimPrefix(token, "Bearer ")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 1, "message": "未登录，请先登录"})
			c.Abort()
			return
		}

		// 2. 解密 Token (因为 Login 发给前端的是 AES 加密后的结果)
		// 必须先解密成 MD5 原文，才能去 Badger 匹配 Key
		decryptedToken, err := utils.DecryptByAes(token, []byte(secretKey))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 1, "message": "无效的 Token 格式"})
			c.Abort()
			return
		}
		tokenRaw := string(decryptedToken)

		// 3. 从 BadgerDB 校验 Token 有效性
		// 注意：这里的 Key 必须和 Login 里的 "token:info:" + newToken 保持一致
		tokenInfoKey := "token:info:" + tokenRaw
		val, err := BadgerDB.Get([]byte(tokenInfoKey))
		if err != nil || val == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 1, "message": "登录已过期或无效"})
			c.Abort()
			return
		}

		// 4. 解析缓存的 JSON 数据 (Login 里存的是 CacheToken 结构体)
		var cacheData struct {
			UserID uint   `json:"user_id"`
			Token  string `json:"token"`
		}
		if err := json.Unmarshal([]byte(val), &cacheData); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 1, "message": "会话数据解析失败"})
			c.Abort()
			return
		}

		// 5. 将用户信息注入 Context
		// 存入 user_id 供后续业务逻辑使用
		c.Set("user_id", cacheData.UserID)

		c.Next()
	}
}
