package log15

import (
	"github.com/petermattis/goid"
	"sync"
)

var (
	requestIDs = map[int64]interface{}{}
	rwm        sync.RWMutex
)

// Set 保存一个 RequestID
func SetReqIDForGoroutine(ID interface{}) {
	goID := getGoID()
	rwm.Lock()
	defer rwm.Unlock()

	requestIDs[goID] = ID
}

// Get 返回设置的 RequestID
func GetReqIDForGoroutine() (interface{},bool) {
	goID := getGoID()
	rwm.RLock()
	defer rwm.RUnlock()

	id,ok := requestIDs[goID]
	return id,ok
	//return requestIDs[goID]
}

// Delete 删除设置的 RequestID
func DeleteReqIDForGoroutine() {
	goID := getGoID()
	rwm.Lock()
	defer rwm.Unlock()

	delete(requestIDs, goID)
}

func getGoID() int64 {
	return goid.Get()
}
