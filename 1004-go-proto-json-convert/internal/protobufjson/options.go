package protobufjson

type MarshalOptions struct {
	EmitUnpopulated bool
}

type UnmarshalOptions struct {
	DiscardUnknown bool
}
