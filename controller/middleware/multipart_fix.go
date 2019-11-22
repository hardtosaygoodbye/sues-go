package middleware

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"mime/multipart"
)

func FixMultipart() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ContentType() == "multipart/form-data" {
			form, _ := c.MultipartForm()
			if len(form.Value) == 0 && len(form.File) > 0 {
				for k, v := range form.File {
					if len(v[0].Filename) == 0 {
						if len(v) > 0 {
							buf, err := getpart(v[0])
							if err != nil {
								continue
							}
							form.Value[k] = append(form.Value[k], buf)
						}
					}
				}
				for k, _ := range form.Value {
					delete(form.File, k)
				}
			}
			c.Request.MultipartForm = form
		}
		c.Next()
	}
}

func getpart(v *multipart.FileHeader) (string, error) {
	f, err := v.Open()
	if err != nil {
		println("Open fail")
		return "", err
	}
	buf, err := ioutil.ReadAll(f)
	defer f.Close()
	if err != nil {
		println("Read fail")
		return "", err
	}
	return string(buf), nil
}
