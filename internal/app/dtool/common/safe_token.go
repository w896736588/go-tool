package common

import (
	"crypto/rand"
	"dev_tool/internal/pkg/p_common"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// SafeTokenClaims Safe Token 负载结构
type SafeTokenClaims struct {
	SessionID       string `json:"session_id"`
	PasswordVersion string `json:"password_version"`
	ExpireAt        int64  `json:"expire_at"`
}

// SafeTokenData 完整的 Token 数据结构（用于加密）
type SafeTokenData struct {
	Claims SafeTokenClaims `json:"claims"`
}

// SafeTokenManager Safe Token 管理器
type SafeTokenManager struct {
	password             string
	sessionExpireMinutes int
	passwordVersion      string
}

// NewSafeTokenManager 创建 Token 管理器
func NewSafeTokenManager(password string, expireMinutes int, appName string) *SafeTokenManager {
	expireMinutes = NormalizeSafeSessionExpireMinutes(expireMinutes)
	passwordVersion := BuildSafePasswordVersion(password, appName)

	return &SafeTokenManager{
		password:             password,
		sessionExpireMinutes: expireMinutes,
		passwordVersion:      passwordVersion,
	}
}

// RefreshConfig 刷新配置（配置变更后调用）
func (m *SafeTokenManager) RefreshConfig(password string, expireMinutes int, appName string) {
	m.password = password
	m.sessionExpireMinutes = NormalizeSafeSessionExpireMinutes(expireMinutes)
	m.passwordVersion = BuildSafePasswordVersion(password, appName)
}

// GenerateToken 生成新的 Safe Token
func (m *SafeTokenManager) GenerateToken() (string, int64, error) {
	// 生成随机会话ID
	sessionID := generateRandomSessionID()

	// 计算过期时间（0 表示永不过期）
	var expireAt int64
	if m.sessionExpireMinutes > 0 {
		expireAt = time.Now().Add(time.Duration(m.sessionExpireMinutes) * time.Minute).Unix()
	} else {
		expireAt = 0 // 永不过期
	}

	// 构建 Token 数据
	tokenData := SafeTokenData{
		Claims: SafeTokenClaims{
			SessionID:       sessionID,
			PasswordVersion: m.passwordVersion,
			ExpireAt:        expireAt,
		},
	}

	// 序列化为 JSON
	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		return "", 0, fmt.Errorf("token json marshal failed: %w", err)
	}

	// 使用 AES-GCM 加密
	encrypted, err := p_common.AesGcmClient.Encrypt(jsonData)
	if err != nil {
		return "", 0, fmt.Errorf("token encrypt failed: %w", err)
	}

	return encrypted, expireAt, nil
}

// ParseToken 解析并验证 Safe Token
// 返回值: (claims, isValid, errorCode, error)
// errorCode: 0=有效, 40101=token缺失, 40102=过期, 40103=密码版本不匹配, 40104=token非法
func (m *SafeTokenManager) ParseToken(token string) (*SafeTokenClaims, int, error) {
	if token == "" {
		return nil, 40101, fmt.Errorf("token is empty")
	}

	// 解密 token
	decrypted, err := p_common.AesGcmClient.Decrypt(token)
	if err != nil {
		return nil, 40104, fmt.Errorf("token decrypt failed: %w", err)
	}

	// 解析 JSON
	var tokenData SafeTokenData
	if err := json.Unmarshal([]byte(decrypted), &tokenData); err != nil {
		return nil, 40104, fmt.Errorf("token json unmarshal failed: %w", err)
	}

	claims := &tokenData.Claims

	// 检查过期时间（expireAt == 0 表示永不过期）
	if claims.ExpireAt > 0 && time.Now().Unix() > claims.ExpireAt {
		return nil, 40102, fmt.Errorf("token expired")
	}

	// 检查密码版本（密码修改后旧 token 失效）
	if claims.PasswordVersion != m.passwordVersion {
		return nil, 40103, fmt.Errorf("password version mismatch")
	}

	return claims, 0, nil
}

// RenewToken 续期 Token
// 每次请求成功时调用，返回新的 token 和过期时间
func (m *SafeTokenManager) RenewToken(claims *SafeTokenClaims) (string, int64, error) {
	// 生成新的过期时间（0 表示永不过期）
	var expireAt int64
	if m.sessionExpireMinutes > 0 {
		expireAt = time.Now().Add(time.Duration(m.sessionExpireMinutes) * time.Minute).Unix()
	} else {
		expireAt = 0 // 永不过期
	}

	// 构建新的 Token 数据（保留相同的 sessionID）
	tokenData := SafeTokenData{
		Claims: SafeTokenClaims{
			SessionID:       claims.SessionID,
			PasswordVersion: m.passwordVersion,
			ExpireAt:        expireAt,
		},
	}

	// 序列化为 JSON
	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		return "", 0, fmt.Errorf("token json marshal failed: %w", err)
	}

	// 使用 AES-GCM 加密
	encrypted, err := p_common.AesGcmClient.Encrypt(jsonData)
	if err != nil {
		return "", 0, fmt.Errorf("token encrypt failed: %w", err)
	}

	return encrypted, expireAt, nil
}

// VerifyPassword 验证登录密码
func (m *SafeTokenManager) VerifyPassword(inputPassword string) bool {
	if m.password == "" {
		// 未启用密码保护，任何密码都通过（但不建议这样使用）
		return true
	}
	return m.password == inputPassword
}

// IsEnabled 是否启用了后台密码保护
func (m *SafeTokenManager) IsEnabled() bool {
	return m.password != ""
}

// GetExpireMinutes 获取会话过期时间（分钟）
func (m *SafeTokenManager) GetExpireMinutes() int {
	return m.sessionExpireMinutes
}

// generateRandomSessionID 生成随机会话ID
func generateRandomSessionID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机生成失败，使用时间戳
		return fmt.Sprintf("session_%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}
