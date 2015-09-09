// 入库操作
package common

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/ylqjgm/SCDht/models"
	"gopkg.in/mgo.v2/bson"
)

type File struct {
	Path   string // 文件路径
	Length int64  // 文件长度
}

// 种子入库操作
func PullTorrent(hash string) (int, error) {
	// 定义一个int变量t
	t := 0

	// 无限循环进行入库
	for {
		// 定义url和host变量
		var url, host string
		// 将infohash转换为大写格式
		hash = strings.ToUpper(hash)

		// 查看t为多少
		switch t {
		case 0:
			// 为0则使用torrent-cache.bitcomet.org下载
			url = fmt.Sprintf("http://torrent-cache.bitcomet.org:36869/get_torrent?info_hash=%s&size=226920869&key=%s", strings.ToLower(hash), GetKey(hash))
			host = "torrent-cache.bitcomet.org"
			break
		case 1:
			// 为1则使用bt.box.n0808.com下载
			url = fmt.Sprintf("http://bt.box.n0808.com/%s/%s/%s.torrent", hash[0:2], hash[len(hash)-2:], hash)
			host = "bt.box.n0808.com"
			break
		case 2:
			// 为2则使用torcache.net下载
			url = fmt.Sprintf("https://torcache.net/torrent/%s.torrent", hash)
			host = "torcache.net"
			break
		default:
			// 对infohash进行自增处理
			models.SetAdd(models.DbHash, bson.M{"infohash": hash}, "invalid", true)
			return 1, nil
		}

		// t自增
		t++

		// 新建请求
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			// 失败则跳过本次循环
			continue
		}

		// 设置头部信息
		req.Header.Add("User-Agent", "Mozilla/5.0")
		req.Header.Add("Host", host)
		req.Header.Add("Accept", "*/*")
		req.Header.Add("Connection", "Keep-Alive")

		// 设置超时时间
		client := &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					deadline := time.Now().Add(3 * time.Second)
					c, err := net.DialTimeout(netw, addr, time.Second*3)
					if err != nil {
						return nil, err
					}
					c.SetDeadline(deadline)
					return c, nil
				},
			},
		}

		// 请求链接
		resp, err := client.Do(req)
		if err != nil {
			// 失败则跳过本次循环
			continue
		}
		// 保证关闭
		defer resp.Body.Close()

		// 定义一个MetaInfo
		var metaTorrent MetaInfo
		// 读取种子信息
		metaTorrent, err = ReadTorrent(resp.Body)
		if err != nil {
			// 失败则跳过本次循环
			continue
		}

		err = PutTorrent(metaTorrent)

		return 0, err
	}
}

// 入库主函数
func Put() {
	// 设置最大允许使用CPU核心数
	runtime.GOMAXPROCS(2)
	// 定义一个通道
	chReq := make(chan models.SC_Hash, 10)
	// 定义一个通道
	chRes := make(chan string, 10)

	// 无限循环入库种子
	for {
		// 获取未入库hash总量
		allcount := models.Count(models.DbHash, bson.M{"isput": false, "invalid": bson.M{"$lte": 3}})

		// 如果数量小于1
		if allcount < 1 {
			// 停顿10秒
			time.Sleep(10 * time.Second)
			// 跳过本次循环
			continue
		}

		for ai := 0; ai < allcount; {
			// 定义一个SC_Hash列表
			var sc_hash []models.SC_Hash
			// 获取100条infohash
			models.GetDataByQuery(models.DbHash, ai, (ai + 100), "-hot", bson.M{"isput": false, "invalid": bson.M{"$lte": 3}}, &sc_hash)
			// 获取到的总量
			count := len(sc_hash)

			// 循环处理infohash, 每次处理进程为自定义进程数
			for i := 0; i < 10; i++ {
				go func() {
					// 循环处理
					for {
						// 循环接受通道传递过来的infohash
						schash := <-chReq

						// 检查infohash是否已经入库
						if models.Has(models.DbInfo, bson.M{"infohash": strings.ToUpper(schash.InfoHash)}) {
							// 入库则输出跳过信息
							chRes <- fmt.Sprintf("'%s' Skip......", schash.InfoHash)
							// 将hash设置为已入库
							models.SetPut(schash.InfoHash)
							// 跳过本次循环
							continue
						}

						// 定义一个string变量, 记录处理结果
						var str string

						// 入库种子信息
						ret, err := PullTorrent(schash.InfoHash)
						if err == nil && ret == 0 {
							// 设置成功信息
							str = fmt.Sprintf("Storage InfoHash '%s' Success......", schash.InfoHash)
						} else {
							// 设置失败信息
							str = fmt.Sprintf("Can not download '%s' torrent file......", schash.InfoHash)
						}
						// 传递处理结果
						chRes <- str
					}
				}()
			}

			go func() {
				// 定义一个hashs列表
				hashs := make([]models.SC_Hash, count)

				// 循环将infohash加入hashs列表
				for i := 0; i < count; i++ {
					hashs[i] = sc_hash[i]
				}

				// 传递infohash给通道
				for i := 0; i < count; i++ {
					chReq <- hashs[i]
				}
			}()

			// 循环处理入库结果
			for i := 0; i < count; i++ {
				// 接收入库结果
				str := <-chRes

				// 如果允许显示则显示
				if models.DbConfig.ShowMsg && str != "" {
					fmt.Println(str)
				} else {
					// 否则不显示
					_ = str
				}
			}
		}
	}
}
