package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Error struct {
	URL   string
	Index string
}

func (s Error) Error() string {
	return fmt.Sprintf("error index %v,Url %v", s.Index, s.URL)
}
func syncRouter(s *sync.Map) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.13:2379"},
	})
	if err != nil {
		log.Panic(err)
	}
	log.Println("etcd connect success")
	defer cli.Close()
	ctx := context.TODO()
	prefix := "/service"
	response, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		log.Panic(err)
	}
	for i := range response.Kvs {
		key := string(response.Kvs[i].Key)[len(prefix):]
		s.Store(key, string(response.Kvs[i].Value))
	}
	serviceChan := cli.Watch(ctx, prefix, clientv3.WithPrefix())
	for wresp := range serviceChan {
		for _, ev := range wresp.Events {
			var key string
			key = string(ev.Kv.Key[len(prefix):])
			if ev.Type == mvccpb.PUT {
				s.Store(key, string(ev.Kv.Value))
			} else {
				s.Delete(key)
			}
		}
	}
}
func parseRouter(router *sync.Map, request *http.Request) (*url.URL, error) {
	var prefix string
	router.Range(func(i, value interface{}) bool {
		if strings.Index(request.URL.Path, i.(string)) == 0 {
			if len(prefix) < len(i.(string)) {
				prefix = i.(string)
			}
		}
		return true
	})
	throwErr := func() Error {
		return Error{
			URL:   request.RequestURI,
			Index: prefix,
		}
	}
	if prefix == "" {
		return nil, throwErr()
	}
	value, ok := router.Load(prefix)
	if ok != true {
		return nil, throwErr()
	}
	uri, err := url.Parse(value.(string))
	if err != nil {
		return nil, throwErr()
	}
	// 透明化前缀
	request.URL.Path = request.URL.Path[len(prefix):]
	return uri, nil
}
