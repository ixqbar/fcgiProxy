package proxy

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	TProxyIsNone  = iota
	TProxyIsSocks
	TProxyIsHttp
)

type TProxyParams struct {
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

type TMysqlConfig struct {
	Ip       string `xml:"ip"`
	Username string `xml:"username"`
	Password string `xml:"password"`
	Port     uint   `xml:"port"`
	Database string `xml:"database"`
}

type TProxyConfig struct {
	Type    int
	Address string
}

func (obj *TProxyConfig) String() string {
	switch obj.Type {
	case TProxyIsSocks:
		return fmt.Sprintf("sock5://%s", obj.Address)
	case TProxyIsHttp:
		return fmt.Sprintf("http://%s", obj.Address)
	case TProxyIsNone:
		return fmt.Sprint("None")
	}

	return ""
}

func (obj *TProxyConfig) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}

	if len(content) == 0 {
		obj.Type = TProxyIsNone
		obj.Address = ""
		return nil
	}

	if strings.Index(content, "socks://") >= 0 {
		obj.Type = TProxyIsSocks
		obj.Address = strings.Replace(content, "socks://", "", -1)
		return nil
	}

	if strings.Index(content, "http://") >= 0 || strings.Index(content, "https://") >= 0 {
		obj.Type = TProxyIsHttp
		obj.Address = content
		return nil
	}

	return errors.New(fmt.Sprintf("error proxy server %s", content))
}

type TQpushDevice struct {
	Group     string
	Name      string
	Code      string
	Sign      string
	Available bool
}

func (obj *TQpushDevice) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}

	if strings.Index(content, ",") >= 0 {
		deviceInfo := strings.Split(content, ",")
		obj.Group = deviceInfo[0]
		obj.Name = deviceInfo[1]
		obj.Code = deviceInfo[2]
		obj.Sign = ""
		obj.Available = false
		return nil
	}

	return errors.New(fmt.Sprintf("error qpush device %s", content))
}

func (obj *TQpushDevice) String() string {
	return fmt.Sprintf("group:%s,name:%s,code:%s,sign:%s", obj.Group, obj.Name, obj.Code, obj.Sign)
}

type FConfig struct {
	AdminServerAddress  string          `xml:"admin_server"`
	HttpServerAddress   string          `xml:"http_server"`
	HttpServerSSLCert   string          `xml:"http_ssl_cert"`
	HttpServerSSLKey    string          `xml:"http_ssl_key"`
	HttpStaticRoot      string          `xml:"http_static_root"`
	FcgiServerAddress   string          `xml:"fcgi_server"`
	ScriptFileName      string          `xml:"script_filename"`
	QueryString         string          `xml:"query_string"`
	HeaderParams        []TProxyParams  `xml:"header_params>param"`
	Origins             OrignList       `xml:"origins"`
	LoggerMysqlConfig   TMysqlConfig    `xml:"logger>mysql"`
	LoggerRc4EncryptKey string          `xml:"logger>rc4_encrypt_key"`
	ProxyList           []TProxyConfig  `xml:"proxy>server"`
	QpushDevices        []*TQpushDevice `xml:"qpush>device"`
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
