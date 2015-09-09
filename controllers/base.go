package controllers

import (
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/ylqjgm/SCDht/common"
	"github.com/ylqjgm/SCDht/models"
	"gopkg.in/mgo.v2/bson"
)

type BaseController struct {
	beego.Controller
	i18n.Locale
}

// 数据初始化
func (this *BaseController) Prepare() {
	// 获取今天日期
	t := time.Now().Format("20060102")
	// 定义一个日志模型
	var sclog models.SC_Log
	// 查询今天入库数量
	models.GetOneByQuery(models.DbLog, bson.M{"day": t}, &sclog)
	// 添加到模板中
	this.Data["Today"] = sclog.PutNums

	// 获取所有种子数量
	all := models.Count(models.DbInfo, nil)
	// 使用到模板中
	this.Data["All"] = all

	// 判断用户语言
	if this.setLang() {
		i := strings.Index(this.Ctx.Request.RequestURI, "?")
		this.Redirect(this.Ctx.Request.RequestURI[:i], 302)
		return
	}
}

// 设置语言
func (this *BaseController) setLang() bool {
	isNeedRedir := false
	hasCookie := false

	langs := common.Langs

	lang := this.Input().Get("lang")

	if len(lang) == 0 {
		lang = this.Ctx.GetCookie("lang")
		hasCookie = true
	} else {
		isNeedRedir = true
	}

	if !i18n.IsExist(lang) {
		lang = ""
		isNeedRedir = false
		hasCookie = false
	}

	if len(lang) == 0 {
		al := this.Ctx.Request.Header.Get("Accept-Language")

		if len(al) > 4 {
			al := al[:5]

			if i18n.IsExist(al) {
				lang = al
			}
		}
	}

	if len(lang) == 0 {
		lang = "en-US"
		isNeedRedir = false
	}

	if !hasCookie {
		this.Ctx.SetCookie("lang", lang, 60*60*24*365, "/", nil, nil, false)
	}

	this.Data["Lang"] = lang
	this.Data["Langs"] = langs
	this.Data["CurrentLang"] = lang
	this.Lang = lang

	if lang == "zh-CN" {
		this.Data["CN"] = true
	}

	return isNeedRedir
}
