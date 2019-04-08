//Package gowechat 一个简单易用的wechat封装.
package gowechat

import (
	"fmt"
	"sync"

	"github.com/MrCHI/gowechat/wxcontext"
	"github.com/astaxie/beego/cache"
)

//memCache if wxcontext.Config no cache, this will give a default memory cache.
var memCache cache.Cache

// Wechat struct
type Wechat struct {
	Context *wxcontext.Context
}

// NewWechat init
func NewWechat(cfg wxcontext.Config) *Wechat {
	context := new(wxcontext.Context)
	initContext(&cfg, context)
	return &Wechat{context}
}

func initContext(cfg *wxcontext.Config, context *wxcontext.Context) {
	if cfg.Cache == nil {
		if memCache == nil {
			memCache, _ = cache.NewCache("memory", `{"interval":60}`)
		}
		cfg.Cache = memCache
	}
	context.Config = cfg

	context.SetAccessTokenLock(new(sync.RWMutex))
	context.SetJsAPITicketLock(new(sync.RWMutex))

}

//MchMgr 商户平台
func (wc *Wechat) MchMgr() (mch *MchMgr, err error) {
	err = wc.checkCfgMch()
	if err != nil {
		return
	}
	mch = new(MchMgr)
	mch.Wechat = wc
	return
}

//MpMgr 公众平台
func (wc *Wechat) MpMgr() (mp *MpMgr, err error) {
	err = wc.checkCfgBase()
	if err != nil {
		return
	}
	mp = new(MpMgr)
	mp.Wechat = wc
	return
}

// 开放平台管理实例
func (wc *Wechat) GetOpenPlatform() (op *OpenPlatformManage, err error) {
	err = wc.checkOpenPlatformConfig()

	if err != nil {
		return nil, err
	}

	op = new(OpenPlatformManage)
	op.Wechat = wc

	return op, nil
}

//checkCfgBase 检查配置基本信息
func (wc *Wechat) checkCfgBase() (err error) {
	if wc.Context.AppID == "" {
		return fmt.Errorf("%s", "配置中没有AppID")
	}
	if wc.Context.AppSecret == "" {
		return fmt.Errorf("%s", "配置中没有AppSecret")
	}
	if wc.Context.Token == "" {
		return fmt.Errorf("%s", "配置中没有Token")
	}
	return
}

// 检查商务平台参数
func (wc *Wechat) checkCfgMch() (err error) {
	err = wc.checkCfgBase()
	if err != nil {
		return
	}
	if wc.Context.MchID == "" {
		return fmt.Errorf("%s", "配置中没有MchID")
	}
	if wc.Context.MchAPIKey == "" {
		return fmt.Errorf("%s", "配置中没有MchAPIKey")
	}
	if wc.Context.SslCertFilePath == "" && wc.Context.SslCertContent == "" {
		return fmt.Errorf("%s", "配置中没有SslCert")
	}
	if wc.Context.SslKeyFilePath == "" && wc.Context.SslKeyContent == "" {
		return fmt.Errorf("%s", "配置中没有SslKey")
	}
	//初始化 http client, 有错误会出错误
	err = wc.Context.InitHTTPClients()
	return
}

// 检查开放平台参数
func (wc *Wechat) checkOpenPlatformConfig() (err error) {
	if wc.Context.ComponentAppId == "" {
		return fmt.Errorf("%s", "请指定第三方平台组件APPID")
	}

	if wc.Context.ComponentAppSecret == "" {
		return fmt.Errorf("%s", "请指定第三方平台组件SECRET")
	}

	if wc.Context.ComponentAppToken == "" {
		return fmt.Errorf("%s", "第三方平台组件TOKEN")
	}

	if wc.Context.ComponentAppKey == "" {
		return fmt.Errorf("%s", " 第三方平台组件AESKEY")
	}

	return nil
}
