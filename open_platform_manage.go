// 微信开放平台管理
package gowechat

import (
	"github.com/MrCHI/gowechat/open/component"
)

type OpenPlatformManage struct {
	*Wechat
}

// 返回组件实例
func (_this *OpenPlatformManage) GetComponent() *component.Component {
	return component.NewComponent(_this.Context)
}
