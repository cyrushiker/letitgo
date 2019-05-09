// +build go1.12

package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/cyrushiker/letitgo/cmd"
	"github.com/cyrushiker/letitgo/pkg/setting"
)

// AppVer app version
const AppVer = "0.0.1"

func init() {
	setting.AppVer = AppVer
}

func main() {
	app := cli.NewApp()
	app.Name = "Letitgo"
	app.Usage = "An simple search engine"
	app.Version = AppVer
	app.Commands = []cli.Command{
		cmd.Web,
	}
	app.Run(os.Args)
}
