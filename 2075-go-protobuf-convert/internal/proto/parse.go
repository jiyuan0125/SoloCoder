package proto

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type FileInfo struct {
	Path        string
	Messages    map[string]protoreflect.MessageDescriptor
	Enums       map[string]protoreflect.EnumDescriptor
	Desc        *desc.FileDescriptor
	rawMessages map[string]*desc.MessageDescriptor
}

type ParseError struct {
	Line    int
	Column  int
	Message string
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d:%d: %s", e.Line, e.Column, e.Message)
	}
	return e.Message
}

func ParseFile(path string) (*FileInfo, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New("proto file not found: " + path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	parser := protoparse.Parser{
		ImportPaths: []string{filepath.Dir(absPath), ".", "/"},
	}

	filedescs, err := parser.ParseFiles(filepath.Base(absPath))
	if err != nil {
		errStr := err.Error()
		line := 0
		column := 0

		parts := strings.Split(errStr, ":")
		if len(parts) >= 3 {
			fmt.Sscanf(parts[1], "%d", &line)
			fmt.Sscanf(parts[2], "%d", &column)
		}

		return nil, &ParseError{
			Line:    line,
			Column:  column,
			Message: errStr,
		}
	}

	if len(filedescs) == 0 {
		return nil, errors.New("no file descriptors found")
	}

	fileDesc := filedescs[0]

	messages := make(map[string]protoreflect.MessageDescriptor)
	enums := make(map[string]protoreflect.EnumDescriptor)
	rawMessages := make(map[string]*desc.MessageDescriptor)

	collectFromFile(fileDesc, messages, enums, rawMessages)

	return &FileInfo{
		Path:        path,
		Messages:    messages,
		Enums:       enums,
		Desc:        fileDesc,
		rawMessages: rawMessages,
	}, nil
}

func collectFromFile(fd *desc.FileDescriptor,
	messages map[string]protoreflect.MessageDescriptor,
	enums map[string]protoreflect.EnumDescriptor,
	rawMessages map[string]*desc.MessageDescriptor,
) {
	for _, md := range fd.GetMessageTypes() {
		collectMessage(md, messages, enums, rawMessages)
	}

	for _, ed := range fd.GetEnumTypes() {
		fullName := string(ed.GetFullyQualifiedName())
		enums[fullName] = ed.UnwrapEnum()
	}
}

func collectMessage(md *desc.MessageDescriptor,
	messages map[string]protoreflect.MessageDescriptor,
	enums map[string]protoreflect.EnumDescriptor,
	rawMessages map[string]*desc.MessageDescriptor,
) {
	fullName := string(md.GetFullyQualifiedName())
	messages[fullName] = md.UnwrapMessage()
	rawMessages[fullName] = md

	for _, nested := range md.GetNestedMessageTypes() {
		collectMessage(nested, messages, enums, rawMessages)
	}

	for _, nestedEnum := range md.GetNestedEnumTypes() {
		nestedEnumFull := string(nestedEnum.GetFullyQualifiedName())
		enums[nestedEnumFull] = nestedEnum.UnwrapEnum()
	}
}

func (fi *FileInfo) GetMessageDescriptor(name string) (protoreflect.MessageDescriptor, error) {
	if md, ok := fi.Messages[name]; ok {
		return md, nil
	}

	for fullName, md := range fi.Messages {
		if strings.HasSuffix(fullName, "."+name) || fullName == name {
			return md, nil
		}
	}

	return nil, fmt.Errorf("message type '%s' not found in proto file", name)
}

func (fi *FileInfo) NewMessage(name string) (*dynamicpb.Message, error) {
	md, err := fi.GetMessageDescriptor(name)
	if err != nil {
		return nil, err
	}
	return dynamicpb.NewMessage(md), nil
}

func (fi *FileInfo) ListMessages() []string {
	names := make([]string, 0, len(fi.Messages))
	for name := range fi.Messages {
		names = append(names, name)
	}
	return names
}
