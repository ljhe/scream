package baseserver

import (
	"common"
	"common/iface"
	plugins "common/plugins/etcd"
	"log"
	"sync"
)

var (
	serviceNode sync.RWMutex
	// 服务器节点之间的连接
	serviceConnBySid = map[string]iface.ISession{}
)

func AddServiceNode(session iface.ISession, sid, name string, from string) {
	serviceNode.Lock()
	defer serviceNode.Unlock()
	typ, zone, idx, err := plugins.ParseServiceId(sid)
	if err != nil {
		log.Println("AddServiceNode error:", err)
		return
	}
	session.(common.ContextSet).SetContextData(common.ContextSetCtxKey, &plugins.ETCDServiceDesc{
		Id:    sid,
		Name:  name,
		Typ:   typ,
		Zone:  zone,
		Index: idx,
	})
	serviceConnBySid[sid] = session
	log.Printf("AddServiceNode success. from:%v sid:%v", from, sid)
	return
}

func RemoveServiceNode(session iface.ISession) string {
	sid := ""
	if session == nil {
		return sid
	}

	ed := SessionContextEtcd(session)
	if ed == nil {
		return sid
	}
	serviceNode.Lock()
	defer serviceNode.Unlock()
	delete(serviceConnBySid, ed.Id)
	sid = ed.Id
	log.Printf("remove service node success. sessionId:%v sid:%v \n", session.GetId(), ed.Id)
	return sid
}

// GetServiceNode 通过sid获取服务器节点连接的session
func GetServiceNode(sid string) iface.ISession {
	serviceNode.RLock()
	defer serviceNode.RUnlock()

	if sess, ok := serviceConnBySid[sid]; ok {
		return sess
	}
	return nil
}

func SessionContextEtcd(session iface.ISession) *plugins.ETCDServiceDesc {
	if ed, ok := session.(common.ContextSet).GetContextData(common.ContextSetCtxKey); ok {
		return ed.(*plugins.ETCDServiceDesc)
	}
	return nil
}
