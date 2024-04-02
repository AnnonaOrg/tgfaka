package cache

var Cache *CacheStruct
var ScheduleCache *CacheStruct

func InitCache() {
	Cache = NewCache()
	ScheduleCache = NewCache()
}
