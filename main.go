package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/operations"
	"github.com/go-openapi/strfmt"
	"github.com/sirupsen/logrus"
)

func main() {

	callAPI()
	callAPI()
	callAPI()

	logrus.Info("Sleep a bit")

	time.Sleep(time.Second)

	getConnectionListFromSystem(os.Getpid())
}

func callAPI() {
	fcClient := client.NewHTTPClient(strfmt.NewFormats())
	logger := logrus.NewEntry(logrus.New())

	socketPath := "/tmp/firecracker.socket"
	transport := firecracker.NewUnixSocketTransport(
		socketPath,
		logger,
		true,
	)

	fcClient.SetTransport(transport)

	resp, err := fcClient.Operations.DescribeInstance(
		operations.NewDescribeInstanceParams(),
	)
	if err != nil {
		logrus.Error(err.Error())

		return
	}

	logrus.
		WithField("state", *resp.Payload.State).
		Info("Firecracker API response")
}

func getConnectionListFromSystem(pid int) {
	logger := logrus.WithField("pid", os.Getpid())

	logger.Info("Check connections")

	connections := map[string]int{}
	removeRepeatingSpace := regexp.MustCompile(" +")
	procPath := fmt.Sprintf("/proc/%d/fd", pid)

	entries, err := os.ReadDir(procPath)
	if err != nil {
		logrus.Error(err.Error())

		return
	}

	for _, file := range entries {
		if file.Type() != os.ModeSymlink {
			continue
		}

		link, err := os.Readlink(path.Join(procPath, file.Name()))
		if err != nil {
			continue
		}

		if !strings.HasPrefix(link, "socket:") {
			continue
		}

		sockID := link[8 : len(link)-1]
		loggerPerSocket := logger.WithField("unixgram", sockID)

		var out bytes.Buffer

		cmd := exec.Command("ss", "--no-header", "dst", fmt.Sprintf(":%s", sockID))
		cmd.Stdout = &out

		runErr := cmd.Run()
		if runErr != nil {
			loggerPerSocket.Error(runErr.Error())
		}

		result := removeRepeatingSpace.ReplaceAllString(out.String(), " ")
		parts := strings.Split(result, " ")

		connections[parts[4]] += 1
	}

	for target, count := range connections {
		logger.
			WithFields(
				logrus.Fields{"target": target, "count": count},
			).
			Infof("Open connection")
	}
}
