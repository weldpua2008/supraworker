package cmdtest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
	// "github.com/sirupsen/logrus"
)

type execFunc func(command string, args ...string) *exec.Cmd

func getFakeExecCommand(validator func(string, ...string)) execFunc {
	return func(command string, args ...string) *exec.Cmd {
		validator(command, args...)
		return FakeExecCommand(command, args...)
	}
}

func FakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func FakeExecCommandContext(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		// fmt.Fprintf(os.Stdout, fmt.Sprintf("os.Args '%v'",os.Args))
		return
	}
	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build…/…/_test/job_test.test -test.run=TestHelperProcess --
	// cmd, args := args[3], args[4:]
	cmd := args[3]

	re := regexp.MustCompile(`echo (.+?)`)
	// Handle the case where args[0] is dir:...
	// TODO: Futher Validation
	if !strings.Contains(cmd, "bash") {
		fmt.Fprintf(os.Stderr, "Expected command to be 'bash'. Got: '%s' %s", cmd, args)
		os.Exit(2)
	}

	exitCode := 127
	switch {
	case strings.Contains(strings.Join(args, " "), "exit 0"):
		exitCode = 0
	}
	// some code here to check arguments perhaps?
	switch {
	case strings.Contains(strings.Join(args, " "), "sleep "):
        // fmt.Fprintf(os.Stdout, "Sleep for 10 seconds")
		time.Sleep(10 * time.Second)
	}
	res := re.FindStringSubmatch(strings.Join(args, " "))
	out := ""
	if len(res) > 1 {
		out = res[1]
	}

	if (len(out) > 0) && (out != string(' ')) {
		fmt.Fprintf(os.Stdout, fmt.Sprintf("'%v'", out))
	}

	os.Exit(exitCode)
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
