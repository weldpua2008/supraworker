package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	// "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"html/template"
	"io/ioutil"
	"net/http"
	// "strings"
	"time"
	// "runtime"

	config "github.com/weldpua2008/supraworker/config"
)

// GetNewJobs fetch from your API the jobs for execution
func NewRemoteApiRequest(ctx context.Context, section string, method string, url string) (error, []map[string]interface{}) {

	// localctx, cancel := context.WithCancel(ctx)
	// defer cancel()
	var rawResponseArray []map[string]interface{}
	var rawResponse map[string]interface{}

	// c := NewApiJobRequest()
	t := viper.GetStringMapString(section)
	c := make(map[string]string)
	for k, v := range t {
		var tpl_bytes bytes.Buffer
		tpl := template.Must(template.New("params").Parse(v))
		err := tpl.Execute(&tpl_bytes, config.C)
		if err != nil {
			log.Warn("executing template:", err)
		}
		c[k] = tpl_bytes.String()
		// log.Info(fmt.Sprintf("%s -> %s\n", k, tpl_bytes.String()))
	}
	var req *http.Request
	var err error

	if len(c) > 0 {
		jsonStr, err := json.Marshal(&c)

		if err != nil {
			return fmt.Errorf("Failed to marshal request due %s", err), nil
		}
		log.Trace(fmt.Sprintf("New Job request %s  to %s \nwith %s", method, url, jsonStr))
		// req, err = http.NewRequestWithContext(localctx,
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonStr))

	} else {
		// req, err = http.NewRequestWithContext(localctx,
		req, err = http.NewRequest(method, url, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: time.Duration(15 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request due %s", err), nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error read response body got %s", err), nil
	}
	err = json.Unmarshal(body, &rawResponseArray)
	if err != nil {
		err = json.Unmarshal(body, &rawResponse)
		if err != nil {
			return fmt.Errorf("error Unmarshal response: %s due %s", body, err), nil
		}
		rawResponseArray = append(rawResponseArray, rawResponse)
	}
	return nil, rawResponseArray

}
