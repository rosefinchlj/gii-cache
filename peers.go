package giicache

type PeerGetter interface {
	// Get 获取key对应的value
	Get(group, key string) ([]byte, error)
}

type PeerPicker interface {
	// PickPeer 获取一个http客户端
	PickPeer(key string) (peer PeerGetter, ok bool)
}
