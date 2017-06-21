package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/michaelklishin/rabbit-hole"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func getRMQClient(c *cli.Context) (*rabbithole.Client, error) {
	return rabbithole.NewClient(c.String("url"), c.String("username"), c.String("password"))
}

func listFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Usage: "RabbitMQ HTTP API url",
		},
	}
}

func validateListFlags(c *cli.Context) error {
	if len(c.String("url")) == 0 {
		return errors.New("RabbitMQ HTTP API url not specified")
	}
	return nil
}

func deleteFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Usage: "RabbitMQ HTTP API url",
		},
		cli.StringFlag{
			Name:  "vhost",
			Usage: "VHost name",
		},
		cli.BoolFlag{
			Name:  "match, m",
			Usage: "Match regex against name",
		},
	}
}

func validateDeleteFlags(c *cli.Context) error {
	if len(c.String("url")) == 0 {
		return errors.New("RabbitMQ HTTP API url not specified")
	}

	if len(c.String("vhost")) == 0 {
		return errors.New("VHost not specified")
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

func listQueues(c *cli.Context) error {
	rmqc, err := newRMQClient(c)
	if err != nil {
		return errors.Wrap(err, "creating RabbitMQ HTTP client failed")
	}

	queueInfoList, err := rmqc.ListQueues()
	if err != nil {
		return errors.Wrap(err, "listing RabbitMQ queue failed")
	}

	fmt.Printf("Name,VHost,Durable,AutoDelete\n")
	for _, queueInfo := range queueInfoList {
		fmt.Printf("%s,%s,%t,%t\n", queueInfo.Name, queueInfo.Vhost, queueInfo.Durable, queueInfo.AutoDelete)
	}

	return nil
}

func deleteQueues(c *cli.Context) error {
	rmqc, err := newRMQClient(c)
	if err != nil {
		return errors.Wrap(err, "creating RabbitMQ HTTP client failed")
	}

	queueInfoList, err := rmqc.ListQueues()
	if err != nil {
		return errors.Wrap(err, "listing RabbitMQ queue failed")
	}

	var queueNames []string

	if c.Bool("match") {
		var patterns []*regexp.Regexp

		for _, arg := range c.Args() {
			pattern, err := regexp.Compile(arg)
			if err != nil {
				return err
			}

			patterns = append(patterns, pattern)
		}

		for _, queueInfo := range queueInfoList {
			for _, pattern := range patterns {
				if pattern.MatchString(queueInfo.Name) {
					queueNames = append(queueNames, queueInfo.Name)
				}
			}
		}
	} else {
		queueNames = c.Args()
	}

	vhost := c.String("vhost")
	for _, queueName := range queueNames {
		fmt.Println("Deleting queue", queueName)

		resp, err := rmqc.DeleteQueue(vhost, queueName)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("%s: %s", resp.Status, body)
		}
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
					Flags:  listFlags(),
					Before: validateListFlags,
					Action: listUsers,
				},
			},
		},
		{
			Name:  "queues",
			Usage: "queue management commands",
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "lists queues",
					Flags:  listFlags(),
					Before: validateListFlags,
					Action: listQueues,
				},
				{
					Name:      "delete",
					Usage:     "deletes one or more queue",
					Flags:     deleteFlags(),
					ArgsUsage: "<name-or-regex> ...",
					Before:    validateDeleteFlags,
					Action:    deleteQueues,
				},
			},
		},
	}

	app.Run(os.Args)
}
