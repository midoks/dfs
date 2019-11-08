package main

import (
	// "database/sql"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	// _ "github.com/mattn/go-sqlite3"
	"encoding/json"
	"github.com/midoks/godfs/common"
	"github.com/midoks/godfs/config"
	"github.com/midoks/godfs/database"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

var server *Server

const (
	STORE_DIR_NAME               = "files"
	LOG_DIR_NAME                 = "log"
	DATA_DIR_NAME                = "data"
	CONF_DIR_NAME                = "conf"
	STATIC_DIR_NAME              = "static"
	CONST_BIG_UPLOAD_PATH_SUFFIX = "/big/upload/"
)

var (
	ptr                  unsafe.Pointer
	DOCKER_DIR           = ""
	STORE_DIR            = STORE_DIR_NAME
	CONF_DIR             = CONF_DIR_NAME
	LOG_DIR              = LOG_DIR_NAME
	DATA_DIR             = DATA_DIR_NAME
	STATIC_DIR           = STATIC_DIR_NAME
	CONST_CONF_FILE_NAME = CONF_DIR + "/cfg.json"
)

type QueueUploadChan struct {
	c       *gin.Context
	tmpPath string
	done    chan bool
}

type Server struct {
	queueUpload chan QueueUploadChan
	db          *database.DB
}

func NewServer() *Server {
	var srv = &Server{
		queueUpload: make(chan QueueUploadChan, 100),
	}
	return srv
}

func Config() *config.GloablConfig {
	return (*config.GloablConfig)(atomic.LoadPointer(&ptr))
}

func GetOtherPeers() []string {

	npeers := []string{}
	peers := Config().Peers
	host := Config().Host
	for i := 0; i < len(peers); i++ {
		if host == peers[i] {
			continue
		}
		npeers = append(npeers, peers[i])
	}
	return npeers
}

func init() {
	fmt.Println("init start")

	DOCKER_DIR = os.Getenv("GODFS_DIR")
	if DOCKER_DIR != "" {
		if !strings.HasSuffix(DOCKER_DIR, "/") {
			DOCKER_DIR = DOCKER_DIR + "/"
		}
	}
	STORE_DIR = DOCKER_DIR + STORE_DIR_NAME
	CONF_DIR = DOCKER_DIR + CONF_DIR_NAME
	DATA_DIR = DOCKER_DIR + DATA_DIR_NAME
	LOG_DIR = DOCKER_DIR + LOG_DIR_NAME
	STATIC_DIR = DOCKER_DIR + STATIC_DIR_NAME
	folders := []string{DATA_DIR, STORE_DIR, CONF_DIR, STATIC_DIR}
	for _, folder := range folders {
		os.MkdirAll(folder, 0775)
	}

	peerId := fmt.Sprintf("%d", common.RandInt(0, 9))
	if f, _ := common.FileExists(CONST_CONF_FILE_NAME); !f {
		var ip string
		if ip = os.Getenv("GODFS_IP"); ip == "" {
			ip = common.GetPulicIP()
		}
		peer := "http://" + ip + ":8080"
		cfg := fmt.Sprintf(config.CONFIG_JSON, peerId, peer, peer)
		common.WriteFile(CONST_CONF_FILE_NAME, cfg)
	}

	ptr = config.Parse(CONST_CONF_FILE_NAME)

	server = NewServer()
	server.initComponent()
	server.initDb()
	fmt.Println("init end")
}

func (this *Server) initComponent() {

	if Config().ReadTimeout == 0 {
		Config().ReadTimeout = 60 * 10
	}
	if Config().WriteTimeout == 0 {
		Config().WriteTimeout = 60 * 10
	}
	if Config().SyncWorker == 0 {
		Config().SyncWorker = 200
	}
	if Config().UploadWorker == 0 {
		Config().UploadWorker = runtime.NumCPU() + 4
		if runtime.NumCPU() < 4 {
			Config().UploadWorker = 8
		}
	}
	if Config().UploadQueueSize == 0 {
		Config().UploadQueueSize = 200
	}
	if Config().RetryCount == 0 {
		Config().RetryCount = 3
	}
}

func (this *Server) initDb() {
	this.db = database.Open("data/dfs.db")
}

func (this *Server) initUploadTask() {
	uploadFunc := func() {
		for {

			task := <-this.queueUpload
			this.uploadChan(task.c, task.tmpPath)
			// this.upload(*wr.w, wr.r)
			// this.rtMap.AddCountInt64(CONST_UPLOAD_COUNTER_KEY, wr.r.ContentLength)
			// if v, ok := this.rtMap.GetValue(CONST_UPLOAD_COUNTER_KEY); ok {
			// 	if v.(int64) > 1*1024*1024*1024 {
			// 		var _v int64
			// 		this.rtMap.Put(CONST_UPLOAD_COUNTER_KEY, _v)
			// 		debug.FreeOSMemory()
			// 	}
			// }
			task.done <- true
		}
	}
	for i := 0; i < Config().UploadWorker; i++ {
		go uploadFunc()
	}
}

func (this *Server) retOk(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"msg":  "ok",
		"code": 0,
		"data": data,
	})
}

func (this *Server) retFail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": -1,
		"msg":  msg,
	})
}

func (this *Server) uploadChan(c *gin.Context, tmpFilePath string) {
	var (
		err     error
		fname   string
		file    *multipart.FileHeader
		folder  string
		outPath string
		fileMd5 string
	)

	folder = time.Now().Format("20060102/15/04")

	scene := c.PostForm("scene")
	if scene != "" {
		folder = fmt.Sprintf(STORE_DIR+"/%s/%s", scene, folder)
	} else {
		folder = fmt.Sprintf(STORE_DIR+"/%s", folder)
	}

	file, err = c.FormFile("file")

	if err != nil {
		this.retFail(c, "upload request fail!")
	}

	_, fname = filepath.Split(file.Filename)
	if Config().RenameFile {
		fname = common.MD5UUID() + path.Ext(fname)
	}

	if f, _ := common.FileExists(folder); !f {
		os.MkdirAll(folder, 0777)
	}
	outPath = fmt.Sprintf(folder+"/%s", fname)

	tmpFile, _ := os.Open(tmpFilePath)
	defer tmpFile.Close()

	if Config().EnableDistinctFile {
		fileMd5 = common.GetFileSum(tmpFile, Config().FileSumArithmetic)
	} else {
		fileMd5 = common.MD5(outPath)
	}

	findData, _ := this.db.FindFileByMd5(fileMd5)

	if findData.Md5 == fileMd5 {
		outPath = findData.Path
	} else {
		err = c.SaveUploadedFile(file, outPath)
		if err != nil {
			this.retFail(c, "upload fail!")
			return
		}
		outPath = strings.Replace(outPath, STORE_DIR+"/", "", 1)

		node_data := [...]string{Config().Host}
		node, _ := json.Marshal(node_data)
		fmt.Println(string(node))
		err = this.db.AddFileRow(fileMd5, outPath, 1, string(node), "attr")
		fmt.Println(err)
		go this.AyncUpload(fileMd5)

	}
	GetOtherPeers()
	data := make(map[string]interface{})
	data["size"] = file.Size
	data["src"] = outPath
	data["scene"] = scene
	data["md5"] = fileMd5
	data["group"] = Config().Group

	this.retOk(c, data)
}

func (this *Server) AyncUpload(md5 string) {
	fmt.Println("AyncUpload:", md5)

	peers := GetOtherPeers()
	for i := 0; i < len(peers); i++ {
		fmt.Println(peers[i])
	}

}

func (this *Server) Upload(c *gin.Context) {
	var (
		file   *multipart.FileHeader
		folder string
	)

	file, _ = c.FormFile("file")
	folder = time.Now().Format("20060102")
	folder = fmt.Sprintf(STORE_DIR+"/_tmp/%s", folder)

	if f, _ := common.FileExists(folder); !f {
		os.MkdirAll(folder, 0777)
	}

	outFile := fmt.Sprintf(folder+"/%s", common.GetUUID())
	defer func() {
		os.Remove(outFile)
	}()
	c.SaveUploadedFile(file, outFile)

	done := make(chan bool, 1)
	this.queueUpload <- QueueUploadChan{c, outFile, done}
	<-done
}

func (this *Server) Download(c *gin.Context) {

	if c.Request.RequestURI == "/" ||
		c.Request.RequestURI == "" ||
		c.Request.RequestURI == "/"+Config().Group ||
		c.Request.RequestURI == "/"+Config().Group+"/" {
		this.Index(c)
		return
	}

	fullpath := c.Param("path")
	c.File("files/" + fullpath)
}

func (this *Server) Delete(c *gin.Context) {
	md5 := c.PostForm("md5")
	data, find := this.db.FindFileByMd5(md5)
	if find {
		os.Remove(data.Path)
		this.db.DeleteRowById(data.Id)
		this.retOk(c, "file deleted successfully!")
	}
	this.retFail(c, "file does not exist!")
}

func (this *Server) Search(c *gin.Context) {
	md5 := c.PostForm("md5")
	data, find := this.db.FindFileByMd5(md5)
	if find {
		r := make(map[string]interface{})
		r["group"] = Config().Group
		r["path"] = data.Path
		this.retOk(c, r)
	}
	this.retFail(c, "file does not exist!")

}

func (this *Server) CheckFileExists(c *gin.Context) {
	md5 := c.PostForm("md5")
	_, find := this.db.FindFileByMd5(md5)
	if find {
		this.retOk(c, "ok")
		return
	}
	this.retFail(c, "not find!")
}

func (this *Server) SyncFile(c *gin.Context) {

}

func (this *Server) Status(c *gin.Context) {

	data := make(map[string]interface{})
	data["peers"] = Config().Peers

	this.retOk(c, data)
}

func (this *Server) Index(c *gin.Context) {
	var (
		uploadUrl    string
		uploadBigUrl string
		uppy         string
	)

	uploadUrl = "/upload"
	uploadBigUrl = CONST_BIG_UPLOAD_PATH_SUFFIX

	if Config().EnableWebUpload {
		if Config().SupportGroupManage {
			uploadBigUrl = "/file"
		}
		uppy = config.UPLOAD_TPL
		uppyFileName := STATIC_DIR + "/uppy.html"
		if common.IsExist(uppyFileName) {
			if data, err := common.ReadBinFile(uppyFileName); err != nil {
				log.Error(err)
			} else {
				uppy = string(data)
			}
		} else {
			common.WriteFile(uppyFileName, uppy)
		}
		fmt.Fprintf(c.Writer, fmt.Sprintf(uppy, uploadUrl, Config().DefaultScene, uploadBigUrl))
	} else {
		c.Writer.Write([]byte("web upload deny"))
	}
}

func (this *Server) Run() {

	go this.initUploadTask()

	router := gin.Default()

	groupRoute := ""
	if Config().SupportGroupManage {
		groupRoute = "/" + Config().Group
	}

	if groupRoute == "" {
		router.GET(fmt.Sprintf("%s", "/"), this.Download)
	} else {

		router.GET(fmt.Sprintf("%s", "/"), this.Download)
		router.GET(fmt.Sprintf("%s", groupRoute), this.Download)
		router.GET(fmt.Sprintf("%s/*path", groupRoute), this.Download)
	}

	router.GET("/upload.html", this.Index)
	router.GET("/status", this.Status)

	router.POST("/upload", this.Upload)
	router.POST("/delete", this.Delete)
	router.POST("/serach", this.Search)
	router.POST("/check_file_exists", this.CheckFileExists)
	router.POST("/sync_files", this.SyncFile)

	fmt.Println("Listen Port on", Config().Addr)
	router.Run(Config().Addr)
}

func main() {
	server.Run()
}
