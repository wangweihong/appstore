package store

import (
	"appstore/pkg/env"
	"appstore/pkg/fl"
	"appstore/pkg/group"
	"appstore/pkg/log"
	"appstore/pkg/watcher"
	"os"
	"strings"
)

const (
	GroupEventKind = "store"
)

func handleGroupEvent(ch <-chan group.ExternalGroupEvent) {
	for {
		ge := <-ch

		switch ge.Action {
		case group.EventDelete:
			if strings.TrimSpace(ge.Group) == "" {
				log.DebugPrint("event handler catch invalid group %v", ge.Group)
				fl.ReleaseLock()
				continue
			}

			err := fl.WatchAndWaitLock()
			if err != nil {
				log.ErrorPrint("err")
				continue
			}
			path := env.StoreHome + "/" + ge.Group
			err = os.RemoveAll(path)
			if err != nil && !os.IsNotExist(err) {
				fl.ReleaseLock()
				log.ErrorPrint(err)
				continue
			}
			fl.ReleaseLock()

		case group.EventCreate:

			if strings.TrimSpace(ge.Group) == "" {
				log.DebugPrint("event handler catch invalid group %v", ge.Group)
				continue
			}

			err := fl.WatchAndWaitLock()
			if err != nil {
				log.ErrorPrint("err")
				continue
			}

			path := env.StoreHome + "/" + ge.Group
			err = os.MkdirAll(path, 0755)
			if err != nil {
				log.DebugPrint("event handler catch invalid group %v", ge.Group)
				fl.ReleaseLock()
				continue
			}

			err = env.InitHelmEnv(path)
			if err != nil {
				log.DebugPrint("init helm home %v fail: %v", path, err)
				fl.ReleaseLock()
				continue

			}
			fl.ReleaseLock()
		}

	}
}

func handleStoreEvent(ch chan watcher.Event) {
	for {
		e := <-ch
		log.DebugPrint("recieve event...", e)
		switch e.Type {
		//需要加载文件
		case watcher.RepoFile:
			if e.Event == watcher.Write {

				err := fl.WatchAndWaitLock()
				if err != nil {
					log.DebugPrint(err)
					continue
				}

				repoGroup, err := loadGroupRepo(e.Path)
				if err != nil {
					log.DebugPrint(err)
					fl.ReleaseLock()
					continue
				}
				Locker.Lock()
				hm.RepoGroups[e.Group] = *repoGroup
				Locker.Unlock()

			}
			//组删除,不需要初始化helm arch,已经被初始化过了
			//但还是需要读文件以获取内容
		case watcher.GroupFile:
			if e.Event == watcher.Create {

				err := fl.WatchAndWaitLock()
				if err != nil {
					log.DebugPrint(err)
					continue
				}

				repoGroup, err := loadGroupRepo(e.Path)
				if err != nil {
					log.DebugPrint(err)
					fl.ReleaseLock()
					continue
				}

				Locker.Lock()
				hm.RepoGroups[e.Group] = *repoGroup

				Locker.Unlock()
			}
			if e.Event == watcher.Remove {
				Locker.Lock()
				delete(hm.RepoGroups, e.Group)
				Locker.Unlock()

			}
			//更新Local repo的数据
			//chart不通过内存缓存
		case watcher.LocalIndexFile:
			if e.Event == watcher.Write {

			}
		}
	}
}
