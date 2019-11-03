package common

import (
	"crypto/md5"
	// "crypto/sha1"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	random "math/rand"
	"net"
	"os"
	"strings"
	"time"
)

func FileExists(fileName string) (bool, error) {
	_, err := os.Stat(fileName)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func RandInt(min, max int) int {
	return func(min, max int) int {
		r := random.New(random.NewSource(time.Now().UnixNano()))
		if min >= max {
			return max
		}
		return r.Intn(max-min) + min
	}(min, max)
}

func GetPulicIP() string {
	var (
		err  error
		conn net.Conn
	)
	if conn, err = net.Dial("udp", "8.8.8.8:80"); err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return localAddr[0:idx]
}

func ReadBinFile(path string) ([]byte, error) {
	if IsExist(path) {
		fi, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer fi.Close()
		return ioutil.ReadAll(fi)
	} else {
		return nil, errors.New("not found")
	}
}

func WriteFile(path string, data string) bool {
	if err := ioutil.WriteFile(path, []byte(data), 0775); err == nil {
		return true
	} else {
		return false
	}
}

func MD5(str string) string {
	md := md5.New()
	md.Write([]byte(str))
	return fmt.Sprintf("%x", md.Sum(nil))
}

func GetUUID() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	id := MD5(base64.URLEncoding.EncodeToString(b))
	return fmt.Sprintf("%s-%s-%s-%s-%s", id[0:8], id[8:12], id[12:16], id[16:20], id[20:])
}

func MD5UUID() string {
	return MD5(GetUUID())
}
