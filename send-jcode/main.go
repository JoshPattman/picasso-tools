package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JoshPattman/jcode"
	"github.com/tarm/serial"
)

func main() {
	// Parse flags
	port := flag.String("port", "/dev/ttyUSB0", "The serial port to connect to the JCode hardware on")
	baud := flag.Int("baud", 115200, "The baud rate to use to cimmunicate with the JCode hardware")
	inputFile := flag.String("input", "", "The input JCode file, required")
	bufferSize := flag.Int("buffer", 4, "How many instructions should there be with the robot at once")
	flag.Parse()

	if *inputFile == "" {
		failE(errors.New("must specify an input file"))
	}

	// Read and parse code file
	codeFile, err := os.Open(*inputFile)
	if err != nil {
		failE(errors.Join(errors.New("could not open input file"), err))
	}
	defer codeFile.Close()
	dec := jcode.NewDecoder(codeFile)
	queue := make([]jcode.Instruction, 0)
	for {
		ins, err := dec.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			failE(errors.Join(errors.New("could not parse input file"), err))
		}
		queue = append(queue, ins)
	}
	codeFile.Close()
	fmt.Println("Read input file")

	// Connect to device (and wait for a bit to connect)
	serialPort, err := serial.OpenPort(&serial.Config{
		Name: *port,
		Baud: *baud,
	})
	if err != nil {
		failE(err)
	}
	defer serialPort.Close()
	fmt.Println("Connected to device, waiting for init")
	time.Sleep(time.Second * 2)
	fmt.Println("Device ready, beginning streaming")

	// Set up quit
	go func() {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		<-ctx.Done()
		serialPort.Close()
		failE(errors.New("user interrupted streaming, gracefully closed port"))
	}()

	// Stream data
	currentBuffer := 0
	robotEnc := jcode.NewEncoder(serialPort)
	robotDec := jcode.NewDecoder(serialPort)
	remainingN := len(queue)
	completedN := 0
	for len(queue) > 0 {
		if currentBuffer < *bufferSize {
			robotEnc.Write(queue[0])
			queue = queue[1:]
			currentBuffer += 1
		} else {
			ins, _ := robotDec.Read()
			switch ins := ins.(type) {
			case jcode.Consumed:
				currentBuffer -= 1
				remainingN -= 1
				completedN += 1
				fmt.Printf("\r%.2f%%        ", 100-float64(remainingN*100)/float64(completedN+remainingN))
			case jcode.Log:
				fmt.Printf("\r> %s     ", ins.Message)
				fmt.Printf("\n%.2f%%        ", 100-float64(remainingN*100)/float64(completedN+remainingN))
			}
		}
	}
	fmt.Println("\nStreaming complete, waiting for final commands to complete")
	time.Sleep(time.Second * 5)
}

func failE(err error) {
	fmt.Println("Fatal:", err)
	os.Exit(1)
}
