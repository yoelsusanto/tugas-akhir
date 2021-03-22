package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

// GenerateLoad does a sha256 operation for some Repetition
func GenerateLoad(repetition int) string {
	hash := "someString"
	for i := 0; i <= repetition; i++ {
		hash = fmt.Sprintf("%x", sha256.Sum256([]byte(hash)))
	}
	return hash
}

func main() {
	e := echo.New()

	e.POST("/work", func(ctx echo.Context) error {
		req := &ReceivedTraffic{}
		err := ctx.Bind(req)
		if err != nil {
			return err
		}

		var forwardResponses []ForwardResponse

		for _, forward := range req.Forwards {
			forwardJson, err := json.Marshal(forward)
			if err != nil {
				return err
			}

			url := fmt.Sprintf("http://%s", forward.Service)
			r, err := http.Post(url, "application/json", bytes.NewReader(forwardJson))
			if err != nil {
				fmt.Printf("err creating a request: %+v\n", err)
				continue
			}

			svcSelfResponse := &SelfResponse{}
			err = json.NewDecoder(r.Body).Decode(svcSelfResponse)
			if err != nil {
				fmt.Printf("err unmarshalling request: %+v\n", err)
				continue
			}
			forwardResponses = append(forwardResponses, ForwardResponse{
				Service:         forward.Service,
				Result:          svcSelfResponse.Result,
				Repetition:      svcSelfResponse.Repetition,
				ForwardResponse: svcSelfResponse.ForwardResponse,
			})
		}

		result := GenerateLoad(req.Repetition)
		return ctx.JSON(http.StatusOK, SelfResponse{
			Result:          result,
			Repetition:      req.Repetition,
			ForwardResponse: forwardResponses,
		})
	})

	e.Logger.Fatal(e.Start(":1234"))
}
