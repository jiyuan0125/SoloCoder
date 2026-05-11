package zorder

const (
	max2DBits = 32
	max3DBits = 21
)

type Code uint64

type Point2D struct {
	X uint32
	Y uint32
}

type Point3D struct {
	X uint32
	Y uint32
	Z uint32
}

type Rect2D struct {
	Min Point2D
	Max Point2D
}

type Range struct {
	Start Code
	End   Code
}

func (c Code) MarshalJSON() ([]byte, error) {
	return []byte("\"" + uint64ToString(uint64(c)) + "\""), nil
}

func (c *Code) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return &ErrInvalidCodeFormat{}
	}
	val, err := stringToUint64(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*c = Code(val)
	return nil
}

func uint64ToString(n uint64) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append(buf, '0'+byte(n%10))
		n /= 10
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

func stringToUint64(s string) (uint64, error) {
	if s == "" {
		return 0, &ErrInvalidCodeFormat{}
	}
	var n uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, &ErrInvalidCodeFormat{}
		}
		d := uint64(c - '0')
		if n > (1<<64-1-d)/10 {
			return 0, &ErrOverflow{}
		}
		n = n*10 + d
	}
	return n, nil
}

type ErrInvalidCodeFormat struct{}

func (e *ErrInvalidCodeFormat) Error() string {
	return "invalid code format"
}

type ErrOverflow struct{}

func (e *ErrOverflow) Error() string {
	return "value overflows uint64"
}

type ErrCoordOverflow struct {
	Coord uint32
	Max   uint32
}

func (e *ErrCoordOverflow) Error() string {
	return "coordinate overflow"
}
