package fdpass

import (
	"errors"
	"syscall"
)

var (
	ErrNoFDReceived = errors.New("no file descriptor received")
	ErrSendFailed   = errors.New("failed to send file descriptor")
	ErrRecvFailed   = errors.New("failed to receive file descriptor")
)

func PackControlMessage(fds []int) []byte {
	return syscall.UnixRights(fds...)
}

func ParseControlMessage(oob []byte) ([]int, error) {
	scms, err := syscall.ParseSocketControlMessage(oob)
	if err != nil {
		return nil, err
	}
	var fds []int
	for _, scm := range scms {
		parsed, err := syscall.ParseUnixRights(&scm)
		if err != nil {
			return nil, err
		}
		fds = append(fds, parsed...)
	}
	return fds, nil
}

func SendFD(socketFD int, fdToSend int) error {
	oob := PackControlMessage([]int{fdToSend})

	dummy := []byte{0}
	err := syscall.Sendmsg(socketFD, dummy, oob, nil, 0)
	if err != nil {
		return errors.New(ErrSendFailed.Error() + ": " + err.Error())
	}
	return nil
}

func ReceiveFD(socketFD int) (int, error) {
	oob := make([]byte, syscall.CmsgSpace(4))
	dummy := make([]byte, 1)

	_, oobn, _, _, err := syscall.Recvmsg(socketFD, dummy, oob, 0)
	if err != nil {
		return -1, errors.New(ErrRecvFailed.Error() + ": " + err.Error())
	}

	if oobn == 0 {
		return -1, ErrNoFDReceived
	}

	fds, err := ParseControlMessage(oob[:oobn])
	if err != nil {
		return -1, err
	}
	if len(fds) == 0 {
		return -1, ErrNoFDReceived
	}

	return fds[0], nil
}

func CreateSocketPair() (int, int, error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return -1, -1, err
	}
	return fds[0], fds[1], nil
}

func CloseFD(fd int) error {
	return syscall.Close(fd)
}
