package store

import (
	"appstore/pkg/log"

	"github.com/rjeczalik/notify"
)

type FileType int

const (
	RepoDir   FileType = 1
	RepoFile  FileType = 2
	Cache     FileType = 3
	CacheFile FileType = 4
)

type Event struct {
	Type  FileType //
	Event notify.Event
	Path  string
}

func WatchDir(home string) error {
	c := make(chan notify.EventInfo, 1)

	// Set up a watchpoint listening for events within a directory tree rooted
	// at current working directory. Dispatch remove events to c.

	//必须是绝对路径
	log.DebugPrint("watching %v", home)
	if err := notify.Watch(home+"/...", c, notify.All); err != nil {
		return log.ErrorPrint(err)
	}
	//defer notify.Stop(c)

	// Block until an event is received.
	go func() {
		for {
			ei := <-c
			processEvent(home, ei)
		}
	}()
	return nil
}

//能够递归watchM
func processEvent(home string, ei notify.EventInfo) {
	//检测文件是
	//	if event.IsRename()

	e := genereteEvent(ei.Path(), home, ei.Event())
	if e == nil {
		//		log.DebugPrint("path(%v) ignore", path)
		return
	}

	log.DebugPrint("<---watch file event: %v, type: %v, path:%v", e.Event, e.Type, e.Path)
	//

	//修改repositoryFile文件
	//更新repos的信息
	switch e.Type {
	case RepoDir:
	case RepoFile:
		if e.Event == notify.Write || e.Event == notify.Create {
			//读取文件更新内容
			go doRepoChange(home)
		}

		if e.Event == notify.Remove {

		}
	case Cache:
	case CacheFile:
		//??
	}

}

func genereteEvent(path string, home string, event notify.Event) *Event {

	/*
		if path == home.RepositoryFile() {
			return &Event{
				Type:  RepoFile,
				Event: event,
				Path:  path,
			}
		}

		if path == home.Cache() {
			return &Event{
				Type:  Cache,
				Event: event,
				Path:  path,
			}
		}
	*/

	//cacheReg := home.CacheIndex("\\w")
	//log.DebugPrint(cacheReg)

	/*
		if strings.HasPrefix(path, home.Cache()) {
			r, err := regexp.Compile(home.CacheIndex(cacheReg))
			if err != nil {
				//			log.ErrorPrint(err)
				return nil
			}
			if r.MatchString(path) == true {
				return &Event{
					Type:  CacheFile,
					Event: event,
					Path:  path,
				}
			}
		}

	*/
	return nil
}

func doRepoChange(home string) {
	Locker.Lock()
	defer Locker.Unlock()
	//需要加锁
	log.DebugPrint("catching helm repo file change, flush to memory")
	err := InitHelmManager(home)
	if err != nil {
		log.ErrorPrint(err)
	}
	log.DebugPrint(helm)
}
