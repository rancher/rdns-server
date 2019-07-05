package main

import (
	"fmt"
	"os"

	"github.com/rancher/rdns-server/command/etcdv3"
	"github.com/rancher/rdns-server/command/route53"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	DNSVersion = "v0.5.2"
	DNSDate    string
)

func init() {
	cli.VersionPrinter = versionPrinter
}

func main() {
	app := cli.NewApp()
	app.Author = "Rancher Labs, Inc."
	app.Before = beforeFunc
	app.EnableBashCompletion = true
	app.Name = os.Args[0]
	app.Usage = fmt.Sprintf("control and configure RDNS(%s)", DNSDate)
	app.Version = DNSVersion
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug, d",
			EnvVar: "DEBUG",
			Usage:  "used to set debug mode.",
		},
		cli.StringFlag{
			Name:   "listen",
			EnvVar: "LISTEN",
			Usage:  "used to set listen port.",
			Value:  ":9333",
		},
		cli.StringFlag{
			Name:   "frozen",
			EnvVar: "FROZEN",
			Usage:  "used to set the duration when the domain name can be used again.",
			Value:  "2160h",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "route53",
			Aliases: []string{"r53"},
			Usage:   "use aws route53 backend",
			Flags:   route53.Flags(),
			Action:  route53.Action,
		},
		{
			Name:    "etcdv3",
			Aliases: []string{"ev3"},
			Usage:   "use etcd-v3 backend",
			Flags:   etcdv3.Flags(),
			Action:  etcdv3.Action,
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func beforeFunc(c *cli.Context) error {
	if os.Getuid() != 0 {
		logrus.Fatalf("%s: need to be root", os.Args[0])
	}
	return nil
}

func versionPrinter(c *cli.Context) {
	if _, err := fmt.Fprintf(c.App.Writer, DNSVersion); err != nil {
		logrus.Error(err)
	}
}
