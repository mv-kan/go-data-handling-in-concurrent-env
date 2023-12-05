package example_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/mv-kan/go-data-handling-in-concurrent-env/pkg/example"
	"github.com/stretchr/testify/assert"
)

func TestCarExample(t *testing.T) {
	logger := log.New(os.Stdout, "CarExample: ", log.Ltime)
	done := make(chan struct{})
	example.RunCarExample(logger, done)
	time.Sleep(time.Second * 20)
	// exit from all go routines created by RunCarExample
	close(done)
	b, err := os.ReadFile("log.txt") // just pass the file name
	assert.Nil(t, err)
	str := string(b)
	assert.NotEmpty(t, str)
}
