package main

import (
	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/ylqjgm/SCDht/common"
	"github.com/ylqjgm/SCDht/controllers"
	"github.com/ylqjgm/SCDht/models"
)

func main() {
	// 初始化
	models.Init()

	// 启动dht
	go common.Dht()
	// 启动入库
	go common.Put()

	// 主页路由
	beego.Router("/", &controllers.IndexController{}, "get:Index")
	// 搜索页路由
	beego.Router("/search/:k", &controllers.IndexController{}, "get:Search")
	// 搜索页排序路由
	beego.Router("/search/:k/:sort", &controllers.IndexController{}, "get:Search")
	// 种子转磁力链
	beego.Router("/magnet", &controllers.IndexController{}, "*:Magnet")
	// 最新入库
	beego.Router("/new", &controllers.IndexController{}, "get:Newly")
	// 磁力链转种子
	beego.Router("/torrent", &controllers.IndexController{}, "*:Torrent")
	// 显示页路由
	beego.Router("/:infohash", &controllers.IndexController{}, "get:View")
	// 设置静态目录
	beego.SetStaticPath("/static", "static")

	// 增加自定义函数
	beego.AddFuncMap("HightLight", common.HightLight)
	beego.AddFuncMap("SizeFormat", common.Size)
	beego.AddFuncMap("DateFormat", common.DateFormat)
	beego.AddFuncMap("FileFormat", common.FileType)
	beego.AddFuncMap("Thunder", common.Thunder)
	beego.AddFuncMap("FileList", common.TreeShow)
	beego.AddFuncMap("i18n", i18n.Tr)

	// 自定义错误页
	beego.ErrorController(&controllers.ErrorController{})

	// 启动Web
	beego.Run()
}
