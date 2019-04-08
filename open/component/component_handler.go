package component

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"github.com/MrCHI/WXBizMsgCrypt"

	"github.com/MrCHI/gowechat/wxcontext"

	"models.umilive.com/module_common"

	"github.com/MrCHI/gowechat/open/base"
)

const (
	componentAccessTokenURL = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"
	getPreCodeURL           = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token=%s"
	queryAuthURL            = "https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token=%s"
	refreshTokenURL         = "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token=%s"
	getComponentInfoURL     = "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_info?component_access_token=%s"
	getComponentConfigURL   = "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_option?component_access_token=%s"
	bindComponentURL        = "https://mp.weixin.qq.com/safe/bindcomponent?action=%s"
)

type Component struct {
	base.OpenBase
	componentVerifyTicket string
	componentAccessToken  string
	authorizationCode     string
	cryptTools            WXBizMsgCrypt.WXBizMsgCrypt // 加密、解密
}

func NewComponent(context *wxcontext.Context) *Component {
	component := new(Component)
	component.Context = context
	return component
}

// 处理微信10分钟1次的推送消息
func (_this *Component) HandlerCallBack(bodyEncrypt string, nonce string, encryptType string, msgSign string, timestamp int64) (*AuthNotifyResponse, error) {
	fmt.Printf("收到微信开放平台推送消息，10分钟/次\n")

	if _this.cryptTools == nil {
		_this.cryptTools, _ = WXBizMsgCrypt.NewWXBizMsgCrypt(_this.ComponentAppToken, _this.ComponentAppKey, _this.ComponentAppId)
	}

	result, decryp_xml := _this.cryptTools.DecryptMsg(bodyEncrypt, msgSign, timestamp, nonce)

	if result != 0 {
		return nil, errors.New(fmt.Sprintf("decryp error:%v", result))
	}

	// 提取信息
	authNotify := &AuthNotifyResponse{}
	err := xml.Unmarshal([]byte(decryp_xml), authNotify)

	if err != nil {
		return nil, err
	}

	fmt.Printf("AppId:%v\n", authNotify.AppID)
	fmt.Printf("ComponentVerifyTicket:%v\n", authNotify.ComponentVerifyTicket)

	// 通过APPID过虑
	if _this.Context.ComponentAppId != authNotify.AppID {
		return nil, fmt.Errorf("%v", "app_id is invalid.")
	}

	// 更新Ticket内容
	_this.componentVerifyTicket = authNotify.ComponentVerifyTicket

	// 获取开放平台开发者凭据
	_this.GetComponentAccessToken(_this.componentVerifyTicket)

	return authNotify, nil
}

// 获取第三方平台开发者凭据
func (_this *Component) GetComponentAccessToken(componentVerifyTicket string) (access_token *ApiComponentTokenResponse, e error) {
	fmt.Printf("获取第三方平台开发者凭据\n")

	if componentVerifyTicket == "" {
		return nil, errors.New("component_verify_ticket is invalid.")
	}

	// 优先从缓存中读取
	component_access_token_key := fmt.Sprintf("component_access_token_%s", _this.Context.AppID)
	component_access_token := _this.Context.Cache.Get(component_access_token_key)

	if component_access_token != nil {
		return component_access_token.(*ApiComponentTokenResponse), nil
	}

	// 从微信服务器获取
	jsonData := ApiComponentTokenRequest{
		ComponentAppId:        _this.ComponentAppId,
		ComponentAppSecret:    _this.ComponentAppSecret,
		ComponentVerifyTicket: componentVerifyTicket,
	}

	jsonString, err := json.Marshal(jsonData)

	if err != nil {
		return nil, err
	}

	result, err := module_common.HttpPostJson(componentAccessTokenURL, string(jsonString))

	if err != nil {
		return nil, err
	}

	componentToken := &ApiComponentTokenResponse{}
	err = json.Unmarshal([]byte(result), componentToken)

	if err != nil {
		return nil, err
	}

	fmt.Printf("expires_in:%v\n", componentToken.ExpiresIn)
	fmt.Printf("component_access_token:%v\n", componentToken.ComponentAccessToken)

	// 写入缓存
	expires := componentToken.ExpiresIn - 1500
	err = _this.Context.Cache.Put(component_access_token_key, componentToken, time.Duration(expires)*time.Second)

	if err != nil {
		return nil, err
	}

	return componentToken, nil
}

// 获取第三方平台开发者预授权码
func (_this *Component) GetPreAuthCode() (*ApiCreatePreauthCodeResponse, error) {
	fmt.Printf("获取第三方平台开发者预授权码\n")

	componentAccessToken, err := _this.GetComponentAccessToken(_this.componentVerifyTicket)

	if err != nil {
		return nil, errors.New("component_access_token is invalid.")
	}

	jsonData := ApiCreatePreauthCodeRequest{
		ComponentAppId: _this.ComponentAppId,
	}

	jsonString, err := json.Marshal(jsonData)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	result, err := module_common.HttpPostJson(getPreCodeURL+componentAccessToken.ComponentAccessToken, string(jsonString))

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	preAuthCode := &ApiCreatePreauthCodeResponse{}
	err = json.Unmarshal([]byte(result), preAuthCode)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	fmt.Printf("expires_in:%v\n", preAuthCode.ExpiresIn)
	fmt.Printf("pre_auth_code:%v\n", preAuthCode.PreAuthCode)

	return preAuthCode, nil
}

// 获取授权页面
func (_this *Component) GetAuthWeb(preAuthCode string, bindComponentCallbackURL string) (string, error) {
	url := fmt.Sprintf(bindComponentURL, "bindcomponent"+
		"&auth_type=3"+
		"&no_scan=1"+
		"&component_appid="+_this.ComponentAppId+
		"&pre_auth_code="+preAuthCode+
		"&redirect_uri="+bindComponentCallbackURL+
		"&auth_type=3"+
		"#wechat_redirect")

	return url, nil
}

// 授权后回调URI，得到授权码（authorization_code）和过期时间
func (_this *Component) AuthWebCallback(auth_code string, expires_in int64) (authorizationCode string, authorizationExpiresIn int64) {
	_this.authorizationCode = auth_code
	authorizationExpiresIn = expires_in

	fmt.Printf("第三方同意授权\n")
	fmt.Printf("authorization_code:%v\n", auth_code)
	fmt.Printf("authorization_expires_in:%v\n", expires_in)

	_this.QueryAuthCode(_this.authorizationCode, _this.componentAccessToken)

	return _this.authorizationCode, authorizationExpiresIn
}

// 使用授权码换取公众号或小程序的接口调用凭据和授权信息
func (_this *Component) QueryAuthCode(authorizationCode string, componentAccessToken string) (*ApiQueryAuthResponse, error) {
	fmt.Printf("使用授权码换取第三方公众号或小程序的接口调用凭据和授权信息\n")

	if authorizationCode == "" {
		return nil, errors.New("authorization_code is invalid.")
	}

	if componentAccessToken == "" {
		return nil, errors.New("component_access_token is invalid.")
	}

	// 优先从缓存中读取
	authorizer_access_tokenn_key := fmt.Sprintf("authorizer_access_tokenn_key_%s", _this.Context.AppID)
	authorizer_access_tokenn := _this.Context.Cache.Get(authorizer_access_tokenn_key)

	if authorizer_access_tokenn != nil {
		return authorizer_access_tokenn.(*ApiQueryAuthResponse), nil
	}

	jsonData := ApiQueryAuthRequest{
		ComponentAppId:    _this.ComponentAppId,
		AuthorizationCode: authorizationCode,
	}

	jsonString, err := json.Marshal(jsonData)

	if err != nil {
		return nil, err
	}

	result, err := module_common.HttpPostJson(queryAuthURL+componentAccessToken, string(jsonString))

	if err != nil {
		return nil, err
	}

	queryAuth := &ApiQueryAuthResponse{}
	err = json.Unmarshal([]byte(result), queryAuth)

	if err != nil {
		return nil, err
	}

	fmt.Printf("cacheData.QueryAuth:%+v\n", queryAuth)

	// 写入缓存
	expires := queryAuth.AuthorizationInfo.ExpiresIn - 1500
	err = _this.Context.Cache.Put(authorizer_access_tokenn_key, queryAuth, time.Duration(expires)*time.Second)

	if err != nil {
		return nil, err
	}

	return queryAuth, nil
}

// 获取（刷新）授权公众号或小程序的接口调用凭据（令牌）
func (_this *Component) RefreshAuthToken() (*ApiAuthorizerTokenResponse, error) {
	fmt.Printf("获取（刷新）第三方授权公众号或小程序的接口调用凭据（令牌）\n")

	// 优先从缓存中读取
	authorizer_access_tokenn_key := fmt.Sprintf("authorizer_access_tokenn_key_%s", _this.Context.AppID)
	authorizer_access_tokenn := _this.Context.Cache.Get(authorizer_access_tokenn_key)

	if authorizer_access_tokenn == nil {
		fmt.Printf("%v\n", "令牌失效，重新获取接口凭据.")
		_, err := _this.QueryAuthCode(_this.authorizationCode, _this.componentAccessToken)
		return nil, err
	}

	queryAuth := authorizer_access_tokenn.(*ApiQueryAuthResponse)
	jsonData := ApiAuthorizerTokenRequest{
		ComponentAppId:         _this.ComponentAppId,
		AuthorizerAppId:        queryAuth.AuthorizationInfo.AuthorizerAppId,
		AuthorizerRefreshToken: queryAuth.AuthorizationInfo.AuthorizerRefreshToken,
	}

	jsonString, err := json.Marshal(jsonData)

	if err != nil {
		return nil, err
	}

	result, err := module_common.HttpPostJson(refreshTokenURL+_this.componentAccessToken, string(jsonString))

	if err != nil {
		return nil, err
	}

	authToken := &ApiAuthorizerTokenResponse{}
	err = json.Unmarshal([]byte(result), authToken)

	if err != nil {
		return nil, err
	}

	fmt.Printf("authorizer_access_tokenn:%v\n", authToken.AuthorizerAccessToken)
	fmt.Printf("authorizer_refresh_token:%v\n", authToken.AuthorizerRefreshToken)
	fmt.Printf("expires_in:%v\n", authToken.ExpiresIn)

	// 写入缓存
	// expires := authToken.ExpiresIn - 1500
	// err = _this.Context.Cache.Put(authorizer_access_tokenn_key, authToken, time.Duration(expires)*time.Second)

	// if err != nil {
	// 	return nil, err
	// }

	return authToken, nil
}
