package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const usageText = `EventBus Command Line Client

Usage:
  ebclient [global-flags] <command> [command-flags] [args]

Commands:
  create-topic <topic>         创建主题
  delete-topic <topic>         删除主题
  subscribe <topic> <sub-id>   订阅主题 (--priority, --endpoint)
  unsubscribe <topic> <sub-id> 取消订阅
  list <topic>                 列出主题订阅者
  publish <topic> <json>       发布事件 (--async)
  status <event-id>            查询事件状态

Global Flags:
  --server <url>               服务端地址 (默认 http://localhost:8080)
`

func main() {
	server := flag.String("server", "http://localhost:8080", "eventbus server url")
	flag.Usage = func() {
		fmt.Print(usageText)
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	client := NewClient(*server)
	cmd := strings.ToLower(args[0])

	var err error
	switch cmd {
	case "create-topic":
		err = cmdCreateTopic(client, args[1:])
	case "delete-topic":
		err = cmdDeleteTopic(client, args[1:])
	case "subscribe":
		err = cmdSubscribe(client, args[1:])
	case "unsubscribe":
		err = cmdUnsubscribe(client, args[1:])
	case "list":
		err = cmdList(client, args[1:])
	case "publish":
		err = cmdPublish(client, args[1:])
	case "status":
		err = cmdStatus(client, args[1:])
	default:
		fmt.Printf("unknown command: %s\n\n", cmd)
		flag.Usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func cmdCreateTopic(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: create-topic <topic>")
	}
	return c.CreateTopic(args[0])
}

func cmdDeleteTopic(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: delete-topic <topic>")
	}
	return c.DeleteTopic(args[0])
}

func cmdSubscribe(c *Client, args []string) error {
	fs := flag.NewFlagSet("subscribe", flag.ContinueOnError)
	priority := fs.Int("priority", 0, "subscriber priority")
	endpoint := fs.String("endpoint", "", "webhook endpoint url")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) < 2 {
		return fmt.Errorf("usage: subscribe <topic> <sub-id> [--priority=N] [--endpoint=url]")
	}
	return c.Subscribe(rest[0], rest[1], *priority, *endpoint)
}

func cmdUnsubscribe(c *Client, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: unsubscribe <topic> <sub-id>")
	}
	return c.Unsubscribe(args[0], args[1])
}

func cmdList(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: list <topic>")
	}
	return c.ListSubscribers(args[0])
}

func cmdPublish(c *Client, args []string) error {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	async := fs.Bool("async", false, "async mode")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) < 2 {
		return fmt.Errorf("usage: publish <topic> <json-payload> [--async]")
	}
	return c.Publish(rest[0], rest[1], *async)
}

func cmdStatus(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: status <event-id>")
	}
	return c.EventStatus(args[0])
}
