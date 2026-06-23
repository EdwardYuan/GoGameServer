package service_common

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// ServiceKeyPrefix 是 etcd 中服务发现 key 的统一前缀。
const ServiceKeyPrefix = "services/"

// ServerInfo 是写入 etcd 的服务注册信息。
// Type 用来区分 game/gate/proxy，Proxy 转发时依赖该字段判断目标连接类型。
type ServerInfo struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	IP   string `json:"ip"`
	Port int32  `json:"port"`
}

func NewServerInfo(id int32, name string, serviceType string, ip string, port int32) ServerInfo {
	return ServerInfo{
		Id:   id,
		Name: name,
		Type: serviceType,
		IP:   ip,
		Port: port,
	}
}

// ServiceKey 统一生成 etcd 服务 key，避免出现 name -> type 和 services/name -> JSON 两套格式。
func ServiceKey(name string) string {
	return ServiceKeyPrefix + name
}

// MarshalServerInfo 将服务信息编码成 etcd 中存储的 JSON 字符串。
func MarshalServerInfo(info *ServerInfo) (string, error) {
	data, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalServerInfo 从 etcd 的 JSON 字符串还原服务信息。
func UnmarshalServerInfo(value string) (*ServerInfo, error) {
	info := &ServerInfo{}
	if err := json.Unmarshal([]byte(value), info); err != nil {
		return nil, err
	}
	return info, nil
}

// NormalizeEtcdEndpoint 允许配置只写 host:port，实际连接 etcd 时补齐 scheme。
func NormalizeEtcdEndpoint(endpoint string) string {
	if endpoint == "" {
		endpoint = "127.0.0.1:2379"
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}
	return endpoint
}

// RegisterService 使用统一格式把服务注册到 etcd。
// 当前第一阶段只由 proxy/gate/game 调用，dbserver/login 暂不注册。
func RegisterService(endpoint string, info ServerInfo) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{NormalizeEtcdEndpoint(endpoint)},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	defer cli.Close()
	value, err := MarshalServerInfo(&info)
	if err != nil {
		return err
	}
	_, err = cli.Put(context.Background(), ServiceKey(info.Name), value)
	return err
}
