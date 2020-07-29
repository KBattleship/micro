// +build integration

package test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestServerModeCall(t *testing.T) {
	trySuite(t, testServerModeCall, retryCount)
}

func testServerModeCall(t *t) {
	t.Parallel()
	serv := newServer(t, withLogin())

	callCmd := exec.Command("micro", serv.envFlag(), "call", "go.micro.runtime", "Runtime.Read", "{}")
	outp, err := callCmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Call to server should fail, got no error, output: %v", string(outp))
		return
	}

	defer serv.close()
	if err := serv.launch(); err != nil {
		return
	}

	if err := try("Calling Runtime.Read", t, func() ([]byte, error) {
		outp, err = exec.Command("micro", serv.envFlag(), "call", "go.micro.runtime", "Runtime.Read", "{}").CombinedOutput()
		if err != nil {
			return outp, errors.New("Call to runtime read should succeed")
		}
		return outp, err
	}, 5*time.Second); err != nil {
		return
	}
}

func TestRunLocalSource(t *testing.T) {
	trySuite(t, testRunLocalSource, retryCount)
}

func testRunLocalSource(t *t) {
	t.Parallel()
	serv := newServer(t, withLogin())
	defer serv.close()
	if err := serv.launch(); err != nil {
		return
	}

	runCmd := exec.Command("micro", serv.envFlag(), "run", "./service/example")
	outp, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro run failure, output: %v", string(outp))
		return
	}

	if err := try("Find test/example", t, func() ([]byte, error) {
		psCmd := exec.Command("micro", serv.envFlag(), "status")
		outp, err = psCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}

		// The started service should have the runtime name of "service/example",
		// as the runtime name is the relative path inside a repo.
		if !statusRunning("service/example", outp) {
			return outp, errors.New("Can't find example service in runtime")
		}
		return outp, err
	}, 15*time.Second); err != nil {
		return
	}

	if err := try("Find go.micro.service.example in list", t, func() ([]byte, error) {
		outp, err := exec.Command("micro", serv.envFlag(), "services").CombinedOutput()
		if err != nil {
			return outp, err
		}
		if !strings.Contains(string(outp), "go.micro.service.example") {
			return outp, errors.New("Can't find example service in list")
		}
		return outp, err
	}, 50*time.Second); err != nil {
		return
	}
}

func TestRunAndKill(t *testing.T) {
	trySuite(t, testRunAndKill, retryCount)
}

func testRunAndKill(t *t) {
	t.Parallel()
	serv := newServer(t, withLogin())
	defer serv.close()
	if err := serv.launch(); err != nil {
		return
	}

	runCmd := exec.Command("micro", serv.envFlag(), "run", "./service/example")
	outp, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro run failure, output: %v", string(outp))
		return
	}

	if err := try("Find test/example", t, func() ([]byte, error) {
		psCmd := exec.Command("micro", serv.envFlag(), "status")
		outp, err = psCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}

		// The started service should have the runtime name of "service/example",
		// as the runtime name is the relative path inside a repo.
		if !statusRunning("service/example", outp) {
			return outp, errors.New("Can't find example service in runtime")
		}
		return outp, err
	}, 15*time.Second); err != nil {
		return
	}

	if err := try("Find go.micro.service.example in list", t, func() ([]byte, error) {
		outp, err := exec.Command("micro", serv.envFlag(), "services").CombinedOutput()
		if err != nil {
			return outp, err
		}
		if !strings.Contains(string(outp), "go.micro.service.example") {
			return outp, errors.New("Can't find example service in list")
		}
		return outp, err
	}, 50*time.Second); err != nil {
		return
	}

	outp, err = exec.Command("micro", serv.envFlag(), "kill", "service/example").CombinedOutput()
	if err != nil {
		t.Fatalf("micro kill failure, output: %v", string(outp))
		return
	}

	if err := try("Find test/example", t, func() ([]byte, error) {
		psCmd := exec.Command("micro", serv.envFlag(), "status")
		outp, err = psCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}

		// The started service should have the runtime name of "service/example",
		// as the runtime name is the relative path inside a repo.
		if strings.Contains(string(outp), "service/example") {
			return outp, errors.New("Should not find example service in runtime")
		}
		return outp, err
	}, 15*time.Second); err != nil {
		return
	}

	if err := try("Find go.micro.service.example in list", t, func() ([]byte, error) {
		outp, err := exec.Command("micro", serv.envFlag(), "services").CombinedOutput()
		if err != nil {
			return outp, err
		}
		if strings.Contains(string(outp), "go.micro.service.example") {
			return outp, errors.New("Should not find example service in list")
		}
		return outp, err
	}, 20*time.Second); err != nil {
		return
	}
}

func statusRunning(service string, statusOutput []byte) bool {
	reg, _ := regexp.Compile(service + "\\s+latest\\s+\\S+\\s+running")

	return reg.Match(statusOutput)
}

func TestRunGithubSource(t *testing.T) {
	trySuite(t, testRunGithubSource, retryCount)
}

func testRunGithubSource(t *t) {
	t.Parallel()
	p, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(p) == 0 {
		t.Fatal("Git is not available")
		return
	}
	serv := newServer(t, withLogin())
	defer serv.close()
	if err := serv.launch(); err != nil {
		return
	}

	runCmd := exec.Command("micro", serv.envFlag(), "run", "github.com/micro/services/helloworld")
	outp, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro run failure, output: %v", string(outp))
		return
	}

	if err := try("Find hello world", t, func() ([]byte, error) {
		psCmd := exec.Command("micro", serv.envFlag(), "status")
		outp, err = psCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}

		if !statusRunning("helloworld", outp) {
			return outp, errors.New("Output should contain hello world")
		}
		return outp, nil
	}, 60*time.Second); err != nil {
		return
	}

	if err := try("Call hello world", t, func() ([]byte, error) {
		callCmd := exec.Command("micro", serv.envFlag(), "call", "helloworld", "Helloworld.Call", `{"name": "Joe"}`)
		outp, err := callCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}
		rsp := map[string]string{}
		err = json.Unmarshal(outp, &rsp)
		if err != nil {
			return outp, err
		}
		if rsp["msg"] != "Hello Joe" {
			return outp, errors.New("Helloworld resonse is unexpected")
		}
		return outp, err
	}, 60*time.Second); err != nil {
		return
	}

}

func TestRunLocalUpdateAndCall(t *testing.T) {
	trySuite(t, testRunLocalUpdateAndCall, retryCount)
}

func testRunLocalUpdateAndCall(t *t) {
	t.Parallel()
	serv := newServer(t, withLogin())
	defer serv.close()
	if err := serv.launch(); err != nil {
		return
	}

	// Run the example service
	runCmd := exec.Command("micro", serv.envFlag(), "run", "./service/example")
	outp, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro run failure, output: %v", string(outp))
		return
	}

	if err := try("Finding example service with micro status", t, func() ([]byte, error) {
		psCmd := exec.Command("micro", serv.envFlag(), "status")
		outp, err = psCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}

		// The started service should have the runtime name of "service/example",
		// as the runtime name is the relative path inside a repo.
		if !statusRunning("service/example", outp) {
			return outp, errors.New("can't find service in runtime")
		}
		return outp, err
	}, 15*time.Second); err != nil {
		return
	}

	if err := try("Call example service", t, func() ([]byte, error) {
		callCmd := exec.Command("micro", serv.envFlag(), "call", "go.micro.service.example", "Example.Call", `{"name": "Joe"}`)
		outp, err := callCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}
		rsp := map[string]string{}
		err = json.Unmarshal(outp, &rsp)
		if err != nil {
			return outp, err
		}
		if rsp["msg"] != "Hello Joe" {
			return outp, errors.New("Response is unexpected")
		}
		return outp, err
	}, 50*time.Second); err != nil {
		return
	}

	replaceStringInFile(t, "./service/example/handler/handler.go", "Hello", "Hi")
	defer func() {
		// Change file back
		replaceStringInFile(t, "./service/example/handler/handler.go", "Hi", "Hello")
	}()

	updateCmd := exec.Command("micro", serv.envFlag(), "update", "./service/example")
	outp, err = updateCmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := try("Call example service after modification", t, func() ([]byte, error) {
		callCmd := exec.Command("micro", serv.envFlag(), "call", "go.micro.service.example", "Example.Call", `{"name": "Joe"}`)
		outp, err = callCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}
		rsp := map[string]string{}
		err = json.Unmarshal(outp, &rsp)
		if err != nil {
			return outp, err
		}
		if rsp["msg"] != "Hi Joe" {
			return outp, errors.New("Response is not what's expected")
		}
		return outp, err
	}, 15*time.Second); err != nil {
		return
	}
}

func TestExistingLogs(t *testing.T) {
	trySuite(t, testExistingLogs, retryCount)
}

func testExistingLogs(t *t) {
	t.Parallel()
	serv := newServer(t, withLogin())
	defer serv.close()
	if err := serv.launch(); err != nil {
		return
	}

	runCmd := exec.Command("micro", serv.envFlag(), "run", "./service/logger")
	outp, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro run failure, output: %v", string(outp))
		return
	}

	if err := try("logger logs", t, func() ([]byte, error) {
		psCmd := exec.Command("micro", serv.envFlag(), "logs", "test/service/logger")
		outp, err = psCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}

		if !strings.Contains(string(outp), "Listening on") || !strings.Contains(string(outp), "never stopping") {
			return outp, errors.New("Output does not contain expected")
		}
		return outp, nil
	}, 50*time.Second); err != nil {
		return
	}
}

func TestBranchCheckout(t *testing.T) {
	trySuite(t, testBranchCheckout, retryCount)
}

func testBranchCheckout(t *t) {
	t.Parallel()
	serv := newServer(t, withLogin())
	defer serv.close()
	if err := serv.launch(); err != nil {
		return
	}

	runCmd := exec.Command("micro", serv.envFlag(), "run", "github.com/micro/micro/test/service/logger@master")
	outp, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro run failure, output: %v", string(outp))
		return
	}

	if err := try("logger logs", t, func() ([]byte, error) {
		psCmd := exec.Command("micro", serv.envFlag(), "logs", "micro/micro/test/service/logger")
		outp, err = psCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}

		// The log that this branch outputs is different from master, that's what we look for
		if !strings.Contains(string(outp), "Listening on") {
			return outp, errors.New("Output does not contain expected")
		}
		return outp, nil
	}, 30*time.Second); err != nil {
		return
	}
}

func TestStreamLogsAndThirdPartyRepo(t *testing.T) {
	trySuite(t, testStreamLogsAndThirdPartyRepo, retryCount)
}

func testStreamLogsAndThirdPartyRepo(t *t) {
	t.Parallel()
	serv := newServer(t, withLogin())
	defer serv.close()
	if err := serv.launch(); err != nil {
		return
	}

	runCmd := exec.Command("micro", serv.envFlag(), "run", "github.com/micro/micro/test/service/logger")
	outp, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro run failure, output: %v", string(outp))
		return
	}

	if err := try("logger logs", t, func() ([]byte, error) {
		psCmd := exec.Command("micro", serv.envFlag(), "logs", "micro/micro/test/service/logger")
		outp, err = psCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}

		if !strings.Contains(string(outp), "Listening on") || !strings.Contains(string(outp), "never stopping") {
			return outp, errors.New("Output does not contain expected")
		}
		return outp, nil
	}, 50*time.Second); err != nil {
		return
	}

	// Test streaming logs
	cmd := exec.Command("micro", serv.envFlag(), "logs", "-n", "1", "-f", "micro-test-service-logger")

	time.Sleep(7 * time.Second)

	go func() {
		outp, err := cmd.CombinedOutput()
		if err != nil {
			t.Log(err)
		}
		if len(outp) == 0 {
			t.Fatal("No log lines streamed")
			return
		}
		if !strings.Contains(string(outp), "never stopping") {
			t.Fatalf("Unexpected logs: %v", string(outp))
			return
		}
		// Logspammer logs every 2 seconds, so we need 2 different
		now := time.Now()
		// leaving the hour here to fix a docker issue
		// when the containers clock is a few hours behind
		stampA := now.Add(-2 * time.Second).Format("04:05")
		stampB := now.Add(-1 * time.Second).Format("04:05")
		if !strings.Contains(string(outp), stampA) && !strings.Contains(string(outp), stampB) {
			t.Fatalf("Timestamp %v or %v not found in logs: %v", stampA, stampB, string(outp))
			return
		}
	}()

	time.Sleep(7 * time.Second)

	err = cmd.Process.Kill()
	if err != nil {
		t.Fatal(err)
		return
	}
	time.Sleep(2 * time.Second)
}

func replaceStringInFile(t *t, filepath string, original, newone string) {
	input, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
		return
	}

	output := strings.ReplaceAll(string(input), original, newone)
	err = ioutil.WriteFile(filepath, []byte(output), 0644)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestParentDependency(t *testing.T) {
	trySuite(t, testParentDependency, retryCount)
}

func testParentDependency(t *t) {
	t.Parallel()
	serv := newServer(t, withLogin())
	defer serv.close()
	if err := serv.launch(); err != nil {
		return
	}

	runCmd := exec.Command("micro", serv.envFlag(), "run", "./dep-test/dep-test-service")
	outp, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro run failure, output: %v", string(outp))
		return
	}

	if err := try("Find hello world", t, func() ([]byte, error) {
		psCmd := exec.Command("micro", serv.envFlag(), "status")
		outp, err = psCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}

		if !statusRunning("dep-test-service", outp) {
			return outp, errors.New("Output should contain hello world")
		}
		return outp, nil
	}, 30*time.Second); err != nil {
		return
	}
}

func TestFastRuns(t *testing.T) {
	trySuite(t, testFastRuns, retryCount)
}

func testFastRuns(t *t) {
	t.Parallel()
	serv := newServer(t, withLogin())
	defer serv.close()
	if err := serv.launch(); err != nil {
		return
	}

	runCmd := exec.Command("micro", serv.envFlag(), "run", "github.com/m3o/services/signup")
	outp, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro run failure, output: %v", string(outp))
		return
	}

	// Stripe needs some configs to start
	// TODO replace with normal micro config set call
	runCmd = exec.Command("micro", serv.envFlag(), "call", "go.micro.config", "Config.Create", `{"change":{"namespace":"micro", "path":"micro.payments.stripe.api_key", "changeSet" : {"data":"notatruekey", "format":"json"}}}`)
	// runCmd = exec.Command("micro", serv.envFlag(), "config", "set", "micro.payments.stripe.api_key", "notatruekey")
	outp, err = runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro config set failure, output: %v", string(outp))
		return
	}

	runCmd = exec.Command("micro", serv.envFlag(), "run", "github.com/m3o/services/payments/provider/stripe")
	outp, err = runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("micro run failure, output: %v", string(outp))
		return
	}

	if err := try("Find signup and stripe", t, func() ([]byte, error) {
		psCmd := exec.Command("micro", serv.envFlag(), "services")
		outp, err = psCmd.CombinedOutput()
		if err != nil {
			return outp, err
		}

		if !strings.Contains(string(outp), "signup") || !strings.Contains(string(outp), "stripe") {
			return outp, errors.New("Signup or stripe can't be found")
		}
		return outp, nil
	}, 120*time.Second); err != nil {
		return
	}
}
