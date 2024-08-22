package zilliz

import (
	"fmt"
	"testing"

	"start-feishubot/initialization"
)

func TestRetrieve(t *testing.T) {
	config := initialization.LoadConfig("../../config.yaml")

	zilliz := NewZillizPipelineClient(*config)
	resp, err := zilliz.Retrieve("what is vllm", 3)
	if err != nil {
		t.Errorf("TestRetrieve failed with error: %v", err)
	}
	for idx, chunk := range resp {
		fmt.Println(idx)
		fmt.Println(chunk)
	}
}
