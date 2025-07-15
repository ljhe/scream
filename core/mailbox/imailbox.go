package mailbox

type IMailBox interface {
	Start() error
	Stop()
	Push(msg interface{}) error
}
