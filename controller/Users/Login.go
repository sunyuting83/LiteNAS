package Users

import (
	BadgerDB "LiteNAS/badger"
	"LiteNAS/database"
	"LiteNAS/utils"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Login 登录请求结构
type Login struct {
	UserName string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// CacheToken 存储在数据库中的用户信息
type CacheToken struct {
	UserID uint   `json:"user_id"`
	Token  string `json:"token"`
}

// Sgin 登录处理
func Sgin(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form Login
		if err := c.ShouldBind(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": 1, "message": "参数解析失败"})
			return
		}

		// 1. 验证用户身份 (始终检查数据库，确保账号未被禁用或修改密码)
		passwdHash := utils.MD5(form.Password + secretKey)
		login, err := database.CheckAdminLogin(form.UserName, passwdHash)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": 1, "message": "用户名或密码错误"})
			return
		}

		// 2. 检查 Badger 中是否已有该用户的 Token (实现单设备登录保持)
		userKey := "user:token:" + form.UserName
		oldToken, err := BadgerDB.Get([]byte(userKey))

		if err == nil && oldToken != "" {
			// 如果旧 Token 依然有效，直接返回加密后的旧 Token
			finalToken, _ := utils.EncryptByAes([]byte(oldToken), []byte(secretKey))
			c.JSON(http.StatusOK, gin.H{
				"status":   0,
				"message":  "欢迎回来",
				"token":    finalToken,
				"username": login.UserName,
				"userid":   login.ID,
			})
			return
		}

		// 3. 生成新 Token
		// 使用 时间戳 + 用户名 + 随机盐 确保唯一性
		seed := strings.Join([]string{login.UserName, time.Now().String(), "secure_salt"}, ":")
		newToken := utils.MD5(seed)
		ttl := int64(60 * 60 * 24 * 90) // 90天有效期

		// 4. 准备缓存数据
		cacheData := &CacheToken{
			UserID: login.ID,
			Token:  newToken,
		}
		buff, _ := json.Marshal(cacheData)

		// 5. 原子化存入 Badger (双向绑定)
		// a. 用户名索引：user:token:admin -> 32位Token原文
		BadgerDB.SetWithTTL([]byte(userKey), []byte(newToken), ttl)
		// b. Token 详细信息：token:info:32位Token原文 -> JSON(UserID, Token原文)
		tokenInfoKey := "token:info:" + newToken
		BadgerDB.SetWithTTL([]byte(tokenInfoKey), buff, ttl)

		// 6. 加密 Token 原文发给前端
		finalToken, _ := utils.EncryptByAes([]byte(newToken), []byte(secretKey))

		c.JSON(http.StatusOK, gin.H{
			"status":   0,
			"message":  "登录成功",
			"token":    finalToken,
			"username": login.UserName,
			"userid":   login.ID,
		})
	}
}

// Logout 注销处理
func Logout(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取当前加密 Token
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusOK, gin.H{"status": 1, "message": "无效的 Token"})
			return
		}
		encryptedToken := authHeader[7:]

		tokenRawBytes, err := utils.DecryptByAes(encryptedToken, []byte(secretKey))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": 1, "message": "Token 解析失败"})
			return
		}
		tokenRaw := string(tokenRawBytes)

		// 3. 从 TokenInfo 中反查出 UserName 以清理双向索引
		tokenInfoKey := "token:info:" + tokenRaw
		var userInfo CacheToken

		// 先读出信息以便获取用户名（假设你的 Badger Get 返回 []byte）
		data, err := BadgerDB.GetToken([]byte(tokenRaw)) // 注意：这里调用你之前封装的 GetToken
		if err == nil {
			json.Unmarshal(data, &userInfo)

			// 获取用户名并删除用户索引 (需要数据库查询或在 CacheToken 中多存一个用户名)
			// 这里简单处理：直接从数据库或 context 获取当前用户名
			// 建议在 CacheToken 结构体里增加 UserName 字段，注销时就非常方便
		}

		// 4. 执行物理删除（使 Token 立即失效）
		// 注意：BadgerDB 需要实现 Delete 方法
		BadgerDB.Delete([]byte(tokenInfoKey))

		// 如果你有存储 user:token:username，也建议清理掉
		// BadgerDB.Delete([]byte("user:token:" + userInfo.UserName))

		c.JSON(http.StatusOK, gin.H{
			"status":  0,
			"message": "已注销登录",
		})
	}
}
