package oauth

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/MrCHI/gowechat/util"
	"github.com/MrCHI/gowechat/wxcontext"
)

const (
	redirectOauthURL      = "https://open.weixin.qq.com/connect/oauth2/authorize?appid=%v&redirect_uri=%v&response_type=%v&scope=%v&state=%v&component_appid=%v#wechat_redirect"
	accessTokenURL        = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	refreshAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/component/refresh_token?%s"
	checkAccessTokenURL   = "https://api.weixin.qq.com/sns/auth?access_token=%s&openid=%s"
	userInfoURL           = "https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN"
)

// 保存用户授权信息
type Oauth struct {
	*wxcontext.Context
}

// 实例化授权信息
func NewOauth(context *wxcontext.Context) *Oauth {
	auth := new(Oauth)
	auth.Context = context
	return auth
}

// 代公众号发起网页授权
func (oauth *Oauth) GetRedirectURL(redirectURI, serviceAppId, scope, state, componentAppId string) string {
	urlStr := url.QueryEscape(redirectURI)
	return fmt.Sprintf(redirectOauthURL, serviceAppId, urlStr, "code", scope, state, componentAppId)
}

//  代公众号发起网页授权，微信服务器回调
func (oauth *Oauth) AuthCallBack(appid, state, code string) {

	fmt.Printf("appid:%v\n", appid)
	fmt.Printf("state:%v\n", state)
	fmt.Printf("code:%v\n", code)

	// 保存状态
	webCode := code

	_ = webCode
}

// ResAccessToken 获取用户授权access_token的返回结果
type ResAccessToken struct {
	util.CommonError

	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
}

// 通过code换取access_token
func (oauth *Oauth) GetUserAccessToken(code string) (result ResAccessToken, err error) {
	urlStr := fmt.Sprintf(accessTokenURL, oauth.AppID, oauth.AppSecret, code)
	var response []byte
	response, err = util.HTTPGet(urlStr)
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("GetUserAccessToken error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	return
}

// 刷新access_token
func (oauth *Oauth) RefreshAccessToken(serviceAppId, componentAppId, componentAccessToken, refreshToken string) (result ResAccessToken, err error) {
	urlStr := fmt.Sprintf(refreshAccessTokenURL, "appid="+serviceAppId+
		"&grant_type="+"authorization_code"+
		"&component_appid="+componentAppId+
		"&component_access_token="+componentAccessToken+
		"&refresh_token="+refreshToken)

	var response []byte
	response, err = util.HTTPGet(urlStr)
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("GetUserAccessToken error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	return
}

// 检验access_token是否有效
func (oauth *Oauth) CheckAccessToken(accessToken, openID string) (b bool, err error) {
	urlStr := fmt.Sprintf(checkAccessTokenURL, accessToken, openID)
	var response []byte
	response, err = util.HTTPGet(urlStr)
	if err != nil {
		return
	}
	var result util.CommonError
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		b = false
		return
	}
	b = true
	return
}

//UserInfo 用户授权获取到用户信息
type UserInfo struct {
	util.CommonError

	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	Unionid    string   `json:"unionid"`
}

// GetUserInfo 如果scope为 snsapi_userinfo 则可以通过此方法获取到用户基本信息
func (oauth *Oauth) GetUserInfo(accessToken, openID string) (result UserInfo, err error) {
	urlStr := fmt.Sprintf(userInfoURL, accessToken, openID)
	var response []byte
	response, err = util.HTTPGet(urlStr)
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("GetUserInfo error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	return
}
