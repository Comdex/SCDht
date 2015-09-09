// 数据库操作
package models

import (
	"fmt"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 数据库结构
type DB struct {
	Host    string // MongoDB连接地址
	Port    int    // MongoDB连接端口
	Name    string // MongoDB数据库名
	User    string // MongoDB连接用户名
	Pass    string // MongoDB连接密码
	ShowMsg bool   // 是否显示信息
}

var (
	Session  *mgo.Session    // 数据库连接对象
	DbHash   *mgo.Collection // Hash表对象
	DbInfo   *mgo.Collection // 种子信息表对象
	DbLog    *mgo.Collection // 每日统计信息表
	DbSearch *mgo.Collection // 搜索统计表
	DbConfig *DB             // 数据库配置信息
)

// 初始化数据库
func Init() {
	// 获取数据库连接端口
	dbport, _ := beego.AppConfig.Int("dbport")
	// 获取是否允许显示信息
	showmsg, _ := beego.AppConfig.Bool("showmsg")

	// 初始化数据库配置信息
	DbConfig = &DB{
		Host:    beego.AppConfig.String("dbhost"), // 配置数据库地址
		Port:    dbport,                           // 配置数据库端口
		Name:    beego.AppConfig.String("dbname"), // 配置数据库名称
		User:    beego.AppConfig.String("dbuser"), // 配置数据库用户名
		Pass:    beego.AppConfig.String("dbpass"), // 配置数据库密码
		ShowMsg: showmsg,                          // 是否允许显示信息
	}

	// 连接用户名及密码
	userAndPass := DbConfig.User + ":" + DbConfig.Pass + "@"

	// 如果没有设置则表示不需要用户名和密码
	if DbConfig.User == "" || DbConfig.Pass == "" {
		userAndPass = ""
	}

	// 设置连接字符串
	url := "mongodb://" + userAndPass + DbConfig.Host + ":" + fmt.Sprintf("%d", DbConfig.Port) + "/" + DbConfig.Name

	// 定义一个错误变量
	var err error
	// 定义一个索引变量
	var index mgo.Index
	// 连接数据库
	Session, err = mgo.Dial(url)
	if err != nil {
		// 失败则报错
		panic(err)
	}

	// 配置为monotonic驱动
	Session.SetMode(mgo.Monotonic, true)

	// 连接Hash表
	DbHash = Session.DB(DbConfig.Name).C("SC_Hash")
	// 设置Hash表索引
	index = mgo.Index{
		Key:        []string{"infohash"}, // 索引键
		Unique:     true,                 // 唯一索引
		DropDups:   true,                 // 存在数据后创建, 则自动删除重复数据
		Background: true,                 // 不长时间占用写锁
	}
	// 创建索引
	DbHash.EnsureIndex(index)

	// 连接种子信息表
	DbInfo = Session.DB(DbConfig.Name).C("SC_Info")
	// 设置种子表唯一索引
	index = mgo.Index{
		Key:        []string{"infohash"}, // 索引键
		Unique:     true,                 // 唯一索引
		DropDups:   true,                 // 存在数据后创建, 则自动删除重复数据
		Background: true,                 // 不长时间占用写锁
	}
	// 创建索引
	DbInfo.EnsureIndex(index)
	// 设置种子表标题索引
	index = mgo.Index{
		Key:        []string{"caption"}, // 索引键
		Background: true,                // 不长时间占用写锁
	}
	// 创建索引
	DbInfo.EnsureIndex(index)

	// 链接每日统计信息表
	DbLog = Session.DB(DbConfig.Name).C("SC_Log")
	// 设置统计表唯一索引
	index = mgo.Index{
		Key:        []string{"day"}, //索引键
		Unique:     true,            // 唯一索引
		DropDups:   true,            // 存在数据后创建, 则自动删除重复数据
		Background: true,            // 不长时间占用写锁
	}
	// 创建索引
	DbLog.EnsureIndex(index)

	// 链接搜索统计信息表
	DbSearch = Session.DB(DbConfig.Name).C("SC_Search")
	// 设置统计表唯一索引
	index = mgo.Index{
		Key:        []string{"caption"}, // 索引键
		Unique:     true,                // 唯一索引
		DropDups:   true,                // 存在数据后创建, 则自动删除重复数据
		Background: true,                // 不长时间占用写锁
	}
	// 创建索引
	DbSearch.EnsureIndex(index)
}

/********************* 公共操作 *********************/

// 创建一条数据
func Insert(collection *mgo.Collection, data interface{}) error {
	return collection.Insert(data)
}

// 更新一条数据
func Update(collection *mgo.Collection, query, data interface{}) error {
	return collection.Update(query, data)
}

// 删除一条数据
func Delete(collection *mgo.Collection, query interface{}) error {
	return collection.Remove(query)
}

// 通过Id获取一条数据
func GetOneById(collection *mgo.Collection, id bson.ObjectId, val interface{}) {
	collection.FindId(id).One(val)
}

// 通过查询条件获取一条数据
func GetOneByQuery(collection *mgo.Collection, query, val interface{}) {
	collection.Find(query).One(val)
}

// 通过查询条件获取所有数据
func GetAllByQuery(collection *mgo.Collection, query, val interface{}) {
	collection.Find(query).All(val)
}

// 通过查询获取指定数量与排序的数据
func GetDataByQuery(collection *mgo.Collection, start, length int, fields string, query interface{}, val interface{}) {
	collection.Find(query).Limit(length).Skip(start).Sort(fields).All(val)
}

// 获取统计数据
func Count(collection *mgo.Collection, query interface{}) int {
	cnt, err := collection.Find(query).Count()
	if err != nil {
		fmt.Println(err.Error())
	}

	return cnt
}

// 数据是否存在
func Has(collection *mgo.Collection, query interface{}) bool {
	if Count(collection, query) > 0 {
		return true
	}

	return false
}

// 数据自增或自减
func SetAdd(collection *mgo.Collection, query interface{}, field string, add bool) error {
	if add {
		return collection.Update(query, bson.M{"$inc": bson.M{field: 1}})
	} else {
		return collection.Update(query, bson.M{"$inc": bson.M{field: -1}})
	}
}

/********************* SC_Hash 操作 *********************/

// 保存Hash数据
func (this *SC_Hash) Save() error {
	// 查询此条数据是否存在
	if Has(DbHash, bson.M{"infohash": this.InfoHash}) {
		// 如果存在则进行自增处理
		return SetAdd(DbHash, bson.M{"infohash": this.InfoHash}, "hot", true)
	} else {
		// 创建编号
		this.Id = bson.NewObjectId()
		// 添加数据
		return Insert(DbHash, this)
	}
}

// 验证此Hash是否已经入库
func IsPut(hash string) bool {
	// 定义一个SC_Hash
	var schash SC_Hash
	// 获取Hash信息
	GetOneByQuery(DbHash, bson.M{"infohash": hash}, &schash)

	if schash.InfoHash == "" {
		return false
	}

	// 返回是否入库
	return schash.IsPut
}

// 设置Hash为已入库状态
func SetPut(hash string) error {
	return Update(DbHash, bson.M{"infohash": hash}, bson.M{"$set": bson.M{"isput": true}})
}

/********************* SC_Info 操作 *********************/

// 保存种子数据
func (this *SC_Info) Save() error {
	// 查询此条数据是否存在
	if Has(DbInfo, bson.M{"infohash": this.InfoHash}) {
		// 如果存在则更新
		return Update(DbInfo, bson.M{"infohash": this.InfoHash}, bson.M{"$set": bson.M{"caption": this.Caption, "length": this.Length, "hot": this.Hot, "filecount": this.FileCount, "createtime": this.CreateTime, "puttime": this.PutTime}})
	} else {
		// 创建编号
		this.Id = bson.NewObjectId()
		// 添加数据
		return Insert(DbInfo, this)
	}
}

// 修改热度信息
func SetHot(hash string) error {
	// 查询种子信息表中是否存在此hash
	if Has(DbInfo, bson.M{"infohash": hash}) {
		// 修改种子热度
		return SetAdd(DbInfo, bson.M{"infohash": hash}, "hot", true)
	}

	return nil
}

/********************* SC_Log 操作 *********************/

// 保存统计数据
func SaveLog(day, field string) error {
	// 查询此条数据是否存在
	if !Has(DbLog, bson.M{"day": day}) {
		// 定义一个SC_Log
		var sclog SC_Log
		// 设置日期
		sclog.Day = day
		// 判断设置字段
		switch field {
		case "dhtnums":
			sclog.DhtNums = 1
			break
		case "putnums":
			sclog.PutNums = 1
			break
		}
		// 创建编号
		sclog.Id = bson.NewObjectId()
		// 添加数据
		return Insert(DbLog, sclog)
	} else {
		// 自增
		return SetAdd(DbLog, bson.M{"day": day}, field, true)
	}
}

/********************* SC_Search 操作 *********************/

// 保存搜索数据
func (this *SC_Search) Save() error {
	// 查询此条数据是否存在
	if Has(DbSearch, bson.M{"caption": this.Caption}) {
		// 存在则自增
		SetAdd(DbSearch, bson.M{"caption": this.Caption}, "views", true)
		// 更新查询时间
		return Update(DbSearch, bson.M{"caption": this.Caption}, bson.M{"$set": bson.M{"searchtime": bson.Now()}})
	} else {
		// 创建编号
		this.Id = bson.NewObjectId()
		// 设置查询时间
		this.SearchTime = bson.Now()
		// 添加数据
		return Insert(DbSearch, this)
	}
}
