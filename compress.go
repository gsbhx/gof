package gof

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io/ioutil"
)

// @Author WangKan
// @Description //解压消息
// @Date 2021/3/3 10:09
// @Param
// @return
func DeCompress(content []byte, c *Conn) ([]byte, error) {
	content= append(content, 0x00,0x00,0xff,0xff,0x01,0x00,0x00,0xff,0xff)
	fr := flate.NewReader(bytes.NewReader(content))
	content, err := ioutil.ReadAll(fr)
	fmt.Printf("frw:%+v,err:%+v\n", content, err)
	return content, nil
}

// @Author WangKan
// @Description //压缩消息
// @Date 2021/3/3 10:10
// @Param
// @return
func Compress(content []byte, compressLevel int) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	fw, err := flate.NewWriter(buf, compressLevel)
	if err != nil {
		return nil, err
	}
	_, _ = fw.Write(content)
	_ = fw.Flush()
	_ = fw.Close()
	return []byte(buf.String()), nil
}
