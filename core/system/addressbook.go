package system

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ljhe/scream/3rd/etcd"
	"github.com/ljhe/scream/3rd/log"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/def"
	"strings"
	"sync"
)

type AddressBook struct {
	Id   string
	Ip   string
	Port int

	IDMap map[string]iface.AddressInfo

	ctx    context.Context
	cancel context.CancelFunc
	sync.RWMutex
}

func New(info iface.AddressInfo) *AddressBook {
	ab := &AddressBook{
		IDMap: make(map[string]iface.AddressInfo),
		Id:    info.Process,
		Ip:    info.Ip,
		Port:  info.Port,
	}

	ab.ctx, ab.cancel = context.WithCancel(context.Background())

	go ab.Watch(ab.ctx)
	return ab
}

func genKey(key, id string) string {
	return fmt.Sprintf(key + "/" + id)
}

func splitKey(key string) string {
	str := strings.Split(key, "/")
	if len(str) < 2 {
		panic("split key error")
	}
	return str[1]
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

	return nil
}

func (ab *AddressBook) Unregister(ctx context.Context, id string, weight int) error {
	log.InfoF("addressBook unregister id:%s weight:%d", id, weight)

	if id == "" {
		return fmt.Errorf("node id or type is empty")
	}

	err := etcd.UnRegister(ctx, genKey(def.AddressBookIDField, id))

	if err == nil {
		ab.delIDMap(id)
	}

	return err
}

func (ab *AddressBook) Watch(ctx context.Context) {
	etcd.Discovery(ctx, def.AddressBookIDField, func(key string, val []byte) {
		// if val's len=0 EventTypeDelete, else EventTypePut
		if len(val) == 0 {
			ab.delIDMap(splitKey(key))
			return
		}
		ab.setIDMap(splitKey(key), val)
	})
}

func (ab *AddressBook) GetByID(ctx context.Context, id string) (iface.AddressInfo, error) {

	if id == "" {
		return iface.AddressInfo{}, fmt.Errorf("addressbook node not found id is empty")
	}

	ab.RLock()
	defer ab.RUnlock()
	if val, ok := ab.IDMap[id]; ok {
		return iface.AddressInfo{Process: val.Process, Ip: val.Ip, Port: val.Port, NodeId: val.NodeId}, nil
	}

	return iface.AddressInfo{}, fmt.Errorf("addressbook node not found by id:%s", id)
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
	return nil
}

func (ab *AddressBook) setIDMap(key string, val []byte) {
	ab.Lock()
	defer ab.Unlock()
	var addr iface.AddressInfo
	err := json.Unmarshal(val, &addr)
	if err != nil {
		panic("failed to unmarshal address")
	}
	ab.IDMap[key] = addr
}

func (ab *AddressBook) delIDMap(key string) {
	ab.Lock()
	defer ab.Unlock()
	delete(ab.IDMap, key) // try delete
}
