package httpvf

import (
	"fmt"
	"github.com/toukii/goutils"
	yaml "gopkg.in/yaml.v2"
)

type Resp struct {
	Code int
	Cost int
	Body string
}

const (
	GET  = "GET"
	POST = "POST"
)

type Req struct {
	URL    string
	Method string
	Body   string
	Resp   Resp
	Filename string
}

func Reqs(filename string) (reqs []*Req, err error) {
	in := goutils.ReadFile(filename)
	reqs = make([]*Req, 0, 1)
	if len(in) > 0 && in[0] != byte('-') {
		var req1 *Req
		req1, err = req(in)
		if goutils.CheckErr(err) {
			return nil, err
		}
		reqs = append(reqs, req1)
		return
	}
	err = yaml.Unmarshal(in, &reqs)
	if goutils.CheckErr(err) {
		return nil, err
	}
	return reqs, nil
}

func req(in []byte) (*Req, error) {
	var req Req
	err := yaml.Unmarshal(in, &req)
	if goutils.CheckErr(err) {
		return nil, err
	}

	return &req, nil
}

func Test() {
	var req Req
	req.URL = "http://upload.daoapp.io/upload/a.json"
	req.Body = fmt.Sprintf(`{"name":"toukii"}`)
	req.Resp.Body = "world"
	req.Method = GET
	reqs := []Req{req, req}
	bs2, err := yaml.Marshal(reqs)
	goutils.CheckErr(err)
	fmt.Println(goutils.ToString(bs2))

	reqs2 := []Req{}
	err = yaml.Unmarshal(bs2, &reqs2)
	goutils.CheckErr(err)
	fmt.Println(reqs2)

}
