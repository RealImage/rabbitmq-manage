package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/michaelklishin/rabbit-hole"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func getRMQClient(c *cli.Context) (*rabbithole.Client, error) {
	return rabbithole.NewClient(c.String("url"), c.String("username"), c.String("password"))
}

func serverFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Usage: "RabbitMQ HTTP API url",
		},
	}
}

func validateServerFlags(c *cli.Context) error {
	if len(c.String("url")) == 0 {
		return errors.New("RabbitMQ HTTP API url not specified")
	}
	return nil
}

func newRMQClient(c *cli.Context) (*rabbithole.Client, error) {
	u, err := url.Parse(c.String("url"))
	if err != nil {
		return nil, err
	}

	password, _ := u.User.Password()
	return rabbithole.NewClient(u.Scheme+"://"+u.Host, u.User.Username(), password)

}

func listUsers(c *cli.Context) error {
	rmqc, err := newRMQClient(c)
	if err != nil {
		return errors.Wrap(err, "creating RabbitMQ HTTP client failed")
	}

	userInfoList, err := rmqc.ListUsers()
	if err != nil {
		return errors.Wrap(err, "listing RabbitMQ users failed")
	}

	fmt.Printf("Name,Tags\n")
	for _, userInfo := range userInfoList {
		fmt.Printf("%s,%s\n", userInfo.Name, userInfo.Tags)
	}

	return nil
}

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:  "users",
			Usage: "user management commands",
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "lists users",
					Flags:  serverFlags(),
					Before: validateServerFlags,
					Action: listUsers,
				},
			},
		},
	}

	app.Run(os.Args)
}
