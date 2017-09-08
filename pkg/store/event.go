package store

import (
	"appstore/pkg/env"
	"appstore/pkg/group"
	"appstore/pkg/log"
	"os"
	"strings"

	"k8s.io/helm/pkg/helm/helmpath"
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
				continue
			}

			Locker.Lock()
			path := env.StoreHome + "/" + ge.Group
			err := os.RemoveAll(path)
			if err != nil && !os.IsNotExist(err) {
				Locker.Unlock()
				log.ErrorPrint(err)
				continue
			}

			delete(hm.RepoGroups, ge.Group)
			Locker.Unlock()

		case group.EventCreate:

			if strings.TrimSpace(ge.Group) == "" {
				log.DebugPrint("event handler catch invalid group %v", ge.Group)
				continue
			}

			Locker.Lock()
			path := env.StoreHome + "/" + ge.Group
			err := os.MkdirAll(path, 0755)
			if err != nil {
				Locker.Unlock()
				log.DebugPrint("event handler catch invalid group %v", ge.Group)
				continue
			}

			err = env.InitHelmEnv(path)
			if err != nil {
				Locker.Unlock()
				log.DebugPrint("init helm home %v fail: %v", path, err)
				continue

			}

			var rg RepoGroup
			rg.Repos = make(map[string]Repo)
			rg.Home = helmpath.Home(path)

			hm.RepoGroups[ge.Group] = rg

			Locker.Unlock()
		}
	}
}
