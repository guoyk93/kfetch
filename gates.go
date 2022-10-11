package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"strings"
	"sync"
)

var (
	gateValues             = make(map[Gate]bool)
	gateLock   sync.Locker = &sync.Mutex{}
)

type Gate string

func (g Gate) Set() {
	gateLock.Lock()
	defer gateLock.Unlock()
	gateValues[g] = true
}

func (g Gate) Clear() {
	gateLock.Lock()
	defer gateLock.Unlock()
	gateValues[g] = false
}

func (g Gate) IsOn() bool {
	gateLock.Lock()
	defer gateLock.Unlock()
	return gateValues[g]
}

const (
	GateNoUpdate          Gate = "no-update"
	GateNoResourceVersion Gate = "no-resource-version"
	GateZeroReplicas      Gate = "zero-replicas"
)

var Gates = []Gate{
	GateNoUpdate,
	GateNoResourceVersion,
	GateZeroReplicas,
}

func SetupCLIGates(app *cli.App) {
	for _, gate := range Gates {
		app.Flags = append(app.Flags, &cli.BoolFlag{
			Name:  string(gate),
			Usage: "toggle feature gate: " + string(gate),
			EnvVars: []string{
				"KOOP_" + strings.ReplaceAll(strings.ToUpper(string(gate)), "-", "_"),
			},
		})
	}
}

func ExtractCLIGates(c *cli.Context) {
	for _, gate := range Gates {
		if c.Bool(string(gate)) {
			log.Println("GATE:", string(gate))
			gate.Set()
		} else {
			gate.Clear()
		}
	}
}
