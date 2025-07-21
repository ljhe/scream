package pbgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type Codec interface {
	Marshal(msg interface{}) (interface{}, error) // todo...上下文Context
	Unmarshal(data interface{}, msg interface{}) error
	TypeOfName() string
}

type MessageInfo struct {
	Codec Codec
	Type  reflect.Type
	ID    uint16
}

var (
	messageByID   = map[uint16]*MessageInfo{}
	messageByType = map[reflect.Type]*MessageInfo{}
	messageByName = map[string]*MessageInfo{}
)

var registerCodec Codec // 后续有别的解析部分这边可以添加

func init() {
	// 注册proto解析
	RegisterCodec(new(pbCodec))
}

func RegisterCodec(c Codec) {
	registerCodec = c
}

func GetCodec() Codec {
	return registerCodec
}

// pbCodec
type pbCodec struct {
}

func (c *pbCodec) TypeOfName() string {
	return "protobuf"
}
func (c *pbCodec) Marshal(msg interface{}) (interface{}, error) {
	return proto.Marshal(msg.(proto.Message))
}
func (c *pbCodec) Unmarshal(data interface{}, msg interface{}) error {
	return proto.Unmarshal(data.([]byte), msg.(proto.Message))
}

// http json
type httpJsonCodec struct {
}

func (c *httpJsonCodec) TypeOfName() string {
	return "httpjson"
}
func (c *httpJsonCodec) MimeType() string {
	return "application/json"
}
func (c *httpJsonCodec) Marshal(msg interface{}) (interface{}, error) {
	httpData, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	//log.Printf("httpData:%v", httpData)

	return bytes.NewReader(httpData), nil
}
func (c *httpJsonCodec) Unmarshal(data interface{}, msg interface{}) error {
	var reader io.Reader
	switch v := data.(type) {
	case *http.Request:
		reader = v.Body
	case io.Reader:
		reader = v
	}
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	log.Println("httpJsonCodec:", string(body))
	return json.Unmarshal(body, msg)
}

// httpForm
type httpFormCodec struct {
}

func (this *httpFormCodec) TypeOfName() string {
	return "httpform"
}
func (this *httpFormCodec) MimeType() string {
	return "application/x-www-form-urlencoded"
}
func (this *httpFormCodec) Marshal(msg interface{}) (interface{}, error) {
	return strings.NewReader(this.form2UrlValues(msg).Encode()), nil
}
func (this *httpFormCodec) Unmarshal(data interface{}, msg interface{}) error {
	//todo...
	if msg != nil {
		body, err := ioutil.ReadAll(data.(io.Reader))
		if err != nil {
			return err
		}
		//log.Println("body11:", string(body))

		msgValue := reflect.ValueOf(msg)
		if msgValue.Kind() == reflect.Ptr {
			msgValue = msgValue.Elem()
		}
		msgValue.Field(0).SetString(string(body))
	}
	return nil
}
func (this *httpFormCodec) form2UrlValues(obj interface{}) url.Values {
	objValue := reflect.Indirect(reflect.ValueOf(obj))
	objType := reflect.TypeOf(obj)

	formValues := url.Values{}
	for i := 0; i < objValue.NumField(); i++ {
		field := objType.Field(i)
		val := objValue.Field(i)
		//if field {
		formValues.Add(field.Name, this.value2String(val.Interface()))
		//}
	}
	return formValues
}
func (this *httpFormCodec) value2String(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(int64(v), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		panic("Unknown type to convert to string")
	}
}

func RegisterMessageInfo(info *MessageInfo) {
	// 注册时统一为非指针类型
	if info.Type.Kind() == reflect.Ptr {
		info.Type = info.Type.Elem()
	}

	if info.ID == 0 {
		panic(fmt.Sprintf("router ID invalid:%v", info.Type.Name()))
	}

	if _, ok := messageByID[info.ID]; ok {
		panic(fmt.Sprintf("router ID:%v already registered", info.ID))
	} else {
		messageByID[info.ID] = info
	}

	if _, ok := messageByType[info.Type]; ok {
		panic(fmt.Sprintf("router Type:%v already registered", info.Type))
	} else {
		messageByType[info.Type] = info
	}

	if _, ok := messageByName[info.Type.Name()]; ok {
		panic(fmt.Sprintf("router Name:%v already registered", info.Type))
	} else {
		messageByName[info.Type.Name()] = info
	}
}

func MessageInfoById(msgId uint16) *MessageInfo {
	return messageByID[msgId]
}

func MessageInfoByMsg(msg interface{}) *MessageInfo {
	msgType := reflect.TypeOf(msg)
	if msgType.Kind() == reflect.Ptr {
		return messageByType[msgType.Elem()]
	} else {
		return messageByType[msgType]
	}
}

func MessageInfoByName(name string) *MessageInfo {
	return messageByName[name]
}

func MessageToString(msg interface{}) string {
	if msg == nil {
		return ""
	}
	if str, ok := msg.(interface{ String() string }); ok {
		return str.String()
	}
	return ""
}
