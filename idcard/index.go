package idcard

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wbylovesun/xutils/xslice"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var FormatErr = errors.New("idcard format error")

type adm struct {
	Provinces map[string]string `json:"provinces"`
	Cities    map[string]string `json:"cities"`
	Counties  map[string]string `json:"countries"`
	mu        sync.Mutex
}

var (
	adCodeOnce sync.Once
	adCodeVal  *adm
)

type UserInfo struct {
	idCard   string
	province string
	city     string
	country  string
	sex      int
	age      int
}

func NewUserInfo(idCard string) (*UserInfo, error) {
	if _, err := regexp.MatchString(IdCardPattern, idCard); err != nil {
		return nil, FormatErr
	}
	adCode()
	userInfo := &UserInfo{
		idCard: idCard,
	}
	if err := userInfo.setProvince(); err != nil {
		logrus.Error(err)
	}
	if err := userInfo.setCity(); err != nil {
		logrus.Error(err)
	}
	if err := userInfo.setCountry(); err != nil {
		logrus.Error(err)
	}
	if err := userInfo.setSex(); err != nil {
		logrus.Error(err)
	}
	if err := userInfo.setAge(); err != nil {
		logrus.Error(err)
	}
	return userInfo, nil
}

func (ui *UserInfo) GetProvince() string {
	return ui.province
}

func (ui *UserInfo) GetCity() string {
	return ui.city
}

func (ui *UserInfo) GetCountry() string {
	return ui.country
}

func (ui *UserInfo) GetSex() int {
	return ui.sex
}

func (ui *UserInfo) GetAge() int {
	return ui.age
}

func (ui *UserInfo) GetIdCard() string {
	return ui.idCard
}

func (ui *UserInfo) setProvince() error {
	province, ok := adCodeVal.Provinces[ui.idCard[0:2]]
	if !ok {
		return fmt.Errorf("省份不存在:%s", ui.idCard)
	}
	ui.province = province
	return nil
}

func (ui *UserInfo) setCity() error {
	city, ok := adCodeVal.Cities[ui.idCard[0:4]]
	if ok {
		ui.city = city
	} else {
		// 城市编码前四位查询不到，以六位开始查询（直辖市）
		if city, ok := adCodeVal.Cities[ui.idCard[0:6]]; ok {
			ui.city = city
		} else {
			return fmt.Errorf("城市不存在:%s", ui.idCard)
		}
	}
	return nil
}

func (ui *UserInfo) setCountry() error {
	var centralCity = []string{"北京市", "天津市", "上海市", "重庆市"}
	country, ok := adCodeVal.Counties[ui.idCard[0:6]]
	if !ok {
		// 判断是否是直辖市
		if !xslice.Contains(centralCity, ui.GetProvince()) {
			return fmt.Errorf("县区不存在:%s", ui.idCard)
		}
	}
	ui.country = country
	return nil
}

func (ui *UserInfo) setAge() error {
	var birthday string
	if len(ui.idCard) == 18 {
		birthday = ui.idCard[6:14]
	} else {
		birthday = "19" + ui.idCard[6:12]
	}
	_, err := time.ParseInLocation("20060102", birthday, time.Local)
	if err != nil {
		return fmt.Errorf("birthday invalid:%v", err)
	}
	birthdayYear, err := strconv.Atoi(birthday[0:4])
	if err != nil {
		return fmt.Errorf("身份证年获取失败:%v", err)
	}
	age := time.Now().Year() - birthdayYear
	if strings.Compare(
		fmt.Sprintf("%d%d", time.Now().Month(), time.Now().Day()),
		ui.idCard[10:14],
	) == 1 {
		age -= 1
	}
	ui.age = age
	return nil
}

func (ui *UserInfo) setSex() error {
	var (
		sexInt int
		err    error
	)

	if len(ui.idCard) == 18 {
		sexInt, err = strconv.Atoi(ui.idCard[16:17])
	} else {
		sexInt, err = strconv.Atoi(ui.idCard[14:])
	}
	if err != nil {
		return fmt.Errorf("身份证性别获取失败:%v", err)
	}
	var sex int
	if sexInt%2 == 0 {
		sex = SexWomen
	}
	ui.sex = sex
	return nil
}

func adCode() *adm {
	adCodeOnce.Do(func() {
		path, err := filepath.Abs("./")
		if err != nil {
			logrus.Errorf(fmt.Sprintf("获取绝对路径失败：%v", err))
			return
		}
		var s adm
		s.mu.Lock()
		defer s.mu.Unlock()
		bs, err := os.ReadFile(path + "/location.json")
		if err != nil {
			logrus.WithField("path", path).Errorf(fmt.Sprintf("读取文件失败：%v", err))
			return
		}
		err = json.Unmarshal(bs, &s)
		if err != nil {
			logrus.Errorf(fmt.Sprintf("json解析失败：%v", err))
			return
		}
		adCodeVal = &s
	})
	return adCodeVal
}

func handleAdCodeData() error {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	bytes, err := os.ReadFile("./database.json")
	if err != nil {
		return errors.Wrap(err, "读取database.json文件失败")
	}
	var location map[string][]string
	err = json.Unmarshal(bytes, &location)
	if err != nil {
		return errors.Wrap(err, "解析文件失败")
	}

	province, city, country := make(map[string]string), make(map[string]string), make(map[string]string)
	//处理数据
	for key, value := range location {
		if ok, _ := regexp.MatchString(`^\d$`, key); ok {
			// 省份
			for _, val := range value {
				valSpan := strings.Split(val, "|")
				province[valSpan[1][0:2]] = valSpan[0]
			}
		} else if ok, _ := regexp.MatchString(`^\d+_\d+$`, key); ok {
			// 市区
			for _, val := range value {
				valSpan := strings.Split(val, "|")
				if ok, _ := regexp.MatchString(`^\d{4}00$`, valSpan[1]); ok {
					city[valSpan[1][0:4]] = valSpan[0]
				} else {
					city[valSpan[1]] = valSpan[0]
				}
			}
		} else if ok, _ := regexp.MatchString(`^\d+_\d+_\d+`, key); ok {
			for _, val := range value {
				valSpan := strings.Split(val, "|")
				country[valSpan[1]] = valSpan[0]
			}
		}
	}

	locationBytes, err := json.Marshal(map[string]map[string]string{"province": province, "city": city, "country": country})
	if err != nil {
		return errors.Wrap(err, "json转换数据失败")
	}
	return os.WriteFile("./location.json", locationBytes, 0755)
}
