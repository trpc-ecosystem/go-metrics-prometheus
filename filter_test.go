package prometheus

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	"trpc.group/trpc-go/trpc-go/metrics"
)

// TestFilter active and passive calls to report unit tests.
func TestFilter(t *testing.T) {
	setup(t)
	type ExampleStruct struct {
		body string
	}
	ctx, msg := codec.WithNewMessage(context.Background())
	msg.WithCalleeApp("App")
	msg.WithCalleeServer("Server")
	msg.WithCalleeService("Service")
	msg.WithCalleeMethod("Method")
	msg.WithCallerServer("CallerApp")
	msg.WithCallerApp("CallerApp")
	msg.WithCallerServer("CallerServer")
	msg.WithCallerService("CallerService")
	msg.WithCallerMethod("CallerMethod")
	msg.WithCalleeContainerName("ContainerName")
	msg.WithCalleeSetName("SetName")
	testIP := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 80}
	t.Log(testIP.String())
	msg.WithRemoteAddr(testIP)
	msg.WithLocalAddr(testIP)

	clientHandleFunc := func(ctx context.Context, req interface{}, rsp interface{}) (err error) {
		return nil
	}
	serverHandleFunc := func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		return nil, nil
	}
	req := &ExampleStruct{
		body: "req",
	}
	rsp := &ExampleStruct{}
	filterReport(t, ctx, req, rsp, clientHandleFunc, serverHandleFunc)
	t.Log(getMetrics(t))

	clientHandleFunc = func(ctx context.Context, req interface{}, rsp interface{}) (err error) {
		return errs.ErrServerTimeout
	}
	serverHandleFunc = func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		return nil, errs.ErrServerTimeout
	}
	filterReportErr(t, ctx, req, rsp, clientHandleFunc, serverHandleFunc)
	t.Log(getMetrics(t))

	clientHandleFunc = func(ctx context.Context, req interface{}, rsp interface{}) (err error) {
		return errors.New("myerr")
	}
	serverHandleFunc = func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		return nil, errors.New("myerr")
	}
	filterReportErr(t, ctx, req, rsp, clientHandleFunc, serverHandleFunc)
	t.Log(getMetrics(t))

}

func filterReport(t *testing.T, ctx context.Context, req, rsp interface{},
	clientHandleFunc filter.ClientHandleFunc, serverHandleFunc filter.ServerHandleFunc) {
	//case1: active call report.
	err := ClientFilter(ctx, req, rsp, clientHandleFunc)
	assert.Nil(t, err)

	//case2: passive call report.
	_, err = ServerFilter(ctx, req, serverHandleFunc)
	assert.Nil(t, err)
}

func filterReportErr(t *testing.T, ctx context.Context, req, rsp interface{},
	clientHandleFunc filter.ClientHandleFunc, serverHandleFunc filter.ServerHandleFunc) {
	//case1: active call report.
	err := ClientFilter(ctx, req, rsp, clientHandleFunc)
	assert.NotNil(t, err)

	//case2: passive call report.
	_, err = ServerFilter(ctx, req, serverHandleFunc)
	assert.NotNil(t, err)
}

func TestSetBounds(t *testing.T) {
	expectedBounds := metrics.NewValueBounds(1.0, 10.0, 20.0)
	SetClientBounds(expectedBounds)
	assert.Equal(t, expectedBounds, clientBounds)

	expectedBounds = metrics.NewValueBounds(1.0, 10.0, 20.0, 40.0)
	SetServerBounds(expectedBounds)
	assert.Equal(t, expectedBounds, serverBounds)
}
