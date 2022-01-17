package network

import (
	"GoGameServer/src/lib"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

//网络属性
type NetworkProperty struct {
	MaxConnections         int32  //最大连接数
	MaxReadPacketLength    uint32 //可接收的最大包长度 不含包头 min(uint32上限,int上限) 与操作系统位数有关
	MaxWritePacketLength   uint32 //可发送的最大包长度 不含包头 min(uint32上限,int上限) 与操作系统位数有关
	EventChanCapacity      uint32 //事件chan的长度上限
	WriteChanCapacity      uint32 //发送chan的长度上限
	HeartBeatWriteInterval int64  //心跳发送间隔
	HeartBeatReadTimeout   int64  //心跳接收超时时间 超时后自动关闭
}

type Network struct {
	mutex              sync.Mutex
	createSessionMutex sync.Mutex
	property           *NetworkProperty       //属性
	ticker             *time.Ticker           //帧循环
	closeChan          chan int               //用来关闭帧循环的channel
	endChan            chan int               //帧循环结束信号channel
	internalChan       chan map[string]string //内部循环channel
	internalCloseChan  chan int               //用来关闭内部循环的channel
	internalEndChan    chan int               //内部循环结束信号channel

	address           string
	log               *zap.SugaredLogger
	listener          net.Listener
	sessionMap        map[uint64]*Session //TODO session通过sessionCreateChan加入map 会有一定延迟 要注意 TODO 用最小堆优化
	sessionNum        int32               //session数量 可能不准
	sessionCreateChan chan *Session       //用于新建session
	sessionCloseChan  chan *Session       //用于关闭移除session
	sessionUniqueId   uint64              //TODO 用atomic
	isRunning         *lib.AtomBool       //是否在运行
	isClosing         *lib.AtomBool       //是否在关闭

	//EventChan chan *Event //会话事件需要外部读取处理
}

func (n *Network) newAcceptSession(c net.Conn) (session *Session, err error) {
	n.sessionUniqueId++
	session = NewSession(int(n.sessionUniqueId), n, c)
	return
}
