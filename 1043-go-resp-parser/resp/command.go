package resp

import (
	"strings"
)

type CommandProcessor struct {
	store *Store
}

func NewCommandProcessor() *CommandProcessor {
	return &CommandProcessor{
		store: NewStore(),
	}
}

func NewCommandProcessorWithStore(store *Store) *CommandProcessor {
	return &CommandProcessor{
		store: store,
	}
}

func (p *CommandProcessor) Process(args []string) Value {
	if len(args) == 0 {
		return NewError("ERR empty command")
	}

	cmd := strings.ToUpper(args[0])
	args = args[1:]

	switch cmd {
	case "SET":
		return p.cmdSet(args)
	case "GET":
		return p.cmdGet(args)
	case "DEL":
		return p.cmdDel(args)
	case "EXISTS":
		return p.cmdExists(args)
	case "INCR":
		return p.cmdIncr(args)
	case "EXPIRE":
		return p.cmdExpire(args)
	case "TTL":
		return p.cmdTTL(args)
	case "PING":
		return p.cmdPing(args)
	case "KEYS":
		return p.cmdKeys(args)
	case "COMMAND":
		return p.cmdCommand(args)
	default:
		return NewError("ERR unknown command '" + cmd + "'")
	}
}

func (p *CommandProcessor) ProcessBatch(commands [][]string) []Value {
	results := make([]Value, len(commands))
	for i, cmd := range commands {
		results[i] = p.Process(cmd)
	}
	return results
}

func (p *CommandProcessor) cmdSet(args []string) Value {
	if len(args) < 2 {
		return NewError("ERR wrong number of arguments for 'set' command")
	}
	key := args[0]
	value := args[1]
	p.store.Set(key, value)
	return NewSimpleString("OK")
}

func (p *CommandProcessor) cmdGet(args []string) Value {
	if len(args) < 1 {
		return NewError("ERR wrong number of arguments for 'get' command")
	}
	key := args[0]
	value, ok := p.store.Get(key)
	if !ok {
		return NewNullBulkString()
	}
	return NewBulkString(value)
}

func (p *CommandProcessor) cmdDel(args []string) Value {
	if len(args) < 1 {
		return NewError("ERR wrong number of arguments for 'del' command")
	}
	count := p.store.Del(args...)
	return NewInteger(int64(count))
}

func (p *CommandProcessor) cmdExists(args []string) Value {
	if len(args) < 1 {
		return NewError("ERR wrong number of arguments for 'exists' command")
	}
	count := p.store.Exists(args...)
	return NewInteger(int64(count))
}

func (p *CommandProcessor) cmdIncr(args []string) Value {
	if len(args) < 1 {
		return NewError("ERR wrong number of arguments for 'incr' command")
	}
	key := args[0]
	val, err := p.store.Incr(key)
	if err != nil {
		return NewError("ERR value is not an integer or out of range")
	}
	return NewInteger(val)
}

func (p *CommandProcessor) cmdExpire(args []string) Value {
	if len(args) < 2 {
		return NewError("ERR wrong number of arguments for 'expire' command")
	}
	key := args[0]
	seconds, err := parseInt64(args[1])
	if err != nil {
		return NewError("ERR value is not an integer or out of range")
	}
	result := p.store.Expire(key, seconds)
	return NewInteger(int64(result))
}

func (p *CommandProcessor) cmdTTL(args []string) Value {
	if len(args) < 1 {
		return NewError("ERR wrong number of arguments for 'ttl' command")
	}
	key := args[0]
	ttl := p.store.TTL(key)
	return NewInteger(ttl)
}

func (p *CommandProcessor) cmdPing(args []string) Value {
	if len(args) == 0 {
		return NewSimpleString("PONG")
	}
	return NewBulkString(args[0])
}

func (p *CommandProcessor) cmdKeys(args []string) Value {
	keys := p.store.GetAllKeys()
	arr := make([]Value, len(keys))
	for i, k := range keys {
		arr[i] = NewBulkString(k)
	}
	return NewArray(arr)
}

func (p *CommandProcessor) cmdCommand(args []string) Value {
	return NewSimpleString("OK")
}
