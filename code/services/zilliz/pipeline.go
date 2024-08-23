package zilliz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"start-feishubot/initialization"
	"start-feishubot/logger"
)

type QueryData struct {
	QueryText string `json:"query_text"`
}
type Params struct {
	Limit        int      `json:"limit"`
	Offset       int      `json:"offset"`
	OutputFields []string `json:"outputFields"`
}
type SearchRequest struct {
	Data   QueryData `json:"data"`
	Params Params    `json:"params"`
}

type Usage struct {
	Embedding int `json:"embedding"`
	Rerank    int `json:"rerank"`
}

type Result struct {
	ID        int64   `json:"id"`
	Distance  float64 `json:"distance"`
	DocName   string  `json:"doc_name"`
	ChunkID   int     `json:"chunk_id"`
	ChunkText string  `json:"chunk_text"`
}

type ResultData struct {
	Result []Result `json:"result"`
	Usage  Usage    `json:"usage"`
}

type SearchResponse struct {
	Code    int        `json:"code"`
	Data    ResultData `json:"data"`
	Message string     `json:"message"`
}

type ZillizPipelineClient struct {
	APIKey     string
	Region     string
	PipelineID string
}

func NewZillizPipelineClient(config initialization.Config) *ZillizPipelineClient {
	return &ZillizPipelineClient{
		Region:     config.ZillizRegion,
		APIKey:     config.ZillizAPIKey,
		PipelineID: config.ZillizPipelineID,
	}
}

func (zilliz *ZillizPipelineClient) Retrieve(question string, topk int) (chunks []string, err error) {
	logger.Info("Topk: ", topk)
	searchReq := SearchRequest{
		Data: QueryData{
			QueryText: question,
		},
		Params: Params{
			Limit:        topk,
			Offset:       0,
			OutputFields: []string{"chunk_text", "chunk_id", "doc_name"},
		},
	}
	postBody, _ := json.Marshal(searchReq)

	reqBody := bytes.NewBuffer(postBody)
	apiEndpoint := fmt.Sprintf("https://controller.api.%s.zillizcloud.com/v1/pipelines/%s/run", zilliz.Region, zilliz.PipelineID)
	req, err := http.NewRequest("POST", apiEndpoint, reqBody)
	if err != nil {
		logger.Error("Composing req failed. Req: " + reqBody.String() + "Error: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", zilliz.APIKey))
	// send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("Req failed. Req: " + reqBody.String() + "Error: " + err.Error())
	}
	defer resp.Body.Close()

	searchResponse := SearchResponse{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&searchResponse)
	if decodeErr != nil {
		logger.Error("Req failed. Req: " + reqBody.String() + "Error: " + decodeErr.Error())
	}

	if searchResponse.Code != 200 {
		errMsg := fmt.Sprintf("Search resp error code: %v for question: %s, Error resp: %s", searchResponse.Code, question, searchResponse.Message)
		logger.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	for _, result := range searchResponse.Data.Result {
		chunks = append(chunks, result.ChunkText)
		if err != nil {
			logger.Error(fmt.Sprintf("%s while appending %v", err, result))
		}
	}
	return chunks, nil
}
