package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type UrlEndcodeReq struct {
	GrantType string `schema:"grant_type,omitempty" json:"grant_type"`
}

func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func RespondArray(w http.ResponseWriter, data []map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func CallExternalAPI(r *http.Request) (err error, isObject bool, data map[string]interface{}, dataArray []map[string]interface{}) {
	var client = &http.Client{}
	var req = &http.Request{}

	if r.FormValue("contentType") == "multipart/form-data" {
		_ = r.ParseMultipartForm(0)
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		var fileName = ""

		param := map[string]io.Reader{}

		for key := range r.MultipartForm.Value {
			param[key] = strings.NewReader(r.Form.Get(key))
		}

		for key := range r.MultipartForm.File {
			f, handler, err := r.FormFile(key)
			if err != nil {
				print(err.Error())

			}
			fileName = handler.Filename
			param[key] = f
		}

		for key, r := range param {
			var fw io.Writer
			if x, ok := r.(io.Closer); ok {
				defer x.Close()
			}

			if _, ok := r.(multipart.File); ok {
				if fw, err = w.CreateFormFile(key, fileName); err != nil {
					return
				}
			} else {
				if fw, err = w.CreateFormField(key); err != nil {
					return
				}
			}
			if _, err = io.Copy(fw, r); err != nil {
				return err, false, nil, nil
			}

		}

		_ = w.Close()

		req, err = http.NewRequest(r.FormValue("method"), r.FormValue("url"), &b)
		if err != nil {
			print(err.Error())
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

		for name, values := range r.Header {
			for _, value := range values {
				if name != "Content-Type" {
					req.Header.Set(name, value)
				}
			}
		}

		user, pass, _ := r.BasicAuth()
		if len(user) > 0 && len(pass) > 0 {
			req.SetBasicAuth(user, pass)
		}

	} else {
		request := make(map[string]interface{})
		reqBody, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			print(err.Error())
		} else {
			err = json.Unmarshal(reqBody, &request)
			if err != nil {
				print(err.Error())
			}
		}

		if fmt.Sprint(request["content_type"]) == "application/x-www-form-urlencoded" {
			var reqUrlEncode UrlEndcodeReq
			form := url.Values{}
			value, _ := json.Marshal(request["body"])
			_ = json.Unmarshal(value, &reqUrlEncode)

			encoder := schema.NewEncoder()
			err = encoder.Encode(&reqUrlEncode, form)
			if err != nil {
				return err, true, map[string]interface{}{"message": err.Error()}, dataArray
			} else {
				req, _ = http.NewRequest(fmt.Sprint(request["method"]), fmt.Sprint(request["url"]), strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		} else {
			value, _ := json.Marshal(request["body"])
			req, _ = http.NewRequest(fmt.Sprint(request["method"]), fmt.Sprint(request["url"]), bytes.NewBuffer(value))
		}

		if request["headers"] != nil {
			for k, v := range request {
				if k == "headers" {
					for _, v := range v.([]interface{}) {
						req.Header[fmt.Sprintf("%v", v.(map[string]interface{})["key"])] = []string{fmt.Sprintf("%v", v.(map[string]interface{})["value"])}
					}
				}
			}
		}

		if request["basic_auth"] != nil {
			if len(fmt.Sprint(request["basic_auth"].(map[string]interface{})["username"])) > 0 && len(fmt.Sprint(request["basic_auth"].(map[string]interface{})["password"])) > 0 {
				req.SetBasicAuth(fmt.Sprint(request["basic_auth"].(map[string]interface{})["username"]), fmt.Sprint(request["basic_auth"].(map[string]interface{})["password"]))
			}
		}

	}

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: TlsSkipVerify(),
		},
	}

	resp, errReq := client.Do(req)
	if errReq != nil {
		print(errReq.Error())
		return
	} else {
		defer resp.Body.Close()
	}
	body, errBody := ioutil.ReadAll(resp.Body)
	if errBody != nil {
		print(errBody)
		return
	}

	x := bytes.TrimLeft(body, " \t\r\n")
	if len(x) > 0 && x[0] == '[' {
		err = json.Unmarshal(body, &dataArray)
		if err != nil {
			fmt.Println(err)
		}
		return err, false, data, dataArray
	} else {
		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println(err)
		}
		return err, true, data, dataArray
	}

}

func GetDateNow() time.Time {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	tm := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(),
		time.Now().Minute(), time.Now().Second(), time.Now().Nanosecond(), time.UTC)
	return tm.In(loc)
}

func Base64ToFile(base64str string) []byte {
	dec, err := base64.StdEncoding.DecodeString(base64str)
	if err != nil {
		// handle error
	}
	return dec
}

func TlsSkipVerify() *tls.Config {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	return config
}
