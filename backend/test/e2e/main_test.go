package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/holdennekt/sgame/backend/test/e2e/testhelper"
)

var containers *testhelper.Containers

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	containers, err = testhelper.StartContainers(ctx)
	if err != nil {
		panic("start containers: " + err.Error())
	}

	code := m.Run()

	containers.Terminate(ctx)
	os.Exit(code)
}
