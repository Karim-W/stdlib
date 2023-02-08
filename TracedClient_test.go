package stdlib

import (
	"context"
	"sync"
	"testing"

	trace "github.com/BetaLixT/appInsightsTrace"
	"github.com/soreing/trex"
	"go.uber.org/zap"
)

func TestClientExternalLinks(t *testing.T) {
	l, err := zap.NewProduction()
	if err != nil {
		t.Error(err)
	}
	client := TracedClientProvider(GetTracer(), l)
	ctx := context.TODO()
	ctx = context.WithValue(ctx, TRACE_INFO_KEY, trex.TxModel{
		Ver: "1",
		Tid: "2",
		Pid: "3",
		Rid: "4",
		Flg: "5",
	})
	var res interface{}
	code, err := client.Get(ctx, "https://randomuser.me/api", nil, &res)
	if err != nil {
		t.Error(err)
	}
	if code != 200 {
		t.Error("failed to get data")
	}
}

func TestClientNilPost(t *testing.T) {
	l, err := zap.NewProduction()
	if err != nil {
		t.Error(err)
	}
	client := TracedClientProvider(GetTracer(), l)
	code, err := client.Post(context.Background(), "http://localhost:8080/null", nil, nil, nil)
	if err != nil {
		t.Fatal("Expected error", err)
	}
	if code != 200 {
		t.Fatal("Expected code 200", code)
	}
	code, err = client.Post(context.Background(), "http://localhost:8080/null/error", nil, nil, nil)
	if err == nil {
		t.Fatal("Expected error")
	}
	if code != 400 {
		t.Fatal("Expected code 400", code)
	}
	code, err = client.Post(context.Background(), "http://localhost:8080/null/error/string", nil, nil, nil)
	if err == nil {
		t.Fatal("Expected error")
	}
	if code != 400 {
		t.Fatal("Expected code 400", code)
	}
	code, err = client.Post(context.Background(), "http://localhost:8080/null/error/json", nil, nil, nil)
	if err == nil {
		t.Fatal("Expected error")
	}
	if code != 400 {
		t.Fatal("Expected code 400", code)
	}
	code, err = client.Post(context.Background(), "http://localhost:8080/null/error/bool", nil, nil, nil)
	if err == nil {
		t.Fatal("Expected error")
	}
	if code != 400 {
		t.Fatal("Expected code 400", code)
	}
}

var instance *trace.AppInsightsCore

var mtx sync.Mutex

func GetTracer() *trace.AppInsightsCore {
	mtx.Lock()
	defer mtx.Unlock()
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	if instance == nil {
		instance = trace.NewAppInsightsCore(&trace.AppInsightsOptions{
			ServiceName:        "karim",
			InstrumentationKey: "d6d6d6d6-d6d6-d6d6-d6d6-d6d6d6d6d6d6",
		}, &DefaultTraceExtractor{}, l)
		if instance == nil {
			panic("failed to create tracer")
		}
	}
	return instance
}

type DefaultTraceExtractor struct {
}

const TRACE_INFO_KEY = "tinfo"

func (*DefaultTraceExtractor) ExtractTraceInfo(
	ctx context.Context,
) (ver, tid, pid, rid, flg string) {
	if tinfo, ok := ctx.Value(TRACE_INFO_KEY).(trex.TxModel); !ok {
		return "", "", "", "", ""
	} else {
		return tinfo.Ver, tinfo.Tid, tinfo.Pid, tinfo.Rid, tinfo.Flg
	}
}
