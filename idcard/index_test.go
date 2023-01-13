package idcard

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"regexp"
	"testing"
)

func Test_handleAdCodeData(t *testing.T) {
	err := handleAdCodeData()
	assert.Nil(t, err)
}

func Test_NewUserInfo(t *testing.T) {
	fmt.Println(NewUserInfo("510232195508152414"))
	fmt.Println(NewUserInfo("460022197112162510"))
}

func TestNewUserInfo(t *testing.T) {
	//读取文件
	file, err := os.Open("./sz_containerlog_202301121029.txt")
	if err != nil {
		t.Fatal("打开文件失败")
	}
	defer func() { _ = file.Close() }()

	idCards := make([]string, 0)
	lineReader := bufio.NewReader(file)
	for {
		line, _, err := lineReader.ReadLine()
		if err == io.EOF {
			break
		}
		// 处理数据
		id := regexp.MustCompile(`\d{18}`).FindStringSubmatch(string(line))
		idCards = append(idCards, id...)
	}

	if len(idCards) == 0 {
		t.Fatal("数据结果为空")
	}
	for _, card := range idCards {
		userInfo, err := NewUserInfo(card)
		if err != nil {
			t.Error(err)
			continue
		}
		fmt.Println(userInfo.GetProvince(), userInfo.GetCity(), userInfo.GetCountry())
	}
}
