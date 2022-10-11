package main

import (
	"errors"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func exit(err *error) {
	if *err != nil {
		log.Println("exited with error:", (*err).Error())
		os.Exit(1)
	} else {
		log.Println("exited")
	}
}

func main() {
	var err error
	defer exit(&err)

	app := cli.NewApp()
	SetupCLIGates(app)
	app.Usage = "file based kubernetes operation tool"
	app.Commands = append(app.Commands, &cli.Command{
		Name:        "pull",
		Description: "pull resources from existing cluster",
		Action: func(c *cli.Context) error {
			ExtractCLIGates(c)
			if c.NArg() != 4 {
				return errors.New("invalid number of arguments")
			}
			return DoPull(c.Context, c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), c.Args().Get(3))
		},
	})
	app.Commands = append(app.Commands, &cli.Command{
		Name:        "push",
		Description: "push resources to existing cluster",
		Action: func(c *cli.Context) error {
			if c.NArg() != 4 {
				return errors.New("invalid number of arguments")
			}
			ExtractCLIGates(c)
			return DoPush(c.Context, c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), c.Args().Get(3))
		},
	})
	err = app.Run(os.Args)
}
