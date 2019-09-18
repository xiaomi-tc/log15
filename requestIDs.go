package log15

import (
	"context"
	"github.com/petermattis/goid"
	"sync"
)

type storeMeta struct  {
	reqID interface{}
	reqContext context.Context
}

var (
	requestIDs = map[int64]storeMeta{}
	rwm        sync.RWMutex
)

// Set 保存一个 RequestID, context
func SetReqMetaForGoroutine(ctx context.Context, ID interface{}) {
	goID := getGoID()
	rwm.Lock()
	defer rwm.Unlock()

	requestIDs[goID] = storeMeta {
		reqID:ID,
		reqContext:ctx,
	}
}

// Get 返回设置的 ReqMeta
func getReqMetaForGoroutine() (interface{},bool) {
	goID := getGoID()
	rwm.RLock()
	defer rwm.RUnlock()

	id,ok := requestIDs[goID]
	return id,ok
	//return requestIDs[goID]
}

// Get 返回设置的 RequestID
func GetReqIDForGoroutine() (interface{},bool) {
	meta,ok := getReqMetaForGoroutine()
	if ok {
		return meta.(storeMeta).reqID, ok
	}
	return nil,ok
}

// Get 返回设置的 ReqContext
func GetReqContextForGoroutine() (context.Context,bool) {
	meta,ok := getReqMetaForGoroutine()
	if ok {
		return meta.(storeMeta).reqContext, ok
	}
	return nil,ok
}

// Delete 删除设置的 RequestID
func DeleteMetaForGoroutine() {
	goID := getGoID()
	rwm.Lock()
	defer rwm.Unlock()

	delete(requestIDs, goID)
}

func getGoID() int64 {
	return goid.Get()
}
