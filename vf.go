package httpvf

import (
	//"github.com/astaxie/beego/httplib"
	"github.com/toukii/goutils"

	"fmt"
	"time"
	//"io/ioutil"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	//"net/http"
	"path/filepath"
	//"net/url"
	"io/ioutil"
	"strings"
	"sync"
	"regexp"
)

func verify(req *Req) *Msg {
	msg := newMsg(req)
	var resp *http.Response
	var request *http.Request
	var err error

	if len(req.Filename) > 0 {
		request, err = newfileUploadRequest(req.URL, nil, "filename", req.Filename)
		if goutils.CheckErr(err) {
			msg.Append(FATAL, err.Error())
			buf:=bytes.NewBufferString(req.Body)
			request,err = http.NewRequest(req.Method, req.URL, buf)
		}
	}else{
		buf:=bytes.NewBufferString(req.Body)
		request,err = http.NewRequest(req.Method, req.URL, buf)
	}
	if goutils.CheckErr(err) {
		msg.Append(FATAL, err.Error())
	}

	for k,v:=range req.Header{
		request.Header.Add(k,v)
	}

	//fmt.Printf("Request starting: %s\n",req.URL)
	//  start
	start := time.Now()


	c := http.Client{}
	resp, _ = c.Do(request)


	// end
	duration := time.Now().Sub(start)
	//fmt.Println("End.")

	// fmt.Printf("Request[%s] cost:%d ms\n", req.URL, duration.Nanoseconds()/1e6)
	cost := int(duration.Nanoseconds() / 1e6)
	if cost > req.Resp.Cost {
		msg.Append(ERROR, fmt.Sprintf("time cost: %d ms more than %d ms;", cost, req.Resp.Cost))
	} else if cost > req.Resp.Cost*3/4 {
		msg.Append(WARN, fmt.Sprintf("time cost: %d ms near by %d ms;", cost, req.Resp.Cost))
	} else {
		msg.Append(INFO, fmt.Sprintf("time cost: %d ms / %d ms;", cost, req.Resp.Cost))
	}
	msg.Req.Resp.RealCost = cost
	if resp == nil {
		msg.Append(ERROR, "nil response")
	} else {
		if req.Resp.Code != resp.StatusCode {
			msg.Append(ERROR, fmt.Sprintf("error code::%d gotten, %d wanted", resp.StatusCode, req.Resp.Code))
		}
		bs,respReadErr:=ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if len(req.Resp.Body)>0{
			if(goutils.CheckErr(respReadErr)){
				msg.Append(ERROR, respReadErr.Error())
			}
			if !strings.EqualFold(req.Resp.Body,goutils.ToString(bs)){
				msg.Append(ERROR, fmt.Sprintf("response body is: %s, not wanted: %s\n",goutils.ToString(bs),req.Resp.Body))
			}
		}

		if len(req.Resp.ReBody)>0 {
			if matched,errg:=regexp.Match(req.Resp.ReBody, bs);!matched || goutils.LogCheckErr(errg){
				msg.Append(ERROR, fmt.Sprintf("response body is: %s, not wanted regexp: %s\n",goutils.ToString(bs),req.Resp.ReBody))
			}
		}
	}
	return msg
}


func Verify(vf string) {
	reqs, _ := Reqs(vf)
	var wg sync.WaitGroup
	for _, it := range reqs {
		wg.Add(1)
		go func(it *Req){
			i:=0
			cost := 0
			var tps string
			logs := make([]*Log,0,64)
			for {
				msg := verify(it)
				//fmt.Println(msg)
				cost += msg.Req.Resp.RealCost
				i++
				logs = append(logs, msg.Logs()...)
				if i>= it.N {
					tps += fmt.Sprint("avg cost: ",cost/i," ms")
					if cost == 0 {
						tps += fmt.Sprint("TPS: +INF")
					}else{
						tps += fmt.Sprint(", TPS: ",1000.0*float32(i)/float32(cost))
					}
					msg = newMsg(it)
					msg.Append(INFO, tps)
					msg.AppendLogs(logs)
					fmt.Println(msg)
					break
				}
			}
			wg.Done()
		}(it)
	}
	wg.Wait()
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	return req, err
}
