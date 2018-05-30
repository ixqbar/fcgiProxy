package proxy

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"math/rand"
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
	Time    int64
	IsSys   bool
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
		obj.Time = time.Now().Unix()
		obj.IsSys = false
		return nil
	}

	if strings.Index(content, "socks://") >= 0 {
		obj.Type = TProxyIsSocks
		obj.Address = strings.Replace(content, "socks://", "", -1)
		obj.Time = time.Now().Unix()
		obj.IsSys = true
		return nil
	}

	if strings.Index(content, "http://") >= 0 || strings.Index(content, "https://") >= 0 {
		obj.Type = TProxyIsHttp
		obj.Address = content
		obj.Time = time.Now().Unix()
		obj.IsSys = true
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
	sync.Mutex
	ProxyConfigIndex    int
	AdminServerAddress  string          `xml:"admin_server"`
	HttpServerAddress   string          `xml:"http_server"`
	HttpServerSSLCert   string          `xml:"http_ssl_cert"`
	HttpServerSSLKey    string          `xml:"http_ssl_key"`
	HttpStaticRoot      string          `xml:"http_static_root"`
	HttpRc4EncryptKey   string          `xml:"http_rc4_key"`
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

func (obj *FConfig) ClearEmptyProxy() {
	obj.Lock()
	defer obj.Unlock()

	obj.ProxyConfigIndex = -1

	var tp []TProxyConfig
	var currentTime = time.Now().Unix()
	for _, v := range obj.ProxyList {
		if v.IsSys == false && (v.Type == TProxyIsNone || v.Time+86400 <= currentTime) {
			continue
		}
		tp = append(tp, v)
	}

	obj.ProxyList = tp
}

func (obj *FConfig) AddProxyConfig(category int, address string, port string) {
	obj.Lock()
	defer obj.Unlock()

	obj.ProxyList = append(obj.ProxyList, TProxyConfig{category, fmt.Sprintf("%s:%s", address, port), time.Now().Unix(), false})
}

func (obj *FConfig) GetOneProxyConfig(index int) (TProxyConfig, int) {
	obj.Lock()
	defer obj.Unlock()

	if len(obj.ProxyList) == 0 {
		return TProxyConfig{TProxyIsNone, "", 0, false}, -1
	}

	proxyIndex := 0
	if index >= 0 && index < len(obj.ProxyList) {
		proxyIndex = index
	} else {
		proxyIndex = rand.Intn(len(obj.ProxyList))
	}

	return obj.ProxyList[proxyIndex], proxyIndex
}

func (obj *FConfig) RemoveProxyConfig(index int) {
	obj.Lock()
	defer obj.Unlock()

	obj.ProxyConfigIndex = -1

	if index < 0 || index >= len(obj.ProxyList) {
		return
	}

	Logger.Printf("remove unavailable proxy config index %d, %s", index, obj.ProxyList[index].String())

	obj.ProxyList = append(obj.ProxyList[:index], obj.ProxyList[index+1:]...)
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

	Config = &FConfig{
		ProxyConfigIndex: -1,
	}

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
