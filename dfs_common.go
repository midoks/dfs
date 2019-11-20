package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/midoks/godfs/database"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strings"
)

func dPrint(args ...interface{}) {
	if Config().Debug {
		fmt.Println("[", Config().Host, "]:[start]")
		for i := 0; i < len(args); i++ {
			fmt.Print(args[i])
		}
		fmt.Println("\n[end]")
	}
}

func sendToMail(to, subject, body, mailtype string) error {
	host := Config().Mail.Host
	user := Config().Mail.User
	password := Config().Mail.Password
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var contentType string
	if mailtype == "html" {
		contentType = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain" + "; charset=UTF-8"
	}
	msg := []byte("To: " + to + "\r\nFrom: " + user + ">\r\nSubject: " + "\r\n" + contentType + "\r\n\r\n" + body)
	sendTo := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, sendTo, msg)
	return err
}

func getOtherPeers() []string {

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

func checkFileExists(post_url, md5 string) bool {

	resp, _ := http.PostForm(post_url, url.Values{"md5": {md5}})
	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	m := ReturnJsonData{}

	err := json.Unmarshal([]byte(string(respBody)), &m)
	if err == nil {
		if m.Code == 0 {
			return true
		}
	}
	return false
}

func asyncFileUpload(postUrl, groupMd5 string, info *database.BinFile) bool {

	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	filePath := fmt.Sprintf(STORE_DIR+"/%s", info.Path)
	fileWriter, _ := bodyWriter.CreateFormFile("file", filePath)

	file, _ := os.Open(filePath)
	defer file.Close()
	io.Copy(fileWriter, file)

	bodyWriter.WriteField("path", info.Path)
	bodyWriter.WriteField("md5", info.Md5)
	bodyWriter.WriteField("group_md5", groupMd5)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, _ := http.Post(postUrl, contentType, bodyBuffer)
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	m := ReturnJsonData{}
	err := json.Unmarshal([]byte(string(respBody)), &m)
	if err == nil {
		if m.Code == 0 {
			return true
		}
	}
	return false
}

func asyncFileInfo(postUrl string, info *database.BinFile) bool {

	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	bodyWriter.WriteField("node", Config().Host)
	bodyWriter.WriteField("md5", info.Md5)
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, _ := http.Post(postUrl, contentType, bodyBuffer)
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	m := ReturnJsonData{}
	err := json.Unmarshal([]byte(string(respBody)), &m)
	if err == nil {
		if m.Code == 0 {
			return true
		}
	}
	return false
}

func asyncSearch(postUrl string, md5 string, format string) (ReturnJsonData, error) {
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	bodyWriter.WriteField("format", Config().Host)
	bodyWriter.WriteField("md5", md5)
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, _ := http.Post(postUrl, contentType, bodyBuffer)
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	m := ReturnJsonData{}
	err := json.Unmarshal([]byte(string(respBody)), &m)
	return m, err
}
