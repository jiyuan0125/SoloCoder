package main

import (
	"fmt"
	"os"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: generate <output.pb>")
		os.Exit(1)
	}

	outPath := os.Args[1]

	protoContent := `syntax = "proto3";

package test;

import "google/protobuf/timestamp.proto";

message Status {
  enum Code {
    UNKNOWN = 0;
    OK = 1;
    ERROR = 2;
  }
}

message InnerMessage {
  int32 value = 1;
  string text = 2;
}

message OneofTest {
  oneof content {
    int32 int_value = 1;
    string string_value = 2;
  }
}

message TestMessage {
  string name = 1;
  int32 id = 2;
  repeated string tags = 3;
  map<string, int32> scores = 4;
  InnerMessage nested = 5;
  google.protobuf.Timestamp created_at = 6;
  Status.Code status = 7;
  OneofTest oneof_field = 8;
}
`

	fmt.Println("Proto definition:")
	fmt.Println(protoContent)
	fmt.Println("\nNote: Run protoc to generate Go code, or use the CLI tool with a .pb file")

	ts := timestamppb.New(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
	tsBytes, _ := proto.Marshal(ts)
	fmt.Printf("\nTimestamp proto bytes (as reference): %v\n", tsBytes)
}
