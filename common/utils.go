package common

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"html/template"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/wangbin/jiebago"
	"github.com/ylqjgm/SCDht/models"
	"github.com/zeebo/bencode"
	"gopkg.in/mgo.v2/bson"
)

// 目录信息结构
type Directory struct {
	Name  string
	Dirs  []*Directory
	Files []*File
}

// 种子文件信息结构
type FileDict struct {
	Length int64    "length"
	Path   []string "path"
	Path8  []string "path.utf-8" // 文件utf-8格式路径数组
	Md5sum string   "md5sum"
}

// 种子Info信息结构
type InfoDict struct {
	FileDuration []int64 "file-duration"
	FileMedia    []int64 "file-media"

	// Single file
	Name   string "name"
	Name8  string "name.utf-8" // 种子utf-8名称
	Length int64  "length"
	Md5sum string "md5sum"

	// Multiple files
	Files       []FileDict "files"
	PieceLength int64      "piece length"
	Pieces      string     "pieces"
	Private     int64      "private"
}

// 种子信息结构
type MetaInfo struct {
	Info         InfoDict "info"
	InfoHash     string   "info hash"
	Announce     string   "announce"
	CreationDate int64    "creation date"
	Comment      string   "comment"
	CreatedBy    string   "created by"
	Encoding     string   "encoding"
}

// 定义一个分词对象
var (
	seg   jiebago.Segmenter
	Langs []string
	root  Directory
)

func init() {
	// 载入分词字典
	seg.LoadDictionary("dict.txt")
	SetLang()
}

// 设置语言
func SetLang() {
	// 设定语言类型
	langs := "en-US|zh-CN|ja-JP|zh-TW|ko-KR"

	// 对语言进行循环
	for _, lang := range strings.Split(langs, "|") {
		// 去除首尾空格
		lang = strings.TrimSpace(lang)
		// 设置文件路径
		files := []string{"conf/locale_" + lang + ".ini"}

		if fh, err := os.Open(files[0]); err == nil {
			// 打开正确则关闭
			fh.Close()
		} else {
			// 打开失败则报错
			files = nil
			beego.Error(err.Error())
		}

		if err := i18n.SetMessage(lang, "conf/locale_"+lang+".ini", files...); err != nil {
			// 载入语言文件失败则报错并退出
			beego.Error("Fail to set message file: " + err.Error())
			os.Exit(2)
		}
	}

	// 列出所有语言文件
	Langs = i18n.ListLangs()
}

// 分词
func Sego(str string) []string {
	// 使用搜索引擎模式分词
	s := seg.CutForSearch(str, true)

	// 定义一个字符串列表保存分词单词
	var result []string
	// 定义一个map用以去重
	has := make(map[string]string)

	// 循环处理分词单词
	for word := range s {
		// 如果在map中不存在
		if _, ok := has[word]; !ok {
			// 如果单词不为空或空格
			if word != "" && word != " " {
				// 将单词加入字符串列表
				result = append(result, word)
				// 将单词加入map中
				has[word] = word
			}
		}
	}

	// 返回单词列表
	return result
}

// 高亮着色
func HightLight(str, key string) interface{} {
	// 定义一个正则表达式, 以忽略大小写的方式进行匹配
	re, _ := regexp.Compile("(?i:" + key + ")")
	// 替换着色
	str = re.ReplaceAllString(str, `<span class="highlight">$0</span>`)
	// 返回转换为HTML格式的分词列表
	return UnEscaped(str)
}

// 递归循环输出树结构
func TreeShow(dirs []*Directory) interface{} {
	// 定义一个变量保存要输出的内容
	str := ""

	for _, ds := range dirs {
		for _, s := range ds.Dirs {
			str += fmt.Sprintf(`<li class="closed"><span class="folder">%s</span><ul>`, s.Name)

			if len(s.Dirs) > 0 {
				str += TreeShow(s.Dirs).(string)
			}

			if len(s.Files) > 0 {
				for _, f := range s.Files {
					str += fmt.Sprintf(`<li><span><i class="fa %s"></i> %s<small>%s</small></span></li>`, FileType(f.Path), f.Path, Size(f.Length))
				}
			}

			str += `</ul>`
		}
	}

	return UnEscaped(str)
}

// 根据文件名返回文件类型
func FileType(str string) string {
	// 获取最后一次出现.的位置
	index := strings.LastIndex(str, ".")
	// 获取后缀名并转换为大写格式
	ext := strings.ToUpper(str[index+1:])

	// 对后缀进行判断
	switch ext {
	case "RAR", "ISO", "ZIP", "ARJ", "GZ", "Z", "ACE", "AIFC", "CAB", "7Z":
		// 压缩文件
		return "fa-file-archive-o"
	case "MP3", "WMA":
		// 音频文件
		return "fa-file-audio-o"
	case "CS", "C", "GO", "PHP", "JAVA":
		// 代码文件
		return "fa-file-code-o"
	case "XLS", "XLSX", "XLSM", "XLTX", "XLTM", "XLSB", "XLAM":
		// Execl文件
		return "fa-file-excel-o"
	case "JPG", "JPEG", "PNG", "BMP", "GIF":
		// 图片文件
		return "fa-file-image-o"
	case "MKV", "AVI", "RM", "RMVB", "WMV", "MP4":
		// 视频文件
		return "fa-file-video-o"
	case "PPT", "PPTX", "PPTM", "PPSX", "POTX", "POTM", "PPAM":
		// PPT文件
		return "fa-file-powerpoint-o"
	case "RTF", "TXT":
		// 文本文件
		return "fa-file-text-o"
	case "DOC", "DOCX", "DOCM", "DOTX", "DOTM":
		// word文件
		return "fa-file-word-o"
	case "PDF":
		// pdf文件
		return "fa-file-pdf-o"
	}

	return "fa-file-o"
}

// 转换字节数为对应大小格式
func Size(length int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	var mod float64
	mod = 1024.0
	size := float64(length)
	i := 0
	for size >= mod {
		size /= mod
		i++
	}

	return fmt.Sprintf("%.2f %s", size, units[i])
}

// 格式化日期1
func DateFormat(t time.Time) string {
	// 获取时间戳
	unix := time.Now().Unix() - t.Unix()

	// 计算年数
	year := unix / 31104000
	// 将年数所用时间减去
	unix = unix - (year * 31104000)
	// 计算月数
	month := unix / 2592000
	// 将月数所用时间减去
	unix = unix - (month * 2592000)
	// 计算天数
	day := unix / 86400
	// 将天数所用时间减去
	unix = unix - (day * 86400)
	// 计算小时
	hour := unix / 3600
	// 将小时所用时间减去
	unix = unix - (hour * 3600)
	// 计算分钟
	min := unix / 60
	// 将分钟所用时间减去
	unix = unix - (min * 60)

	if year > 0 {
		// 如果年数大于0则返回年数
		return fmt.Sprintf("%d年前", year)
	}

	if month > 0 {
		// 如果月数大于0则返回月数
		return fmt.Sprintf("%d月前", month)
	}

	if day > 0 {
		// 如果天数大于0则返回天数
		return fmt.Sprintf("%d天前", day)
	}

	if hour > 0 {
		// 如果小时大于0则返回小时
		return fmt.Sprintf("%d小时前", hour)
	}

	if min > 0 {
		// 如果分钟大于0则返回分钟
		return fmt.Sprintf("%d分钟前", min)
	}

	// 返回秒数
	return fmt.Sprintf("%d秒前", unix)
}

// InfoHash转迅雷链接
func Thunder(infohash string) string {
	// 先获取磁力链接
	magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s", strings.ToUpper(infohash))

	return base64.StdEncoding.EncodeToString([]byte("AA" + magnet + "ZZ"))
}

// 转换字符串为html
func UnEscaped(x string) interface{} { return template.HTML(x) }

// 获取string
func getString(m map[string]interface{}, k string) string {
	if v, ok := m[k]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt64(m map[string]interface{}, k string) int64 {
	if v, ok := m[k]; ok {
		if s, ok := v.(int64); ok {
			return s
		}
	}

	return time.Now().Unix()
}

// 读取种子信息
func ReadTorrent(r io.Reader) (meta MetaInfo, err error) {
	// 读取文件信息
	s, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	err = bencode.DecodeBytes(s, &meta)
	if err != nil {
		return
	}

	hash := sha1.New()
	err = bencode.NewEncoder(hash).Encode(meta.Info)
	if err != nil {
		return
	}

	meta.InfoHash = fmt.Sprintf("%02X", string(hash.Sum(nil)))

	return
	/*var m interface{}
	m, err = bencode.Decode(r)
	if err != nil {
		return
	}

	topMap, ok := m.(map[string]interface{})
	if !ok {
		return
	}

	infoMap, ok := topMap["info"]
	if !ok {
		return
	}

	var b bytes.Buffer
	if err = bencode.Marshal(&b, infoMap); err != nil {
		return
	}

	hash := sha1.New()
	hash.Write(b.Bytes())

	err = bencode.Unmarshal(&b, &meta.Info)
	if err != nil {
		return
	}

	meta.InfoHash = fmt.Sprintf("%X", string(hash.Sum(nil)))
	meta.Announce = getString(topMap, "announce")
	meta.Comment = getString(topMap, "comment")
	meta.CreatedBy = getString(topMap, "created by")
	meta.CreationDate = getInt64(topMap, "creation date")
	meta.Encoding = getString(topMap, "encoding")

	return*/
}

// 查找目录
func (d *Directory) findDir(name string) *Directory {
	// 循环目录列表
	for i := 0; i < len(d.Dirs); i++ {
		// 如果目录存在
		if d.Dirs[i].Name == name {
			// 返回当前目录
			return d.Dirs[i]
		}
	}

	// 返回空数据
	return nil
}

// 创建目录
func (d *Directory) makeDir(name string) *Directory {
	// 创建一个目录
	dir := &Directory{
		Name: name,
	}
	// 将目录加入自身
	d.Dirs = append(d.Dirs, dir)

	// 返回目录
	return dir
}

// 查找文件
func (d *Directory) findFile(name string) *File {
	// 循环文件列表
	for i := 0; i < len(d.Files); i++ {
		// 如果文件存在
		if d.Files[i].Path == name {
			// 返回当前文件
			return d.Files[i]
		}
	}

	// 返回空数据
	return nil
}

// 创建文件
func (d *Directory) makeFile(name string, length int64) *File {
	// 创建一个文件
	file := &File{
		Path:   name,
		Length: length,
	}
	// 将文件加入自身
	d.Files = append(d.Files, file)

	// 返回文件
	return file
}

// 目录树结构
func treeDir(path []string, length int64) {
	// 引用root变量
	parent := &root

	// 循环分割的路径
	for i := 0; i < len(path)-1; i++ {
		// 查找目录看是否存在了
		dir := parent.findDir(path[i])
		// 如果不存在
		if dir == nil {
			// 创建目录
			dir = parent.makeDir(path[i])
		}
		// 目录复制
		parent = dir
	}

	// 获取文件
	file := path[len(path)-1]
	// 查找文件是否存在
	if parent.findFile(file) == nil {
		// 不存在则创建文件
		parent.makeFile(file, length)
	}
}

// 种子入库
func PutTorrent(metaTorrent MetaInfo) error {
	// 定义一个SC_Info
	var scinfo models.SC_Info

	// 如果有utf-8格式名称
	if metaTorrent.Info.Name8 != "" {
		// 直接使用utf-8格式的
		scinfo.Caption = strings.TrimSpace(metaTorrent.Info.Name8)
	} else {
		// 否则使用非utf-8格式
		scinfo.Caption = strings.TrimSpace(metaTorrent.Info.Name)
	}

	// 设置infohash, 并转换为大写格式
	scinfo.InfoHash = strings.ToUpper(metaTorrent.InfoHash)

	// 如果没有获取到creationdate信息
	if metaTorrent.CreationDate == 0 {
		// 将creationdate设置为当前时间
		scinfo.CreateTime = time.Now()
	} else {
		// 否则为获取到的时间
		scinfo.CreateTime = time.Unix(metaTorrent.CreationDate, 0)
	}

	// 定义一个File
	var file File
	// 定义一个File列表保存路径信息
	var spath []File

	// 判断文件列表是否大于0
	if len(metaTorrent.Info.Files) > 0 {
		// 循环处理文件列表
		for _, FileDict := range metaTorrent.Info.Files {
			// 加上文件长度
			scinfo.Length += FileDict.Length
			// 文件数量+1
			scinfo.FileCount += 1
			// 设置文件长度
			file.Length = FileDict.Length
			// 清空文件路径
			file.Path = ""

			if FileDict.Path8 != nil {
				// 如果存在utf8编码则使用utf8编码
				for _, path := range FileDict.Path8 {
					file.Path += strings.TrimSpace(path) + "/"
				}
			} else {
				// 否则使用默认
				for _, path := range FileDict.Path {
					file.Path += strings.TrimSpace(path) + "/"
				}
			}

			// 获取最后一个/的位置
			index := strings.LastIndex(file.Path, "/")
			if index <= len(file.Path) && index != -1 {
				// 将最后一个/去掉
				file.Path = file.Path[0:index]
			}
			// 将文件信息加入列表
			scinfo.Files = append(scinfo.Files, file)
			// 加入路径信息
			spath = append(spath, file)
		}
	} else {
		// 设置文件长度
		scinfo.Length = metaTorrent.Info.Length
		// 设置文件数量
		scinfo.FileCount = 1
		// 设置文件路径
		file.Path = scinfo.Caption
		// 设置文件长度
		file.Length = scinfo.Length
		// 将文件本身加入列表
		scinfo.Files = append(scinfo.Files, file)
		// 加入路径信息
		spath = append(spath, file)
	}

	// 对路径数组循环
	for _, f := range spath {
		// 分割路径
		p := strings.Split(f.Path, "/")
		// 创建树结构
		treeDir(p, f.Length)
	}

	// 设置树结构
	scinfo.FileList = append(scinfo.FileList, root)
	// 清空
	root.Dirs = nil
	root.Files = nil

	// 设置文件热度为1
	scinfo.Hot = 1

	// 如果都没有问题
	if scinfo.Caption != "" && scinfo.FileCount > 0 && scinfo.InfoHash != "" {
		// 定义一个正则
		re, _ := regexp.Compile("\\pP|\\pS")
		// 去除种子名称中的所有符号
		caption := re.ReplaceAllString(scinfo.Caption, " ")
		// 检测是否存在
		if !models.Has(models.DbInfo, bson.M{"infohash": scinfo.InfoHash}) {
			// 设置种子分词
			scinfo.Keys = Sego(caption)
			// 设置种子发布时间
			scinfo.PutTime = time.Now()
			// 保存种子信息
			err := scinfo.Save()
			if err == nil {
				// 设置当前hash已经入库
				models.SetPut(scinfo.InfoHash)
				// 自增统计数据
				models.SaveLog(time.Now().Format("20060102"), "putnums")

				// 定义一个Qrcode对象
				qr := &Qrcode{
					Version:        0,
					Level:          ECLevelM,
					ModuleSize:     7,
					QuietZoneWidth: 0,
				}

				// 生成二维码
				img, _ := qr.Encode("magnet:?xt=urn:btih:" + scinfo.InfoHash)
				// 截取infohash作为目录
				dir := "./static/qrcode/" + scinfo.InfoHash[0:1] + "/" + scinfo.InfoHash[1:2] + "/" + scinfo.InfoHash[2:3] + "/" + scinfo.InfoHash[3:4] + "/" + scinfo.InfoHash[4:5] + "/" + scinfo.InfoHash[5:6] + "/" + scinfo.InfoHash[6:7]
				// 创建目录
				os.MkdirAll(dir, 0777)
				// 创建二维码图片文件
				f, _ := os.Create(fmt.Sprintf("%s/%s.png", dir, scinfo.InfoHash))
				// 保证正确关闭
				defer f.Close()
				// 将二维码写入文件
				png.Encode(f, img)
			}

			return err
		} else {
			// 如果没有设置为入库
			if !models.IsPut(scinfo.InfoHash) {
				// 设置为入库
				return models.SetPut(scinfo.InfoHash)
			}
		}
	}

	return nil
}

// 获取bitcomet种子库key
func GetKey(hash string) string {
	// 将infohash转换为小写
	hash = strings.ToLower(hash)
	// 获取infohash一半的长度
	count := len(hash) / 2
	// 定义一个byte列表
	var hashHex []byte

	// 循环将infohash编码
	for i := 0; i < count; i++ {
		// 每次截取2个字符, 并转换为十六进制数字
		val, _ := strconv.ParseInt(hash[i*2:i*2+2], 16, 0)
		// 将十六进制数字转换为byte格式并加入列表中
		hashHex = append(hashHex, byte(val))
	}

	// 定义bitcomet的编码字符串
	bc := "bc" + string(hashHex) + "torrent"

	// 对字符串进行sha1编码
	t := sha1.New()
	io.WriteString(t, bc)
	key := fmt.Sprintf("%x", t.Sum(nil))

	// 返回key
	return key
}
