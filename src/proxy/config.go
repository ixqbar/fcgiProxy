package proxy

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"
)

type ProxyParams struct {
	Key   string `xml:"key"`
	Value string `xml:"value"`
}

type OrignList []string

func (l *OrignList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}

	*l = strings.Split(strings.ToLower(content), ",")
	return nil
}

func (l *OrignList) ToString() string {
	return strings.Join(*l, ",")
}

type ProxyConfig struct {
	AdminServerAddress string        `xml:"admin_server"`
	HttpServerAddress  string        `xml:"http_server"`
	FcgiServerAddress  string        `xml:"fcgi_server"`
	ScriptFileName     string        `xml:"script_filename"`
	QueryString        string        `xml:"query_string"`
	HeaderParams       []ProxyParams `xml:"header_params>param"`
	Origins            OrignList     `xml:"origins"`
}

var Config *ProxyConfig

func ParseXmlConfig(path string) (*ProxyConfig, error) {
	if len(path) == 0 {
		return nil, errors.New("not found configure xml file")
	}

	n, err := GetFileSize(path)
	if err != nil || n == 0 {
		return nil, errors.New("not found configure xml file")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	Config = &ProxyConfig{}

	data := make([]byte, n)

	m, err := f.Read(data)
	if err != nil {
		return nil, err
	}

	if int64(m) != n {
		return nil, errors.New(fmt.Sprintf("expect read configure xml file size %d but result is %d", n, m))
	}

	err = xml.Unmarshal(data, &Config)
	if err != nil {
		return nil, err
	}

	Logger.Printf("read config %+v", Config)

	return Config, nil
}
