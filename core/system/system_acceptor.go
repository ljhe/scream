package system

import (
	"context"
	"fmt"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/ljhe/scream/3rd/log"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/lib/grpc"
	"github.com/ljhe/scream/lib/warpwaitgroup"
	"github.com/ljhe/scream/msg"
	"github.com/ljhe/scream/msg/router"
	realgrpc "google.golang.org/grpc"
	"runtime"
	"strconv"
)

type Acceptor struct {
	server *grpc.Server
}

type listen struct {
	router.AcceptorServer
	sys iface.ISystem
}

// Stack returns a formatted stack trace of the goroutine that calls it.
// It calls runtime.Stack with a large enough buffer to capture the entire trace.
// If all is true, Stack formats stack traces of all other goroutines
// into buf after the trace for the current goroutine.
func stack(all bool) []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, all)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

func recoverHandler(r interface{}) error {
	err, ok := r.(error)
	if !ok {
		err = fmt.Errorf("%v", r)
	}
	buf := stack(false)
	fmt.Println(fmt.Errorf("PANIC: %v\n%s", err, buf).Error())
	return fmt.Errorf("[GRPC-SERVER RECOVER] err: %v stack: %s", err, buf)
}

func NewAcceptor(sys iface.ISystem, port int) (*Acceptor, error) {

	var unaryInterceptors []realgrpc.UnaryServerInterceptor

	unaryInterceptors = append(unaryInterceptors,
		grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(recoverHandler)))

	a := &Acceptor{
		server: grpc.BuildServerWithOption(
			grpc.WithServerListen(":"+strconv.Itoa(port)),
			grpc.WithServerGracefulStop(),
			grpc.ServerRegisterHandler(func(s *realgrpc.Server) {
				router.RegisterAcceptorServer(s, &listen{sys: sys})
			}),
			grpc.ServerAppendUnaryInterceptors(unaryInterceptors...),
		),
	}

	err := a.server.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize acceptor server: %v", err)
	}

	return a, nil
}

func (acceptor *Acceptor) Exit() {
	acceptor.server.Close()
}

// Routing acceptor routing
func (s *listen) Routing(ctx context.Context, req *router.RouteReq) (*router.RouteRes, error) {
	res := &router.RouteRes{}

	ctx = context.WithValue(ctx, msg.WaitGroupKey{}, &warpwaitgroup.WrapWaitGroup{})

	routermsg := msg.NewBuilder(ctx).Build()
	routermsg.Req = req.Msg
	routermsg.Req.Header.PrevActorType = "GrpcAcceptor"

	err := s.sys.Call(
		req.Msg.Header.TargetActorID,
		req.Msg.Header.TargetActorType,
		req.Msg.Header.Event, routermsg)

	if err != nil {
		log.InfoF("listen routing %v err %v", req.Msg.Header.Event, err.Error())
	}

	res.Msg = routermsg.Res
	return res, nil
}
