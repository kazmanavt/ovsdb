package main

import (
	"context"
	"fmt"
	"github.com/kazmanavt/ovsdb/v2/client"
	"github.com/kazmanavt/ovsdb/v2/monitor"
	"github.com/kazmanavt/ovsdb/v2/types"
	"log/slog"
	"os"
)

func main() {
	fmt.Println("Hi there!")

	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cli := client.NewClient("unix", "/var/run/openvswitch/db.sock",
		client.WithJLogger(l),
		client.WithLogger(l))
	defer cli.Close()

	sch, err := cli.GetSchema(context.Background(), "Open_vSwitch")
	if err != nil {
		l.Error("failed to get schema", slog.String("error", err.Error()))
		return
	}

	upd, uChan, err := cli.SetMonitorCondSince(context.Background(), "Open_vSwitch", "test",
		monitor.NewMonCondReqSet(sch).Add("Interface", monitor.MonCondReq{
			Columns: []string{
				"name",
				"admin_state",
			},
			Where: []types.Condition{
				types.Equal("name", "eno1u1"),
				types.Equal("name", "eno2u1"),
				types.Equal("name", "eno4u1"),
				types.Equal("name", "eno4d1"),
			},
		}))
	if err != nil {
		l.Error("failed to set monitor", slog.String("error", err.Error()))
		return
	}
	l.Info("monitor set", slog.Any("upd", upd))
	for u := range uChan {
		l.Info("update", slog.Any("u", u))
	}
}
