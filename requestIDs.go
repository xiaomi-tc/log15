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
	requestIDs sync.Map
)

// Set 保存一个 RequestID, context
func SetReqMetaForGoroutine(ctx context.Context, ID interface{}) {
	requestIDs.Store(getGoID(),storeMeta {
		reqID:ID,
		reqContext:ctx,
	})
}

// Get 返回设置的 ReqMeta
func getReqMetaForGoroutine() (interface{},bool) {
	return requestIDs.Load(getGoID())
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
	requestIDs.Delete(getGoID())
}

func getGoID() int64 {
	return goid.Get()
}
