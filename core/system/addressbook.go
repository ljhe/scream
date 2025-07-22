package system

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ljhe/scream/3rd/etcd"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/def"
	"sync"
)

type AddressBook struct {
	Id   string
	Ip   string
	Port int

	IDMap map[string]bool

	sync.RWMutex
}

func New(info iface.AddressInfo) *AddressBook {
	return &AddressBook{
		IDMap: make(map[string]bool),
		Id:    info.Process,
		Ip:    info.Ip,
		Port:  info.Port,
	}
}

func genKey(key, id string) string {
	return fmt.Sprintf(key + "/" + id)
}

func (ab *AddressBook) Register(ctx context.Context, ty, id string, weight int) error {
	if id == "" || ty == "" {
		return fmt.Errorf("node id or type is empty")
	}

	ab.RLock()
	if _, ok := ab.IDMap[id]; ok {
		ab.RUnlock()
		return fmt.Errorf("actor id %v already registered", id)
	}
	ab.RUnlock()

	// serialize address info to json
	addrJSON, _ := json.Marshal(iface.AddressInfo{
		Process: ab.Id,
		NodeId:  id,
		NodeTy:  ty,
		Ip:      ab.Ip,
		Port:    ab.Port},
	)

	etcd.Register(ctx, genKey(def.AddressBookIDField, id), addrJSON)

	ab.Lock()
	ab.IDMap[id] = true
	ab.Unlock()

	return nil
}

func (ab *AddressBook) Unregister(ctx context.Context, id string, weight int) error {
	logrus.Infof("addressBook unregister id:%s weight:%d", id, weight)

	if id == "" {
		return fmt.Errorf("node id or type is empty")
	}

	err := etcd.UnRegister(ctx, genKey(def.AddressBookIDField, id))

	if err == nil {
		ab.Lock()
		delete(ab.IDMap, id) // try delete
		ab.Unlock()
	}

	return err
}

func (ab *AddressBook) GetByID(ctx context.Context, id string) (iface.AddressInfo, error) {

	if id == "" {
		return iface.AddressInfo{}, fmt.Errorf("node id or type is empty")
	}

	ab.RLock()
	if _, ok := ab.IDMap[id]; ok {
		ab.RUnlock()
		return iface.AddressInfo{Process: ab.Id, Ip: ab.Ip, Port: ab.Port, NodeId: id}, nil // return local node info directly
	}
	ab.RUnlock()

	//resp, err := etcd.GetEtcdDiscovery().Cli.Get(ctx, genKey(def.AddressBookIDField, id), clientv3.WithPrefix())
	//if err != nil || resp.Count == 0 {
	//	return iface.AddressInfo{}, fmt.Errorf("etcd get node error:%v id:%s", err, id)
	//}

	var addr iface.AddressInfo
	//err = json.Unmarshal(resp.Kvs[0].Value, &addr)
	//if err != nil {
	//	return iface.AddressInfo{}, fmt.Errorf("failed to unmarshal address: %v", err)
	//}

	return addr, nil
}

func (ab *AddressBook) GetByType(ctx context.Context, s string) ([]iface.AddressInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (ab *AddressBook) GetWildcardNode(ctx context.Context, nodeType string) (iface.AddressInfo, error) {
	panic("implement me")
}

func (ab *AddressBook) GetLowWeightNodeForNode(ctx context.Context, nodeType string) (iface.AddressInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (ab *AddressBook) GetNodeTypeCount(ctx context.Context, nodeType string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (ab *AddressBook) Clear(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
