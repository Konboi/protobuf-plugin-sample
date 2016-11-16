package main

import (
	//	"bytes"
	//	"fmt"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/Konboi/protobuf-plugin-example/middleware"
	"github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	//"github.com/k0kubun/pp"
	"github.com/serenize/snaker"
)

type Proto struct {
	PackageName string
	Services    []Service
	Messages    []Message
}

type Service struct {
	Name    string
	Methods []Method
}

type Method struct {
	Name string
	//	Request  Message
	//	Response Message
	Request  string
	Response string
	Before   []string
	After    []string
}

type Message struct {
	Name   string
	Fields []Field
}

type Field struct {
	Type  string
	Name  string
	Label int
}

var (
	controllerTmpl *template.Template
)

func main() {
	log.SetFlags(log.Lshortfile)
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("read os.Stding error: %v", err)
	}

	req := plugin.CodeGeneratorRequest{}
	err = proto.Unmarshal(data, &req)
	if err != nil {
		log.Fatalf("load data proto.Unmarshal error: %v", err)
	}
	res := plugin.CodeGeneratorResponse{}

	err = pb2go(&req, &res)
	if err != nil {
		log.Fatalf("pb2go error: %v", err)
	}
}

func pb2go(req *plugin.CodeGeneratorRequest, res *plugin.CodeGeneratorResponse) error {
	res.File = make([]*plugin.CodeGeneratorResponse_File, 0, len(req.ProtoFile))
	for _, file := range req.GetProtoFile() {
		//log.Println("load", file.GetName())
		prt := Proto{}
		prt.Services = make([]Service, 0, len(file.GetService()))
		prt.Messages = make([]Message, 0, len(file.GetMessageType()))
		var fileName string
		for _, service := range file.GetService() {
			s := Service{}
			s.Name = service.GetName()

			s.Methods = make([]Method, 0, len(service.GetMethod()))
			fileName = strings.ToLower(service.GetName())

			for _, method := range service.GetMethod() {
				m := Method{}
				m.Name = method.GetName()
				m.Request = method.GetInputType()
				m.Response = method.GetOutputType()
				before, after, err := parseMiddlewareOption(method)
				if err != nil {
					log.Fatal(err)
				}
				m.Before = before
				m.After = after
				s.Methods = append(s.Methods, m)
			}
			prt.Services = append(prt.Services, s)
		}

		for _, message := range file.GetMessageType() {
			msg := Message{}
			msg.Name = message.GetName()
			msg.Fields = make([]Field, 0, len(message.GetField()))
			for _, field := range message.GetField() {
				msg.Fields = append(msg.Fields, Field{
					Type:  field.GetTypeName(),
					Name:  field.GetName(),
					Label: int(field.GetLabel()),
				})
			}
			prt.Messages = append(prt.Messages, msg)
		}

		if fileName != "" {
			file, err := os.OpenFile(fmt.Sprintf("tmp/%s.auto.go", fileName), os.O_RDWR, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			err = controllerTmpl.Execute(file, prt)
			if err != nil {
				log.Fatal(err)
			}
		}

	}
	return nil
}

func parseMiddlewareOption(method *descriptor.MethodDescriptorProto) (before, after []string, err error) {
	before, after = make([]string, 0), make([]string, 0)

	if !proto.HasExtension(method.Options, middleware.E_Middleware) {
		return before, after, nil
	}

	ext, err := proto.GetExtension(method.Options, middleware.E_Middleware)
	if err != nil {
		return before, after, err

	}

	opt, ok := ext.(*middleware.Middleware)
	if !ok {
		return before, after, fmt.Errorf("ext is %T", ext)
	}

	after = append(after, opt.After...)
	before = append(before, opt.Before...)

	return before, after, err
}

func (m Method) Path() string {
	snakerName := snaker.CamelToSnake(m.Name)
	path := strings.Replace(snakerName, "_", "/", -1)

	return path
}

func (m Method) Handler() string {
	handler := fmt.Sprintf("%sHandler", m.Name)
	beforeFilter := ""
	afterFilter := ""

	bracketNum := 0
	for _, filter := range m.Before {
		if filter == "" {
			continue
		}

		beforeFilter = fmt.Sprintf("%s(%s", filter, beforeFilter)
		bracketNum++
	}

	for _, filter := range m.After {
		if filter == "" {
			continue
		}

		afterFilter = fmt.Sprintf("%s(%s", afterFilter, filter)
		bracketNum++
	}

	handler = fmt.Sprintf("%s%s%s%s", beforeFilter, handler, afterFilter, strings.Repeat(")", bracketNum))
	log.Println(handler)
	return handler
}

func (s Service) Path() string {
	snakerName := snaker.CamelToSnake(s.Name)
	path := fmt.Sprintf("/%s", strings.Replace(snakerName, "_", "/", -1))

	return path
}

func (m Message) MethodName() string {
	return fmt.Sprintf("parse%s", m.Name)
}

func (m Message) IsRequest() bool {
	if strings.Contains(m.Name, "Request") {
		return true
	}

	return false
}

func init() {
	tmpl := `package sample

import (
	_ "io/ioutil"
	_ "log"
	_ "net/http"

)
{{ range .Services }}
type {{ .Name }}ImplController interface {
    {{ range .Methods }}
    {{ .Name }}Handler(w http.ResponseWriter, r *http.Request){{ end }}
}
{{ end }}

{{ range .Services }}
func NewServer(c {{ .Name }}) *http.ServeMux {
	server := http.NewServeMux()
    {{ $service := . }}
    {{ range .Methods }}
	server.HandleFunc("{{ $service.Path }}/{{ .Path }}", {{ .Handler }})
    {{ end }}

	return server
}
{{ end }}

{{ range .Messages }}
{{ if .IsRequest }}
func parse{{ .Name }}(r *http.Request) (*service.{{ .Name }}, error) {
	req := &service.{{ .Name }}{}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	err = proto.Unmarshal(buf, req)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer r.Body.Close()

	return req, nil
}
{{ end }}
{{ end }}
`
	controllerTmpl = template.Must(template.New("ctrl").Parse(tmpl))
}
