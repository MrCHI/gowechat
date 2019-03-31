package material

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/MrCHI/gowechat/mp/base"
	"github.com/MrCHI/gowechat/util"
	"github.com/MrCHI/gowechat/wxcontext"
)

const (
	addNewsURL          = "https://api.weixin.qq.com/cgi-bin/material/add_news"
	addMaterialURL      = "https://api.weixin.qq.com/cgi-bin/material/add_material"
	delMaterialURL      = "https://api.weixin.qq.com/cgi-bin/material/del_material"
	batchgetMaterialURL = "https://api.weixin.qq.com/cgi-bin/material/batchget_material"
)

//Material 素材管理
type Material struct {
	base.MpBase
}

//NewMaterial init
func NewMaterial(context *wxcontext.Context) *Material {
	material := new(Material)
	material.Context = context
	return material
}

//Article 永久图文素材
type Article struct {
	Title            string `json:"title"`
	ThumbMediaID     string `json:"thumb_media_id"`
	Author           string `json:"author"`
	Digest           string `json:"digest"`
	ShowCoverPic     int    `json:"show_cover_pic"`
	Content          string `json:"content"`
	ContentSourceURL string `json:"content_source_url"`
}

type Error struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type BatchGetResult struct {
	TotalCount int            `json:"total_count"` // 该类型的素材的总数
	ItemCount  int            `json:"item_count"`  // 本次调用获取的素材的数量
	Items      []MaterialInfo `json:"item"`        // 本次调用获取的素材列表
}

type BatchGetResultEx struct {
	Error
	BatchGetResult
}

type MaterialInfo struct {
	MediaId    string `json:"media_id"`    // 素材id
	Name       string `json:"name"`        // 文件名称
	UpdateTime int64  `json:"update_time"` // 最后更新时间
	URL        string `json:"url"`         // 当获取的列表是图片素材列表时, 该字段是图片的URL
}

//reqArticles 永久性图文素材请求信息
type reqArticles struct {
	Articles []*Article `json:"articles"`
}

//resArticles 永久性图文素材返回结果
type resArticles struct {
	util.CommonError

	MediaID string `json:"media_id"`
}

//AddNews 新增永久图文素材
func (material *Material) AddNews(articles []*Article) (mediaID string, err error) {
	req := &reqArticles{articles}

	responseBytes, err := material.HTTPPostJSONWithAccessToken(addNewsURL, req)
	if err != nil {
		return
	}

	var res resArticles
	err = json.Unmarshal(responseBytes, res)
	if err != nil {
		return
	}
	mediaID = res.MediaID
	return
}

//resAddMaterial 永久性素材上传返回的结果
type resAddMaterial struct {
	util.CommonError

	MediaID string `json:"media_id"`
	URL     string `json:"url"`
}

//AddMaterial 上传永久性素材（处理视频需要单独上传）
func (material *Material) AddMaterial(mediaType MediaType, filename string) (mediaID string, url string, err error) {
	if mediaType == MediaTypeVideo {
		err = errors.New("永久视频素材上传使用 AddVideo 方法")
	}
	var accessToken string
	accessToken, err = material.GetAccessToken()
	if err != nil {
		return
	}

	uri := fmt.Sprintf("%s?access_token=%s&type=%s", addMaterialURL, accessToken, mediaType)
	var response []byte
	response, err = util.PostFile("media", filename, uri)
	if err != nil {
		return
	}
	var resMaterial resAddMaterial
	err = json.Unmarshal(response, &resMaterial)
	if err != nil {
		return
	}
	if resMaterial.ErrCode != 0 {
		err = fmt.Errorf("AddMaterial error : errcode=%v , errmsg=%v", resMaterial.ErrCode, resMaterial.ErrMsg)
		return
	}
	mediaID = resMaterial.MediaID
	url = resMaterial.URL
	return
}

type reqVideo struct {
	Title        string `json:"title"`
	Introduction string `json:"introduction"`
}

//AddVideo 永久视频素材文件上传
func (material *Material) AddVideo(filename, title, introduction string) (mediaID string, url string, err error) {
	var accessToken string
	accessToken, err = material.GetAccessToken()
	if err != nil {
		return
	}

	uri := fmt.Sprintf("%s?access_token=%s&type=video", addMaterialURL, accessToken)

	videoDesc := &reqVideo{
		Title:        title,
		Introduction: introduction,
	}
	var fieldValue []byte
	fieldValue, err = json.Marshal(videoDesc)
	if err != nil {
		return
	}

	fields := []util.MultipartFormField{
		{
			IsFile:    true,
			Fieldname: "video",
			Filename:  filename,
		},
		{
			IsFile:    true,
			Fieldname: "description",
			Value:     fieldValue,
		},
	}

	var response []byte
	response, err = util.PostMultipartForm(fields, uri)
	if err != nil {
		return
	}

	var resMaterial resAddMaterial
	err = json.Unmarshal(response, &resMaterial)
	if err != nil {
		return
	}
	if resMaterial.ErrCode != 0 {
		err = fmt.Errorf("AddMaterial error : errcode=%v , errmsg=%v", resMaterial.ErrCode, resMaterial.ErrMsg)
		return
	}
	mediaID = resMaterial.MediaID
	url = resMaterial.URL
	return
}

type reqDeleteMaterial struct {
	MediaID string `json:"media_id"`
}

//DeleteMaterial 删除永久素材
func (material *Material) DeleteMaterial(mediaID string) error {
	_, err := material.HTTPPostJSONWithAccessToken(delMaterialURL, reqDeleteMaterial{mediaID})
	if err != nil {
		return err
	}
	return nil
}

type reqBatchGat struct {
	MaterialType string `json:"type"`
	Offset       int    `json:"offset"`
	Count        int    `json:"count"`
}

// 获取素材列表
func (material *Material) BatchGet(materialType string, offset, count int) ([]byte, error) {
	if offset < 0 {
		offset = 0
	}

	if count <= 0 {
		count = 20
	}

	request := reqBatchGat{
		MaterialType: materialType,
		Offset:       offset,
		Count:        count,
	}

	result, err := material.HTTPPostJSONWithAccessToken(batchgetMaterialURL, request)

	if err != nil {
		return nil, err
	}

	return result, nil
}
