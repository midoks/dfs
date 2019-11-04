package main

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	"github.com/midoks/godfs/common"
	"github.com/midoks/godfs/config"
	"mime/multipart"
	// "net/http"
	"os"
	// "path"
	// "path/filepath"
	"strings"
	"sync/atomic"
	// "time"
	"unsafe"
)

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

type Server struct {
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
	fmt.Println("init end")
}

func Config() *config.GloablConfig {
	return (*config.GloablConfig)(atomic.LoadPointer(&ptr))
}

func (this *Server) Upload(c *gin.Context) {
	var (
		// folder    string
		// outPath string
		file *multipart.FileHeader
		// fname     string
		md5sum    string
		queryPath string
		output    string
	)

	md5sum = c.Param("md5")
	queryPath = c.Param("path")
	file, _ = c.FormFile("file")
	output = c.PostForm("output")

	fmt.Println(output)
	fmt.Println(md5sum)
	fmt.Println(queryPath)
	fmt.Println(file)
	// _, fname = filepath.Split(file.Filename)
	// if Config().RenameFile {
	// 	fname = common.MD5UUID() + path.Ext(fname)
	// }

	// folder = time.Now().Format("20060102/15/04")
	// folder = fmt.Sprintf(STORE_DIR+"/%s", folder)

	// if f, _ := common.FileExists(folder); !f {
	// 	os.MkdirAll(folder, 0775)
	// }
	// outPath = fmt.Sprintf(folder+"/%s", fname)

	// c.SaveUploadedFile(file, outPath)

	// c.JSON(http.StatusOK, gin.H{
	// 	"src": outPath,
	// })
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

	fmt.Println(fullpath)

	// c.JSON(http.StatusOK, gin.H{
	// 	"src": c.Request.RequestURI,
	// })

	c.File("./" + fullpath)
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
			uploadBigUrl = fmt.Sprintf("%s%s", uploadUrl, CONST_BIG_UPLOAD_PATH_SUFFIX)
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
		router.POST(fmt.Sprintf("%s/*path", groupRoute), this.Download)
	}

	router.GET("/upload.html", this.Index)
	router.POST("/upload", this.Upload)
	router.Any("/upload/*path", this.Upload)

	fmt.Println("Listen Port on", Config().Addr)
	router.Run(Config().Addr)
}

func main() {
	var s *Server
	s.Run()
}
