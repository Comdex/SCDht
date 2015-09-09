package controllers

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/utils/pagination"
	"github.com/ylqjgm/SCDht/common"
	"github.com/ylqjgm/SCDht/models"
	"gopkg.in/mgo.v2/bson"
)

// Controller结构
type IndexController struct {
	BaseController
}

// HotList结构
type HotList struct {
	Caption string // 标题
}

// 首页
func (this *IndexController) Index() {
	// 如果当前是简体中文
	if this.Lang == "zh-CN" {
		// 获取配置中的cnhotlist内容
		str := beego.AppConfig.String("cnhotlist")
		// 如果内容长度大于0
		if len(str) > 0 {
			// 以|分割字符串
			strs := strings.Split(str, "|")
			// 定义一个HotList列表
			var hotlist []HotList

			// 对分割的字符串数组循环处理
			for _, s := range strs {
				// 定义一个HotList结构
				h := HotList{
					Caption: s,
				}
				// 将定义的HotList加入到列表中
				hotlist = append(hotlist, h)
			}

			// 设置热门列表
			this.Data["HotList"] = hotlist
		}
	} else {
		// 定义一个Info列表
		var infos []models.SC_Info
		// 获取热门种子列表
		models.GetDataByQuery(models.DbInfo, 0, 20, "-hot", nil, &infos)
		// 设置热门列表
		this.Data["HotList"] = infos
	}

	// 输出模板
	this.TplNames = "index.html"
}

// 搜索页
func (this *IndexController) Search() {
	// 获取搜索关键字
	key := this.Ctx.Input.Param(":k")
	// 获取排序方式
	sort := this.Ctx.Input.Param(":sort")
	// 设置排序方式
	this.Data["Sort"] = sort
	if sort == "" {
		this.Data["Sort"] = "puttime"
		sort = "-hot"
	}

	sort = "-" + sort

	// 设置搜索词
	this.Data["Key"] = key

	// 定义两个SC_Info列表
	var infos, newlist, hots []models.SC_Info

	// 获取热门种子列表
	models.GetDataByQuery(models.DbInfo, 0, 5, "-hot", nil, &hots)
	// 设置热门列表
	this.Data["HotList"] = hots

	// 获取最新入库
	models.GetDataByQuery(models.DbInfo, 0, 10, "-puttime", nil, &newlist)
	// 设置最新入库
	this.Data["NewList"] = newlist

	// 设置查询条件
	query := bson.M{"$or": []bson.M{bson.M{"caption": bson.M{"$regex": bson.RegEx{key, "i"}}}, bson.M{"keys": bson.M{"$regex": bson.RegEx{key, "i"}}}}}

	// 获取种子数量
	count := models.Count(models.DbInfo, query)

	// 设置数量
	this.Data["Nums"] = count

	// 获取分页数据
	page := pagination.NewPaginator(this.Ctx.Request, 15, count)
	// 设置分页数据
	this.Data["paginator"] = page

	// 获取种子列表
	models.GetDataByQuery(models.DbInfo, page.Offset(), 15, sort, query, &infos)
	// 设置种子列表
	this.Data["Lists"] = infos

	// 定义三个SC_Search列表
	var randomlist, lastsearch, relevantsearch []models.SC_Search

	// 设置查询条件
	query = bson.M{"$or": []bson.M{bson.M{"caption": bson.M{"$regex": bson.RegEx{key, "i"}}}}}
	// 获取相关搜索
	models.GetDataByQuery(models.DbSearch, 0, 10, "-searchtime", query, &relevantsearch)
	// 设置相关搜索
	this.Data["RelevantList"] = relevantsearch

	// 获取热门搜索
	models.GetDataByQuery(models.DbSearch, 0, 10, "-views", nil, &randomlist)
	// 设置热门搜索
	this.Data["RandomList"] = randomlist

	// 获取最近搜索
	models.GetDataByQuery(models.DbSearch, 0, 10, "-searchtime", nil, &lastsearch)
	// 设置最后搜索
	this.Data["LastSearch"] = lastsearch

	// 定义一个SC_Search
	var scsearch models.SC_Search
	// 设置关键字
	scsearch.Caption = key
	// 设置查询次数
	scsearch.Views = 1
	// 设置查询到的数据总量
	scsearch.Count = count
	// 保存数据
	scsearch.Save()

	// 自增查询次数
	models.SaveLog(time.Now().Format("20060102"), "searchnums")

	// 输出模板
	this.TplNames = "list.html"
}

// 种子转磁力链
func (this *IndexController) Magnet() {
	if this.Ctx.Request.Method == "POST" {
		// 获取上传文件
		f, h, err := this.GetFile("torrentFile")
		if err != nil {
			// 跳转404
			this.Abort("404")
		}
		// 保证正确关闭
		defer f.Close()

		// 获取最后一个.出现的位置
		index := strings.LastIndex(h.Filename, ".")
		// 获取文件后缀
		ext := strings.ToUpper(h.Filename[index+1:])
		// 若不是种子文件后缀则报错
		if ext != "TORRENT" {
			// 跳转404
			this.Abort("404")
		}

		// 读取种子信息
		meta, err := common.ReadTorrent(f)
		if err != nil || len(meta.InfoHash) != 40 {
			// 跳转404
			this.Abort("404")
		}

		// 获取种子UTF-8格式名称
		caption := meta.Info.Name8
		if caption == "" {
			// 获取失败则获取默认名称
			caption = meta.Info.Name
		}

		// 种子入库
		err = common.PutTorrent(meta)
		if err == nil {
			// 检测infohash是否已经存在
			if models.Has(models.DbHash, bson.M{"infohash": strings.ToUpper(strings.TrimSpace(meta.InfoHash))}) {
				// 设置当前infohash已经入库
				models.SetPut(strings.ToUpper(strings.TrimSpace(meta.InfoHash)))
			} else {
				// 定义一个SC_Hash
				var schash models.SC_Hash
				// 设置SC_Hash
				schash.Hot = 1
				schash.IsPut = true
				schash.InfoHash = strings.ToUpper(strings.TrimSpace(meta.InfoHash))
				// 保存hash数据
				err = schash.Save()
				if err == nil {
					// 自增统计数据
					models.SaveLog(time.Now().Format("20060102"), "dhtnums")
				}
			}

			this.Redirect("/"+strings.ToUpper(strings.TrimSpace(meta.InfoHash)), 302)
		}
	}

	this.TplNames = "magnet.html"
}

// 最新入库
func (this *IndexController) Newly() {
	// 定义两个SC_Info列表
	var infos, hots []models.SC_Info

	// 获取热门种子列表
	models.GetDataByQuery(models.DbInfo, 0, 5, "-hot", nil, &hots)
	// 设置热门列表
	this.Data["HotList"] = hots

	// 获取最新入库列表
	models.GetDataByQuery(models.DbInfo, 0, 15, "-puttime", nil, &infos)
	// 设置最新入库列表
	this.Data["Lists"] = infos

	// 定义一个SC_Search列表
	var wesearch []models.SC_Search

	// 获取大家都在搜
	models.GetDataByQuery(models.DbSearch, 0, 10, "-searchtime", nil, &wesearch)
	// 设置大家都在搜
	this.Data["SearchList"] = wesearch

	this.TplNames = "new.html"
}

// 磁力链转种子
func (this *IndexController) Torrent() {
	if this.Ctx.Request.Method == "POST" {
		// 获取磁力链接
		magnet := this.GetString("magnetLink")
		// 将磁力链接转换为大写格式
		magnet = strings.ToUpper(magnet)

		// 定义一个正则
		re, _ := regexp.Compile(`MAGNET:\?XT=URN:BTIH:([^&]+)`)
		// 匹配磁力链接
		match := re.FindAllString(magnet, -1)
		// 获取infohash
		magnet = strings.Replace(match[0], "MAGNET:?XT=URN:BTIH:", "", -1)
		// 去除空格并转换为大写
		magnet = strings.ToUpper(strings.TrimSpace(magnet))

		if match[0] == "" {
			// 如果匹配不到磁力链接
			if len(magnet) != 40 {
				// 跳转404
				this.Abort("404")
			}
		}

		// 检测infohash是否已经存在
		if !models.Has(models.DbHash, bson.M{"infohash": magnet}) {
			// 定义一个SC_Hash
			var schash models.SC_Hash
			// 设置SC_Hash
			schash.Hot = 1
			schash.IsPut = true
			schash.InfoHash = magnet
			// 保存hash数据
			err := schash.Save()
			if err == nil {
				// 自增统计数据
				models.SaveLog(time.Now().Format("20060102"), "dhtnums")
			}
		}

		// 检测infohash是否已入库过
		if !models.Has(models.DbInfo, bson.M{"infohash": magnet}) {
			// 下载并入库种子
			ret, err := common.PullTorrent(magnet)
			if err != nil || ret != 0 {
				this.Abort("404")
			}
		}

		// 跳转到种子信息页
		this.Redirect("/"+magnet, 302)
	}

	this.TplNames = "torrent.html"
}

// 显示页
func (this *IndexController) View() {
	// 获取InfoHash
	infohash := this.Ctx.Input.Param(":infohash")

	// 将infohash转换为大写
	infohash = strings.ToUpper(infohash)

	// 定义一个SC_Info
	var scinfo models.SC_Info

	// 获取种子信息
	models.GetOneByQuery(models.DbInfo, bson.M{"infohash": infohash}, &scinfo)

	if scinfo.InfoHash == "" {
		// 如果infohash为空或小于40则报错
		this.Abort("404")
	}

	// 设置种子标题
	this.Data["Caption"] = scinfo.Caption

	// 定义两个SC_Info列表
	var hots []models.SC_Info

	// 获取热门种子列表
	models.GetDataByQuery(models.DbInfo, 0, 5, "-hot", nil, &hots)
	// 设置热门列表
	this.Data["HotList"] = hots

	// 自增下载次数
	models.SetAdd(models.DbInfo, bson.M{"infohash": infohash}, "views", true)

	// 设置创建时间
	this.Data["CreateTime"] = scinfo.CreateTime
	// 设置入库时间
	this.Data["PutTime"] = scinfo.PutTime
	// 设置文件大小
	this.Data["Length"] = scinfo.Length
	// 设置关键词
	this.Data["Keys"] = scinfo.Keys
	// 设置种子热度
	this.Data["Hot"] = scinfo.Hot
	// 设置文件数量
	this.Data["FileCount"] = scinfo.FileCount
	// 设置InfoHash
	this.Data["InfoHash"] = scinfo.InfoHash
	// 设置文件列表
	this.Data["FileList"] = scinfo.FileList
	// 设置下载链接
	this.Data["Down"] = fmt.Sprintf("http://btcache.me/torrent/%s", scinfo.InfoHash)

	// 二维码文件是否存在
	if _, err := os.Stat("/static/qrcode/" + scinfo.InfoHash[0:1] + "/" + scinfo.InfoHash[1:2] + "/" + scinfo.InfoHash[2:3] + "/" + scinfo.InfoHash[3:4] + "/" + scinfo.InfoHash[4:5] + "/" + scinfo.InfoHash[5:6] + "/" + scinfo.InfoHash[6:7] + "/" + scinfo.InfoHash + ".png"); err != nil {
		// 存在则设置二维码图片
		this.Data["Qrcode"] = "/static/qrcode/" + scinfo.InfoHash[0:1] + "/" + scinfo.InfoHash[1:2] + "/" + scinfo.InfoHash[2:3] + "/" + scinfo.InfoHash[3:4] + "/" + scinfo.InfoHash[4:5] + "/" + scinfo.InfoHash[5:6] + "/" + scinfo.InfoHash[6:7] + "/" + scinfo.InfoHash + ".png"
	}

	// 自增查看次数
	models.SaveLog(time.Now().Format("20060102"), "viewnums")

	// 输出模板
	this.TplNames = "view.html"
}
