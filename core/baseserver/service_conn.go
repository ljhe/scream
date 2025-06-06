package baseserver

import (
	trdetcd "github.com/ljhe/scream/3rd/etcd"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/def"
	"log"
	"math/rand"
	"sync"
)

var (
	// 服务器节点之间的连接
	serviceConnBySid = map[string]iface.ISession{}
	mu               sync.RWMutex
)

func AddServiceNode(session iface.ISession, sid, name string, from string) {
	mu.Lock()
	defer mu.Unlock()
	typ, zone, idx, err := trdetcd.ParseServiceId(sid)
	if err != nil {
		log.Println("AddServiceNode error:", err)
		return
	}
	session.(iface.IContextSet).SetContextData(def.ContextSetCtxKey, &trdetcd.ServerInfo{
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
	mu.Lock()
	defer mu.Unlock()
	sid := ""
	if session == nil {
		return sid
	}

	ed := SessionContextEtcd(session)
	if ed == nil {
		return sid
	}
	delete(serviceConnBySid, ed.Id)
	sid = ed.Id
	log.Printf("remove service node success. sessionId:%v sid:%v \n", session.GetId(), ed.Id)
	return sid
}

func RemoveServiceNodeByName(sid string) {
	mu.Lock()
	defer mu.Unlock()
	if sid == "" {
		return
	}
	delete(serviceConnBySid, sid)
}

// GetServiceNode 通过sid获取服务器节点连接的session
func GetServiceNode(sid string) iface.ISession {
	mu.RLock()
	defer mu.RUnlock()

	if sess, ok := serviceConnBySid[sid]; ok {
		return sess
	}
	return nil
}

func SessionContextEtcd(session iface.ISession) *trdetcd.ServerInfo {
	if ed, ok := session.(iface.IContextSet).GetContextData(def.ContextSetCtxKey); ok {
		return ed.(*trdetcd.ServerInfo)
	}
	return nil
}

func GetServiceNodeAndSession(serviceName string, serviceTypeName string, id uint64) (string, iface.ISession) {
	if serviceName != "" {
		tmpSess := GetServiceNode(serviceName)
		if tmpSess == nil {
			RemoveServiceNodeByName(serviceName)
		} else {
			return serviceName, tmpSess
		}
	}

	tmpServiceName, tmpSess := SelectServiceNodeAndSession(serviceTypeName, id)
	if tmpServiceName != "" {
		return tmpServiceName, tmpSess
	} else {
		return "", nil
	}
}

func SelectServiceNodeAndSession(serviceName string, id uint64) (string, iface.ISession) {
	serviceNode := SelectServiceNode(serviceName, id)
	if serviceNode == "" {
		return serviceNode, nil
	}

	serviceSess := GetServiceNode(serviceNode)
	if serviceSess == nil {
		RemoveServiceNodeByName(serviceNode)
		for {
			serviceNode = SelectServiceNode(serviceName, 0)
			if serviceNode == "" {
				break
			}
			serviceSess = GetServiceNode(serviceNode)
			if serviceSess == nil {
				RemoveServiceNodeByName(serviceNode)
			} else {
				break
			}
		}
	}
	return serviceNode, serviceSess
}

func SelectServiceNode(serviceName string, id uint64) string {
	if id == 0 {
		id = uint64(rand.Int31n(100))
	}
	switch serviceName {
	case def.ServiceNodeTypeGateStr:
		fallthrough
	case def.ServiceNodeTypeGameStr:
		return selectServiceNode(serviceName, id)
	default:
		return ""
	}
}

// id确定的某一个服务器节点
func selectServiceNode(serviceName string, id uint64) string {
	mu.RLock()
	defer mu.RUnlock()

	var retIDList []string
	for _, node := range serviceConnBySid {
		if raw, ok := node.(iface.IContextSet).GetContextData("ctx"); ok {
			sid := raw.(*trdetcd.ServerInfo)
			if sid.Name == serviceName {
				retIDList = append(retIDList, sid.Id)
			}
		}
	}
	if len(retIDList) <= 0 {
		return ""
	}
	modNum := int(id % uint64(len(retIDList)))
	return retIDList[modNum]
}
