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

type MysqlConfig struct {
	Ip       string `xml:"ip"`
	Username string `xml:"username"`
	Password string `xml:"password"`
	Port     uint   `xml:"port"`
	Database string `xml:"database"`
}

type FConfig struct {
	AdminServerAddress  string        `xml:"admin_server"`
	HttpServerAddress   string        `xml:"http_server"`
	HttpServerSSLCert   string        `xml:"http_ssl_cert"`
	HttpServerSSLKey    string        `xml:"http_ssl_key"`
	HttpStaticRoot      string        `xml:"http_static_root"`
	FcgiServerAddress   string        `xml:"fcgi_server"`
	ScriptFileName      string        `xml:"script_filename"`
	QueryString         string        `xml:"query_string"`
	HeaderParams        []ProxyParams `xml:"header_params>param"`
	Origins             OrignList     `xml:"origins"`
	LoggerMysqlConfig   MysqlConfig   `xml:"logger>mysql"`
	LoggerRc4EncryptKey string        `xml:"logger>rc4_encrypt_key"`
}

var Config *FConfig

func ParseXmlConfig(path string) (*FConfig, error) {
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

	Config = &FConfig{}

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
