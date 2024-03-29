package controller

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sues-go/model"
	"time"
)

var netTransport *http.Transport

func init() {
	var err error
	proxyAddr := "http://118.25.210.52:8080"
	proxy, err := url.Parse(proxyAddr)
	if err != nil {
		log.Fatal(err)
	}
	netTransport = &http.Transport{
		Proxy:                 http.ProxyURL(proxy),
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: time.Second * time.Duration(5),
	}
}

// 接口
func GetSUESCourses(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	if len(username) == 0 {
		c.JSON(400, gin.H{"detail": "用户名缺失"})
		return
	}
	if len(password) == 0 {
		c.JSON(400, gin.H{"detail": "密码缺失"})
	}
	var err error
	// 获取验证码和cookie
Start:
	captcha, cookie, err := getCaptchaAndCookie()
	if err != nil {
		c.JSON(400, gin.H{"detail": err.Error()})
		return
	}
	// 登录
	err = loginJxxt(username, password, captcha, cookie)
	if err != nil {
		if err.Error() == "验证码错误" {
			goto Start
		} else {
			c.JSON(400, gin.H{"detail": err.Error()})
			return
		}
	}
	// 获取stdID
	stdID, err := getStdID(cookie)
	if err != nil {
		c.JSON(400, gin.H{"detail": err.Error()})
		return
	}
	// 获取课表
	courses, err := kgetCourses(cookie, stdID)
	if err != nil {
		c.JSON(400, gin.H{"detail": err.Error()})
		return
	}
	// 缓存账号
	go saveAccount("SUES", username, password)
	c.JSON(200, courses)
}

// 验证码和Cookie
func getCaptchaAndCookie() (captcha string, cookie string, err error) {
	fmt.Println("验证码获取")
	i := 0
	for {
		if i >= 5 {
			break
		}
		// 获取验证码
		req, _ := http.NewRequest("GET", "http://jxxt.sues.edu.cn/eams/captcha/image.action", strings.NewReader(""))
		client := &http.Client{
			Transport: netTransport,
		}
		resp, _ := client.Do(req)
		body, _ := ioutil.ReadAll(resp.Body)
		// 生成图片名
		imgName := getRandomString(10)
		out, _ := os.Create(imgName)
		// 删除临时文件
		defer os.Remove(imgName)
		io.Copy(out, bytes.NewReader(body))
		// 获取cookie
		cookies := resp.Header["Set-Cookie"]
		JSESSIONID := strings.Split(cookies[0], ";")[0]
		test := strings.Split(cookies[1], ";")[0]
		cookie = JSESSIONID + ";popped='';" + test
		fmt.Println("cookie")
		fmt.Println(cookie)
		// 识别验证码
		cmd := exec.Command("/bin/bash", "-c", "tesseract "+imgName+" "+imgName+" -l eng")
		cmd.Run()
		captchaTxt, _ := ioutil.ReadFile(imgName + ".txt")
		// 删除临时文件
		defer os.Remove(imgName + ".txt")
		captcha = strings.Split(string(captchaTxt), "\n")[0]
		captcha = strings.Replace(captcha, " ", "", -1)
		isValid, _ := regexp.MatchString("^[a-z]{4,5}$", captcha)
		if !isValid {
			fmt.Println("imgName:", imgName, "captcha:", captcha)
		} else {
			return
		}
		i++
	}
	return
}

// 登录教学管理系统
func loginJxxt(username, password, captcha, cookie string) (err error) {
	postValue := url.Values{
		"loginForm.name":     {username},
		"loginForm.password": {password},
		"loginForm.captcha":  {string(captcha)},
		"encodedPassword":    {""},
	}
	postString := postValue.Encode()
	req, _ := http.NewRequest("POST", "http://jxxt.sues.edu.cn/eams/login.action", strings.NewReader(postString))
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{
		Transport: netTransport,
	}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	if strings.Contains(string(body), "Wrong Captcha String") {
		return errors.New("验证码错误")
	} else if strings.Contains(string(body), "Error Password") {
		return errors.New("密码错误")
	}
	defer resp.Body.Close()
	return nil
}

// 获取学生ID
func getStdID(cookie string) (stdID string, err error) {
	req, _ := http.NewRequest("GET", "http://jxxt.sues.edu.cn/eams/courseTableForStd.action?method=stdHome", strings.NewReader(""))
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", "http://jxxt.sues.edu.cn/eams/defaultHome.action?method=moduleList&parentCode=")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{
		Transport: netTransport,
	}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	temp := strings.Split(string(body), "javascript:getCourseTable('std','")[1]
	stdID = strings.Split(temp, "',event)")[0]
	return
}

// 获取课表数据
func kgetCourses(cookie, stdID string) (courses []model.Course, err error) {
	fmt.Println("获取课表数据")
	courses = make([]model.Course, 0)
	req, _ := http.NewRequest("GET", "http://jxxt.sues.edu.cn/eams/courseTableForStd.action?method=courseTable&setting.forSemester=1&setting.kind=std&semester.id=441&ids="+stdID+"&ignoreHead=1", strings.NewReader(""))
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{
		Transport: netTransport,
	}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	temp := strings.Split(string(body), "var activity=null;")[1]
	temp = strings.Split(temp, "table0.marshalTable")[0]
	courseStrs := strings.Split(temp, "activity = new TaskActivity(")
	for i, class := range courseStrs {
		if i == 0 {
			continue
		}
		lines := strings.Split(class, "\n")
		var course model.Course
		for _, line := range lines {
			if len(line) > 80 {
				// 课程
				courseStrArr := strings.Split(line, "\"")
				course.Teacher = courseStrArr[3]
				course.Name = courseStrArr[7]
				course.Address = courseStrArr[11]
				course.Week = courseStrArr[13]
			} else if len(line) < 30 && len(line) > 10 {
				// 星期和节数
				course.Index, _ = strconv.Atoi(string(line[8]))
				if course.Week != "" {
					course.Time = course.Time + ","
				}
				// index =2*unitCount+7;
				course.Time = course.Time + strings.Split(strings.Split(line, "+")[1], ";\r")[0]
			}
		}
		course.Time = course.Time[1:]
		courses = append(courses, course)
	}
	return
}

// 获得随机字符串
func getRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
