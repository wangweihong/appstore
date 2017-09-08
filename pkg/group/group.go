package group

import (
	"appstore/pkg/env"
	"appstore/pkg/log"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"ufleet-deploy/deploy/kv"
)

const (
	etcdGroupExternalKey = "/ufleet/group"
	EventDelete          = "delete"
	EventCreate          = "create"
)

var (
	externalgroupNoticers = make(map[string]chan ExternalGroupEvent)
	externalgroupLock     = sync.Mutex{}
)

type ExternalGroupEvent struct {
	Action string
	Group  string
}

func GetGroupList(addr string) (map[string]string, error) {
	ekv := kv.NewKewStore(addr)
	if ekv == nil {
		return nil, fmt.Errorf("start etcd client failed")
	}
	groups := make(map[string]string, 0)

	resp, err := ekv.GetNode(etcdGroupExternalKey)
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return groups, nil
		}
		return nil, err
	}
	for _, v := range resp.Node.Nodes {
		group := filepath.Base(v.Key)
		groups[group] = group
	}
	return groups, nil
}

func watchGroupChange(addr string) error {
	ekv := kv.NewKewStore(addr)
	wechan, err := ekv.WatchNode(etcdGroupExternalKey)
	if err != nil {
		return err
	}
	go func() {
		for {
			we := <-wechan
			if we.Err != nil {
				log.ErrorPrint("externalgroupWatcher watch error: %v", we.Err)
				continue
			}

			res := we.Resp
			if res.Node.Key == etcdGroupExternalKey {
				continue
			}

			var group string
			s := strings.Split(strings.TrimPrefix(res.Node.Key, etcdGroupExternalKey+"/"), "/")
			if len(s) != 1 {
				continue
			}

			group = s[0]

			var ge ExternalGroupEvent
			switch res.Action {
			case "delete": //忽略根Key的事件
				ge.Group = group
				ge.Action = EventDelete

			case "set":
				ge.Group = group
				ge.Action = EventCreate

			default:
				continue
			}
			log.DebugPrint("catch group event :", ge)
			for _, v := range externalgroupNoticers {
				go func(ch chan ExternalGroupEvent) {
					ch <- ge
				}(v)
			}
		}
	}()

	return nil
}
func RegisterExternalGroupNoticer(kind string) (chan ExternalGroupEvent, error) {
	externalgroupLock.Lock()
	defer externalgroupLock.Unlock()

	if _, ok := externalgroupNoticers[kind]; ok {
		return nil, fmt.Errorf("externalgroup Noticer \"%v\" has registered", kind)
	}
	ch := make(chan ExternalGroupEvent)
	externalgroupNoticers[kind] = ch
	return ch, nil

}

func Init() {
	err := watchGroupChange(env.EtcdHost)
	if err != nil {
		panic(err.Error())
	}

}
