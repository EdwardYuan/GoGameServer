package cacheserver

// 缓存服务器用来快速存取对于性能要求较高的数据，例如战斗时需要获取玩家数据
// 好友关系或公会数据等
type CacheServer struct {
	BaseData     interface{} // 后续改成对应的数据结构类型
	RelationData interface{}
	/* …… */
}
