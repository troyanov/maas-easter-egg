package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

//go:embed ipmipower
var ipmipower []byte

//go:embed libfreeipmi.so.17
var libfreeipmi []byte

//go:embed libipmidetect.so.0
var libipmidetect []byte

func main() {
	memfd1, _ := MemfdCreate("ipmipower")
	memfd2, _ := MemfdCreate("/usr/lib/x86_64-linux-gnu/libipmidetect.so.0")
	memfd3, _ := MemfdCreate("/usr/lib/x86_64-linux-gnu/libfreeipmi.so.17")

	// Copy the binary file's content into the memory file
	_ = CopyToMem(memfd1, ipmipower)
	_ = CopyToMem(memfd2, libipmidetect)
	_ = CopyToMem(memfd3, libfreeipmi)

	// Execute the binary from memory
	memfdPath1 := fmt.Sprintf("/proc/self/fd/%d", memfd1)
	memfdPath2 := fmt.Sprintf("/proc/self/fd/%d", memfd2)
	memfdPath3 := fmt.Sprintf("/proc/self/fd/%d", memfd3)

	cmd := exec.Command(memfdPath1, "--help")
	cmd.Env = append(cmd.Env, fmt.Sprintf("LD_PRELOAD=%s %s", memfdPath2, memfdPath3))
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running embedded binary:", err)
		return
	}

	fmt.Println("out:", outb.String(), "err:", errb.String())
}

func MemfdCreate(path string) (uintptr, error) {
	s, err := syscall.BytePtrFromString(path)
	if err != nil {
		return 0, err
	}

	r1, _, err := syscall.Syscall(319, uintptr(unsafe.Pointer(s)), 0, 0)

	if int(r1) == -1 {
		return r1, err
	}

	return r1, nil
}

func CopyToMem(fd uintptr, buf []byte) (err error) {
	_, err = syscall.Write(int(fd), buf)
	if err != nil {
		return err
	}

	return nil
}
