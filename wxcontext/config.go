package wxcontext

import "github.com/astaxie/beego/cache"

// Config for user
type Config struct {
	AppID          string
	AppSecret      string
	Token          string
	EncodingAESKey string
	Cache          cache.Cache

	// 商户平台参数
	MchID           string
	MchAPIKey       string // 商户平台APIKEY
	SslCertFilePath string // 证书公钥文件的路径
	SslKeyFilePath  string // 证书私钥文件的路径
	SslCertContent  string // 公钥证书的内容
	SslKeyContent   string // 私钥证书的内容

	// 开放平台参数
	ComponentAppId     string // 第三方平台组件APPID
	ComponentAppSecret string // 第三方平台组件SECRET
	ComponentAppToken  string // 第三方平台组件TOKEN
	ComponentAppKey    string // 第三方平台组件AESKEY
}
