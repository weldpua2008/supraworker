package job

import (
	"fmt"
	"github.com/weldpua2008/supraworker/model"
	"github.com/weldpua2008/supraworker/model/cmdtest"

	// "io/ioutil"
	// "net/http"
	// "net/http/httptest"
	// "os/exec"
	// "os"
	"testing"
	"time"
)

func init() {
	cmdtest.StartTrace()
}
func TestHelperProcess(t *testing.T) {
	cmdtest.TestHelperProcess(t)
}

func TestGenerateJobs(t *testing.T) {
	job := model.NewJob(fmt.Sprintf("job-%v", cmdtest.GetFunctionName(t.Name)), cmdtest.CMDForTest("echo GGGG&&exit 0"))

	job.StreamInterval = 1 * time.Millisecond

	err := job.Run()
	if err != nil {
		t.Errorf("Expected no error in %s, got %v", cmdtest.GetFunctionName(t.Name), err)
	}
}
