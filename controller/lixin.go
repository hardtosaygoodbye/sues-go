package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sues-go/model"
)

// 接口
func GetLIXINCourses(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	if len(username) == 0 {
		c.JSON(400, gin.H{"detail": "用户名缺失"})
		return
	}
	if len(password) == 0 {
		c.JSON(400, gin.H{"detail": "密码缺失"})
	}

	// 登录
	info := loginLixin(username, password)

	// 拿课表
	courses := getLixinCourses(info)

	// 缓存账号
	go saveAccount("LIXIN", username, password)

	c.JSON(200, courses)
}

func loginLixin(u, p string) lixinInfo {
	client := &http.Client{}
	url := "http://sso.lixin.edu.cn/authorize.php?client_id=ufsso_supwisdom_jw&response_type=code&redirect_uri=http%3A%2F%2Fnewjw.lixin.edu.cn%2Fsso%2Findex&state=1q2w3e"
	req, _ := http.NewRequest("POST", url, strings.NewReader("username="+u+"&password="+p))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	req.Header.Set("client_id", "ufsso_supwisdom_jw&redirect_uri=http%3A%2F%2Fnewjw.lixin.edu.cn%2Fsso%2Findex&state=1q2w3e&response_type=code")
	resp, _ := client.Do(req)
	cookies := resp.Header.Get("Set-Cookie")
	URP_SID := strings.Split(cookies, ";")[0]
	URP_SID = strings.Split(URP_SID, "=")[1]

	url = "http://newjw.lixin.edu.cn/webapp/std/edu/lesson/home.action"
	req, _ = http.NewRequest("GET", url, strings.NewReader(""))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	req.Header.Set("Cookie", "URP_EDU=%7B%22projectId%22%3A5%2C%22semesterId%22%3A1640420191%7D; JSESSIONID=18254633ADC75E28BE35A8CAF0455258; SERVERID=s6; URP_SID="+URP_SID)
	resp, _ = client.Do(req)
	cookies = resp.Header.Get("Set-Cookie")
	JSESSIONID := strings.Split(cookies, ";")[0]
	JSESSIONID = strings.Split(JSESSIONID, "=")[1]

	url = "http://newjw.lixin.edu.cn/webapp/std/edu/lesson/timetable!innerIndex.action?x-requested-with=1&projectId=5"
	req, _ = http.NewRequest("GET", url, strings.NewReader(""))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	req.Header.Set("Cookie", "URP_EDU=%7B%22projectId%22%3A5%2C%22semesterId%22%3A1640420191%7D; JSESSIONID="+JSESSIONID+"; SERVERID=s6; URP_SID="+URP_SID)
	resp, _ = client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	ids := strings.Split(string(body), "bg.form.addInput(form,\"ids\",\"")[1]
	ids = strings.Split(ids, "\");")[0]

	return lixinInfo{
		IDS:        ids,
		URP_SID:    URP_SID,
		JSESSIONID: JSESSIONID,
	}
}

func getLixinCourses(info lixinInfo) (courses []model.Course) {
	client := &http.Client{}
	url := "http://newjw.lixin.edu.cn/webapp/std/edu/lesson/timetable!courseTable.action"
	req, _ := http.NewRequest("POST", url, strings.NewReader("setting.kind=std&weekSpan=6&semester.id=1640420191&ids="+info.IDS))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	req.Header.Set("Cookie", "URP_EDU=%7B%22projectId%22%3A5%2C%22semesterId%22%3A1640420191%7D; JSESSIONID="+info.JSESSIONID+"; URP_SID="+info.URP_SID+"; SERVERID=s6")
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)

	courses = make([]model.Course, 0)
	temp := strings.Split(string(body), "var activity=null;")[1]
	temp = strings.Split(temp, "table0.marshalTable")[0]
	courseStrs := strings.Split(temp, "activity = table0.newActivity(")
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
				course.Week = dec2bin(strings.Split(courseStrArr[14], ")")[0][1:])
			} else {
				// 星期和节数
				arr := strings.Split(line, "addActivityByTime(")
				if len(arr) > 1 {
					arr = strings.Split(arr[1], ",")
					course.Index, _ = strconv.Atoi(arr[1])
					course.Time = arr[2] + "~" + strings.Split(arr[3], ")")[0]
				}
			}
		}
		courses = append(courses, course)
	}
	return
}

type lixinInfo struct {
	IDS        string
	URP_SID    string
	JSESSIONID string
}

func dec2bin(str string) string {
	n, _ := strconv.ParseInt(str, 10, 64)
	s := ""
	for q := n; q > 0; q = q / 2 {
		m := q % 2
		s = fmt.Sprintf("%v%v", m, s)
	}
	l := 54 - len(s)
	for i := 0; i < l; i++ {
		s = fmt.Sprintf("0%v", s)
	}
	return s
}
