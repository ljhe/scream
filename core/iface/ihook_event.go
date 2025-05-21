package iface

type IHookEvent interface {
	InEvent(iv IProcEvent) IProcEvent  // 接收事件
	OutEvent(ov IProcEvent) IProcEvent // 发送事件
}
