package sample

import (
	_ "io/ioutil"
	_ "log"
	_ "net/http"
)

type PlayerImplController interface {
	InfoHandler(w http.ResponseWriter, r *http.Request)
	EntryHandler(w http.ResponseWriter, r *http.Request)
	CommentHandler(w http.ResponseWriter, r *http.Request)
}

func NewServer(c Player) *http.ServeMux {
	server := http.NewServeMux()

	server.HandleFunc("/player/info", hoge.C(hoge.B(hoge.A(InfoHandler(fuga.A(fuga.B))))))

	server.HandleFunc("/player/entry", EntryHandler)

	server.HandleFunc("/player/comment", CommentHandler)

	return server
}

func parsePlayerInfoRequest(r *http.Request) (*service.PlayerInfoRequest, error) {
	req := &service.PlayerInfoRequest{}
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

func parsePlayerEntryRequest(r *http.Request) (*service.PlayerEntryRequest, error) {
	req := &service.PlayerEntryRequest{}
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

func parsePlayerCommentRequest(r *http.Request) (*service.PlayerCommentRequest, error) {
	req := &service.PlayerCommentRequest{}
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
