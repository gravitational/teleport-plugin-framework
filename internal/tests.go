package internal

import (
	"context"
	"os"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/gravitational/teleport-plugin-framework/lib/wasm"
	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
)

const (
	// defaultTimeout default timeout for WASM method calls
	defaultTimeout = time.Second * 30
	// testWasm WASM test binary file
	testWasm = "build/test.wasm"
)

// RunTests runs tests
func RunTests(log logrus.FieldLogger, concurrency int) {
	b, err := os.ReadFile(testWasm)
	if err != nil {
		bail("%v", trace.Wrap(err))
	}

	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		bail("%v", trace.Wrap(err))
	}

	testRunner, err := wasm.NewTesting(fixturesDir)
	if err != nil {
		log.Fatal(trace.Wrap(err))
	}
	protobufInterop := wasm.NewProtobufInterop()
	api := wasm.NewTeleportAPI(log, testRunner.MockAPIClient, protobufInterop)

	opts := wasm.ExecutionContextPoolOptions{
		Timeout:     defaultTimeout,
		Concurrency: concurrency,
		Bytes:       b,
		TraitFactories: []wasm.TraitFactory{
			wasm.NewAssemblyScriptEnv(log),
			wasm.NewStore(wasm.NewBadgerPersistentStore(db), wasm.DecodeAssemblyScriptString),
			testRunner,
			protobufInterop,
			api,
		},
	}

	pool, err := wasm.NewExecutionContextPool(opts)
	if err != nil {
		bail("%v", trace.Wrap(err))
	}

	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func() {
			_, err = pool.Execute(context.Background(), func(ectx *wasm.ExecutionContext) (interface{}, error) {
				runner, err := testRunner.For(ectx)
				if err != nil {
					return nil, trace.Wrap(err)
				}

				err = runner.Run()
				if err != nil {
					return nil, trace.Wrap(err)
				}

				return nil, nil
			})

			if err != nil {
				bail("%v", trace.Wrap(err))
			}

			wg.Done()
		}()
	}

	wg.Wait()
}
