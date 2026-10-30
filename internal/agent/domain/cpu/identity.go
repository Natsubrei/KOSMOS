//go:build linux

package cpu

import (
	"bytes"
	"os"
	"strconv"
)

func taskIdentity(tid uint32) (pid int32, starttime uint64, ok bool) {
	path := "/proc/" + strconv.FormatUint(uint64(tid), 10)
	stat, err := os.ReadFile(path + "/stat")
	if err != nil {
		return 0, 0, false
	}
	starttime, ok = statStarttime(stat)
	if !ok {
		return 0, 0, false
	}
	st, err := os.ReadFile(path + "/status")
	if err != nil {
		return 0, 0, false
	}
	pid, ok = statusTgid(st)
	return pid, starttime, ok
}

func statStarttime(stat []byte) (uint64, bool) {
	i := bytes.LastIndexByte(stat, ')')
	if i < 0 {
		return 0, false
	}
	fields := bytes.Fields(stat[i+1:])
	// /proc/<pid>/stat field 22 is starttime; fields after comm begin at field 3.
	if len(fields) < 20 {
		return 0, false
	}
	n, err := strconv.ParseUint(string(fields[19]), 10, 64)
	return n, err == nil
}

func statusTgid(status []byte) (int32, bool) {
	for _, line := range bytes.Split(status, []byte{'\n'}) {
		if !bytes.HasPrefix(line, []byte("Tgid:")) {
			continue
		}
		n, err := strconv.ParseInt(string(bytes.TrimSpace(line[len("Tgid:"):])), 10, 32)
		return int32(n), err == nil && n > 0
	}
	return 0, false
}
