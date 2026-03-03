package middleware

import (
	BadgerDB "LiteNAS/badger"
	"LiteNAS/utils"
	"encoding/json"
	"errors"
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
// secretKey 用于后续如果切换到 JWT 时的签名验证
func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Token
		// 优先从 Header 获取 (Authorization: Bearer <token>)，其次从 Query 获取
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
		}

		// 处理 Bearer 前缀
		token = strings.TrimPrefix(token, "Bearer ")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 1, "message": "未登录，请先登录"})
			c.Abort()
			return
		}

		// 2. 从 BadgerDB 校验 Token 有效性
		// 假设登录时我们将 token -> user_info(json) 存入了 Badger
		val, err := BadgerDB.Get([]byte("session:" + token))
		if err != nil || val == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 1, "message": "登录已过期或无效"})
			c.Abort()
			return
		}

		// 3. 解析用户信息 (简单的示范：假设存的是 UserID 的字符串)
		// 实际上你可以在这里解析出完整的 UserInfo 结构体
		userIDStr := string(val)

		// 4. 将用户信息注入 Context，供后续 FilePolicy 中间件使用
		// 注意：这里存的是 uint，确保和数据库 ID 类型一致
		userID := utils.StringToUint(userIDStr)

		// 存入上下文供后续 FilePolicy 使用
		c.Set("user_id", userID)

		// 5. 也可以直接查一次数据库或缓存，把 IsAdmin 状态带上
		// 减少 FilePolicy 里重复查询数据库的次数

		c.Next()
	}
}
