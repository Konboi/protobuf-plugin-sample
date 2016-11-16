package sample

import (
	_ "io/ioutil"
	_ "log"
	_ "net/http"
)

type PingImplController interface {
	PingHandler(w http.ResponseWriter, r *http.Request)
}

func NewServer(c Ping) *http.ServeMux {
	server := http.NewServeMux()

	server.HandleFunc("/ping/ping", PingHandler)

	return server
}

func parsePingRequest(r *http.Request) (*service.PingRequest, error) {
	req := &service.PingRequest{}
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
