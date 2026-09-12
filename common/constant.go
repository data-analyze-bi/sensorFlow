package common

import "sync"

const (
	ContextKeyUserObj = "authedUserObj"
	REDIS_PREFIX_AUTH = "auth:"

	CacheEventsName = "cache:events:name"  //事件名称
	CacheEventsData = "cache:events:data:" //埋点数据缓存

	TrackSignup     = "track_signup" //注册
	TrackProfileSet = "profile_set"  //修改用户属性

	CacheUserInfo = "cache:user:info:" //用户属性
	CacheUsersLog = "cache:users:log"  //用户注册或者更新属性记录
)

type TableSchema struct {
	Columns map[string]bool
	Mutex   sync.RWMutex
}

type UsersSchema struct {
	Columns map[string]bool
	Mutex   sync.RWMutex
}

var (
	CachedTableSchema TableSchema //表结构
	CachedUsersSchema UsersSchema //users表

)
