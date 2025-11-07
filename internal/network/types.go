package network

import (
	"crypto/tls"
	"net/http"
	"time"
)



// 存储Body规则中匹配版本号的信息
type BodyMatch struct {
	Count int
	Name  string // 匹配到的CMS名称（带版本号）
}

// 存储按匹配位置和方法分类的规则
type LocationRules struct {
	// 关键字匹配规则
	BodyKeywordRules   []FingerprintRule
	TitleKeywordRules  []FingerprintRule
	HeaderKeywordRules []FingerprintRule

	// 正则表达式匹配规则
	BodyRegexRules   []FingerprintRule
	TitleRegexRules  []FingerprintRule
	HeaderRegexRules []FingerprintRule

	// Hash匹配规则
	FaviconMmh3Hash []FingerprintRule
	FaviconMd5Hash  []FingerprintRule
}

// finger.json 文件中的每一条具体的指纹匹配规则
type FingerprintRule struct {
	CMS      string   `json:"cms"`
	Method   string   `json:"method"`
	Location string   `json:"location"`
	Keyword  []string `json:"keyword"`
}

// 存储TLS握手和证书的主要信息
type TLSInfo struct {
	Version     string    // 协商后的TLS版本, e.g., "TLS 1.3"
	CipherSuite string    // 协商后的加密套件
	Certificate *CertInfo // 详细的证书信息
}

// 包含了X.509证书的16个关键信息字段
type CertInfo struct {
	FingerprintSHA1       string    `json:"fingerprint_sha1"`        // 证书的SHA1指纹
	FingerprintSHA256     string    `json:"fingerprint_sha256"`      // 证书的SHA256指纹
	Subject               string    `json:"subject"`                 // 主题
	Issuer                string    `json:"issuer"`                  // 颁发者
	NotBefore             time.Time `json:"not_before"`              // 生效日期
	NotAfter              time.Time `json:"not_after"`               // 失效日期
	SerialNumber          string    `json:"serial_number"`           // 序列号
	PublicKeyAlgorithm    string    `json:"public_key_algorithm"`    // 公钥算法
	PublicKey             string    `json:"public_key"`              // 公钥 (Hex)
	SignatureAlgorithm    string    `json:"signature_algorithm"`     // 签名算法
	Version               int       `json:"version"`                 // 证书版本
	OCSPStapling          bool      `json:"ocsp_stapling"`           // 是否包含OCSP Stapling信息
	OCSPReponder          []string  `json:"ocsp_reponder"`           // OCSP响应服务器
	DNSNames              []string  `json:"dns_names"`               // DNS名称 (SAN)
	EmailAddresses        []string  `json:"email_addresses"`         // 电子邮件地址 (SAN)
	IPAddresses           []string  `json:"ip_addresses"`            // IP地址 (SAN)
	URIs                  []string  `json:"uris"`                    // URIs (SAN)
	CRLDistributionPoints []string  `json:"crl_distribution_points"` // CRL分发点
	IssuingCertificateURL []string  `json:"issuing_certificate_url"` // 颁发证书URL
}

// 提供广泛兼容性的密码套件列表
var DefaultCipherSuites = []uint16{
	tls.TLS_RSA_WITH_RC4_128_SHA,
	tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_AES_128_GCM_SHA256,
	tls.TLS_AES_256_GCM_SHA384,
	tls.TLS_CHACHA20_POLY1305_SHA256,
	tls.TLS_FALLBACK_SCSV,
}

// 包含一次HTTP请求的所有相关信息
type ResponseInfo struct {
	URL             string         // 请求的最终URL（处理重定向后）
	Path            string         // 最终URL的路径部分
	StatusCode      int            // HTTP响应状态码
	Headers         http.Header    // 完整的HTTP响应头
	Body            []byte         // 响应正文内容
	Cookies         []*http.Cookie // 响应中设置的Cookies
	RedirectHistory []string       // 记录HTTP重定向的URL链条
	Title           string         // HTML页面标题 (后续解析Body填充)
	JsFiles         []string       // 从HTML中提取的JS文件名 (后续解析Body填充)
	CssFiles        []string       // 从HTML中提取的CSS文件名 (后续解析Body填充)
	Favicon         string         // Favicon图标URL (后续步骤获取)
	FaviconBytes    []byte         // Favicon图标的原始数据
	FaviconMmh3     string         // Favicon图标的mmh3哈希值
	FaviconMd5      string         // Favicon图标的MD5哈希值
	TLS             *TLSInfo       `json:"tls,omitempty"`
}
