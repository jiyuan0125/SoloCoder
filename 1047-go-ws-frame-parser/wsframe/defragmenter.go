package wsframe

type Message struct {
	OpCode  int
	Payload []byte
}

type Defragmenter struct {
	maxMessageSize int
	currentMessage *Message
	haveFirstFrame bool
}

func NewDefragmenter(maxMessageSize int) *Defragmenter {
	if maxMessageSize <= 0 {
		maxMessageSize = DefaultMaxMessageSize
	}
	return &Defragmenter{
		maxMessageSize: maxMessageSize,
	}
}

func (d *Defragmenter) ProcessFrame(frame *Frame) (*Message, error) {
	if frame.IsControl() {
		if frame.OpCode == OpCodeClose {
			return &Message{
				OpCode:  OpCodeClose,
				Payload: frame.Payload,
			}, nil
		}
		return &Message{
			OpCode:  frame.OpCode,
			Payload: frame.Payload,
		}, nil
	}

	if frame.OpCode == OpCodeContinuation {
		if !d.haveFirstFrame {
			return nil, ErrUnexpectedContinuation
		}
	} else {
		if d.haveFirstFrame {
			return nil, ErrMissingContinuation
		}
		d.currentMessage = &Message{
			OpCode:  frame.OpCode,
			Payload: make([]byte, 0),
		}
		d.haveFirstFrame = true
	}

	newSize := len(d.currentMessage.Payload) + len(frame.Payload)
	if newSize > d.maxMessageSize {
		d.Reset()
		return nil, ErrMessageTooLarge
	}

	d.currentMessage.Payload = append(d.currentMessage.Payload, frame.Payload...)

	if frame.Fin {
		msg := d.currentMessage
		d.Reset()
		return msg, nil
	}

	return nil, nil
}

func (d *Defragmenter) Reset() {
	d.currentMessage = nil
	d.haveFirstFrame = false
}

func (d *Defragmenter) IsFragmented() bool {
	return d.haveFirstFrame
}
