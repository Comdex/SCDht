// 数据库结构
package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// SC_Hash表结构
type SC_Hash struct {
	Id       bson.ObjectId `_id`             // 数据编号
	InfoHash string        `bson:"infohash"` // InfoHash
	Hot      int64         `bson:"hot"`      // Hash热度
	Invalid  int           `bson:"invalid"`  // 失败次数
	IsPut    bool          `bson:"isput"`    // 是否已入库
}

// SC_Info表结构
type SC_Info struct {
	Id         bson.ObjectId `_id`               // 数据编号
	InfoHash   string        `bson:"infohash"`   // InfoHash
	Caption    string        `bson:"caption"`    // 种子名称
	Length     int64         `bson:"length"`     // 种子大小, 单位字节
	Hot        int64         `bson:"hot"`        // 种子热度
	Files      []interface{} `bson:"files"`      // 文件列表
	FileList   []interface{} `bson:"filelist"`   // 文件列表
	FileCount  int64         `bson:"filecount"`  // 种子文件数量
	Keys       []string      `bson:"keys"`       // 种子分词记录
	Views      int64         `bson:"views"`      // 查看次数
	CreateTime time.Time     `bson:"createtime"` // 种子创建时间
	PutTime    time.Time     `bson:"puttime"`    // 种子入库时间
}

// SC_Log表结构
type SC_Log struct {
	Id      bson.ObjectId `_id`            // 数据编号
	Day     string        `bson:"day"`     // 统计日期
	DhtNums int64         `bson:"dhtnums"` // 获取到的infohash数量
	PutNums int64         `bson:"putnums"` // 种子入库数量
}

// SC_Search表结构
type SC_Search struct {
	Id         bson.ObjectId `_id`               // 数据编号
	Caption    string        `bson:"caption"`    // 搜索关键字
	SearchTime time.Time     `bson:"searchtime"` // 搜索时间
	Count      int           `bson:"count"`      // 查询到的资源总量
	Views      int64         `bson:"views"`      // 搜索次数
}
