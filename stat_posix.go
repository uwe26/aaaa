// +build !plan9

package sftp

import (
	"os"
	"syscall"
)

const EBADF = syscall.EBADF

func wrapPathError(filepath string, err error) error {
	if errno, ok := err.(syscall.Errno); ok {
		return &os.PathError{Path: filepath, Err: errno}
	}
	return err
}

// translateErrno translates a syscall error number to a SFTP error code.
func translateErrno(errno syscall.Errno) uint32 {
	switch errno {
	case 0:
		return sshFxOk
	case syscall.ENOENT:
		return sshFxNoSuchFile
	case syscall.EPERM:
		return sshFxPermissionDenied
	}

	return sshFxFailure
}

func translateSyscallError(err error) (uint32, bool) {
	switch e := err.(type) {
	case syscall.Errno:
		return translateErrno(e), true
	case *os.PathError:
		debug("statusFromError,pathError: error is %T %#v", e.Err, e.Err)
		if errno, ok := e.Err.(syscall.Errno); ok {
			return translateErrno(errno), true
		}
	}
	return 0, false
}

// toFileMode converts sftp filemode bits to the os.FileMode specification
func toFileMode(mode uint32) os.FileMode {
	var fm = os.FileMode(mode & 0777)
	switch mode & S_IFMT {
	case syscall.S_IFBLK:
		fm |= os.ModeDevice
	case syscall.S_IFCHR:
		fm |= os.ModeDevice | os.ModeCharDevice
	case syscall.S_IFDIR:
		fm |= os.ModeDir
	case syscall.S_IFIFO:
		fm |= os.ModeNamedPipe
	case syscall.S_IFLNK:
		fm |= os.ModeSymlink
	case syscall.S_IFREG:
		// nothing to do
	case syscall.S_IFSOCK:
		fm |= os.ModeSocket
	}
	if mode&syscall.S_ISGID != 0 {
		fm |= os.ModeSetgid
	}
	if mode&syscall.S_ISUID != 0 {
		fm |= os.ModeSetuid
	}
	if mode&syscall.S_ISVTX != 0 {
		fm |= os.ModeSticky
	}
	return fm
}

// fromFileMode converts from the os.FileMode specification to sftp filemode bits
func fromFileMode(mode os.FileMode) uint32 {
	ret := uint32(0)

	if mode&os.ModeDevice != 0 {
		if mode&os.ModeCharDevice != 0 {
			ret |= syscall.S_IFCHR
		} else {
			ret |= syscall.S_IFBLK
		}
	}
	if mode&os.ModeDir != 0 {
		ret |= syscall.S_IFDIR
	}
	if mode&os.ModeSymlink != 0 {
		ret |= syscall.S_IFLNK
	}
	if mode&os.ModeNamedPipe != 0 {
		ret |= syscall.S_IFIFO
	}
	if mode&os.ModeSetgid != 0 {
		ret |= syscall.S_ISGID
	}
	if mode&os.ModeSetuid != 0 {
		ret |= syscall.S_ISUID
	}
	if mode&os.ModeSticky != 0 {
		ret |= syscall.S_ISVTX
	}
	if mode&os.ModeSocket != 0 {
		ret |= syscall.S_IFSOCK
	}

	if mode&os.ModeType == 0 {
		ret |= syscall.S_IFREG
	}
	ret |= uint32(mode & os.ModePerm)

	return ret
}

const (
	s_ISUID = syscall.S_ISUID
	s_ISGID = syscall.S_ISGID
	s_ISVTX = syscall.S_ISVTX
)
