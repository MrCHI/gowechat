package component

import "encoding/xml"

//EncMessage 微信加密消息固定格式
type EncMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"-"` // 开发者微信号
	Encrypt      string   // 加密的消息报文
	MsgSignature string   // 报文签名
	TimeStamp    string   // 时间戳
	Nonce        string   // 随机数
}

type AuthNotifyResponse struct {
	AppID                 string `xml:"AppId"`                 // APPID
	ComponentVerifyTicket string `xml:"ComponentVerifyTicket"` // 明文
}

type ErrorResponse struct {
	ErrCode int    `json:"errcode"` // 错误码
	ErrMsg  string `json:"errmsg"`  // 错误信息
}

// api_component_token
type ApiComponentTokenRequest struct {
	ComponentAppId        string `json:"component_appid"`         // 第三方平台appid
	ComponentAppSecret    string `json:"component_appsecret"`     // 第三方平台appsecret
	ComponentVerifyTicket string `json:"component_verify_ticket"` // 微信后台推送的ticket，此ticket会定时推送，具体请见本页末尾的推送说明
}

type ApiComponentTokenResponse struct {
	ErrorResponse
	ComponentAccessToken string `json:"component_access_token"` // 第三方平台access_token
	ExpiresIn            int64  `json:"expires_in"`             // 有效期
}

// api_create_preauthcode
type ApiCreatePreauthCodeRequest struct {
	ComponentAppId string `json:"component_appid"` // 第三方平台appid

}

type ApiCreatePreauthCodeResponse struct {
	ErrorResponse
	PreAuthCode string `json:"pre_auth_code"` // 预授权码
	ExpiresIn   int64  `json:"expires_in"`    // 有效期，为10分钟
}

type ApiQueryAuthRequest struct {
	ComponentAppId    string `json:"component_appid"`    // 第三方平台appid
	AuthorizationCode string `json:"authorization_code"` // 授权code,会在授权成功时返回给第三方平台，详见第三方平台授权流程说明
}

type ApiQueryAuthResponse struct {
	ErrorResponse
	AuthorizationInfo ApiQueryAuthAuthorizationInfo `json:"authorization_info"`
}

type ApiQueryAuthAuthorizationInfo struct {
	AuthorizerAppId        string                 `json:"authorizer_appid"`
	AuthorizerAccessToken  string                 `json:"authorizer_access_token"`
	ExpiresIn              int64                  `json:"expires_in"`
	AuthorizerRefreshToken string                 `json:"authorizer_refresh_token"`
	FuncInfo               []ApiQueryAuthFuncInfo `json:"func_info"`
}

type ApiQueryAuthFuncInfo struct {
	FuncscopeCategory ApiQueryAuthFuncscopeCategory `json:"funcscope_category"`
}

type ApiQueryAuthFuncscopeCategory struct {
	Id int `json:"id"`
}

// api_authorizer_token
type ApiAuthorizerTokenRequest struct {
	ErrorResponse
	ComponentAppId         string `json:"component_appid"`          // 第三方平台appid
	AuthorizerAppId        string `json:"authorizer_appid"`         // 授权方appid
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"` // 授权方的刷新令牌，刷新令牌主要用于第三方平台获取和刷新已授权用户的access_token，只会在授权时刻提供，请妥善保存。一旦丢失，只能让用户重新授权，才能再次拿到新的刷新令牌
}

type ApiAuthorizerTokenResponse struct {
	ErrorResponse
	AuthorizerAccessToken  string `json:"authorizer_access_token"`  // 授权方令牌
	ExpiresIn              int64  `json:"expires_in"`               // 有效期，为2小时
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"` // 刷新令牌
}
