package def

import "github.com/ljhe/scream/msg"

const (
	KeyNodeID        = "NodeID"
	KeyNodeTy        = "NodeTy"
	KeyTranscationID = "TransactionID"
)

func NodeID(id string) msg.Attr        { return msg.Attr{Key: KeyNodeID, Value: id} }
func NodeTy(ty string) msg.Attr        { return msg.Attr{Key: KeyNodeTy, Value: ty} }
func TransactionID(id string) msg.Attr { return msg.Attr{Key: KeyTranscationID, Value: id} }
