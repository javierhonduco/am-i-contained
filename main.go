package main

import "C"
import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	bpf "github.com/aquasecurity/libbpfgo"
	"github.com/aquasecurity/libbpfgo/helpers"
)

var a int

//go:noinline
func testFunction() int {
	a += 1
	if a == 1 {
		return 3
	}
	return a
}

func main() {
	binaryPath := "/proc/self/exe"
	symbolName := "main.testFunction"

	_, err := os.Stat(binaryPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	bpfModule, err := bpf.NewModuleFromFile("main.bpf.o")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	defer bpfModule.Close()

	bpfModule.BPFLoadObject()
	prog, err := bpfModule.GetProgram("uprobe__test_function")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	offset, err := helpers.SymbolToOffset(binaryPath, symbolName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	_, err = prog.AttachUprobe(-1, binaryPath, offset)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	eventsChannel := make(chan []byte)
	lostChannel := make(chan uint64)

	perfBuf, err := bpfModule.InitPerfBuf("events", eventsChannel, lostChannel, 64)
	if err != nil {
		panic(fmt.Errorf("failed to init perf buffer: %w", err))
	}
	perfBuf.Poll(int(300))

	go func() {
		for {
			// Side-effect to avoid the compiler to optimise it out.
			fmt.Fprintf(io.Discard, "number of foo: %d", testFunction())

		}
	}()

	for {
		b := <-eventsChannel
		val := int(binary.LittleEndian.Uint32(b))

		fmt.Println("value", val, "current PID", os.Getpid())
		if val != os.Getpid() {
			panic("oh no")
		}
	}

	perfBuf.Stop()
	perfBuf.Close()

}
