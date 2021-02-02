package gof

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
	"syscall"
)


var keyGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

func computeAcceptKey(challengeKey string) string {
	h := sha1.New()
	h.Write([]byte(challengeKey))
	h.Write(keyGUID)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

//HandShaker
//
func GetHeaderContent(fd int)  (string,int) {
	for{
		var buf [1024]byte
		nbytes, _ := syscall.Read(fd, buf[:])
		if nbytes > 0 {
			return string(buf[:]),nbytes
		}
	}
}

func FormatHeader(buf string,length int)(map[string]string){
	var header =make(map[string]string)
	str:=""
	index:=0
	for i:=0;i<length;i++{
		if buf[i]==13{
			if index == 0 {
				arr:=strings.Split(str," ")
				fmt.Println(arr)
				header["Method"]=arr[0]
				str=""
				index++
				continue
			}
			if str != ""{
				arr:=strings.Split(str,": ")
				header[arr[0]]=arr[1]
				str=""
			}
		}else if buf[i]==10{
			continue
		}else{
			str += string(buf[i])
		}
	}
	return header
}

