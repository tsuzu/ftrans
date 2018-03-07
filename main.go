package main

import (
	"fmt"
	"log"
	"os"

	easyp2p "github.com/cs3238-tsuzu/go-easyp2p"
	"github.com/urfave/cli"
)

const versionFormat = `ftrans version: %s(%s)

[Details]
ftrans protocol version: %s
go-easyp2p version: %s
`

func stringSlice(s []string) *cli.StringSlice {
	a := cli.StringSlice(s)

	return &a
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf(
			versionFormat,
			binaryVersion,
			binaryRevision,
			protocolVersionLatest,
			easyp2p.P2PVersionString(easyp2p.P2PVersionLatest),
		)
	}

	app := cli.NewApp()
	app.Usage = "Transfer files from one peer to the other peer over encrypted P2P connection"
	app.Version = binaryVersion
	app.UsageText = "ftrans [global options] command [command options] [arguments...]"
	app.Name = "ftrans"

	app.Commands = []cli.Command{
		cli.Command{
			Name:      "send",
			Usage:     "Send files to peer",
			ArgsUsage: "[paths to files you want to send...]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "pass, p",
					Usage: "Password to match the peer(if not set, 6-character password is automatically generated)",
				},
				cli.StringSliceFlag{
					Name:   "stun",
					Usage:  "STUN server addresses(split with ,)",
					EnvVar: "FTRANS_STUN",
					Value:  stringSlice([]string{defaultSTUNServer}),
				},
				cli.StringFlag{
					Name:   "signaling, sig",
					Usage:  "Signaling server address",
					EnvVar: "FTRANS_SIGNALING",
					Value:  defaultSignalingServer,
				},
			},
			Action: func(ctx *cli.Context) error {
				return runClient(false, ctx.String("pass"), []string(ctx.Args()), ctx.StringSlice("stun"), ctx.String("signaling"))
			},
		},
		cli.Command{
			Name:    "receive",
			Aliases: []string{"recv"},
			Usage:   "Receive files from peer",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "pass, p",
					Usage: "Password to match the peer",
				},
				cli.StringSliceFlag{
					Name:   "stun",
					Usage:  "STUN server addresses(split with ,)",
					EnvVar: "FTRANS_STUN",
					Value:  stringSlice([]string{defaultSTUNServer}),
				},
				cli.StringFlag{
					Name:   "signaling",
					Usage:  "Signaling server address",
					EnvVar: "FTRANS_SIGNALING",
					Value:  defaultSignalingServer,
				},
			},
			Action: func(ctx *cli.Context) error {
				return runClient(true, ctx.String("pass"), nil, ctx.StringSlice("stun"), ctx.String("signaling"))
			},
		},
		cli.Command{
			Name:  "signaling",
			Usage: "Launch a signaling server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "addr",
					Usage:  "Listen address",
					EnvVar: "FTRANS_LISTEN",
					Value:  ":80",
				},
			},
			Action: func(ctx *cli.Context) error {
				return runServer(ctx.String("addr"))
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Printf("error: %s", err.Error())
	}
}
