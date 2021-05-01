package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"os"
)

var tracer = otel.Tracer("tugas-akhir")

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

// GenerateLoad does a sha256 operation for some Repetition
func GenerateLoad(ctx context.Context, repetition int) string {
	_, span := tracer.Start(ctx, "GenerateLoad", trace.WithAttributes(attribute.Int("repetition", repetition)))
	defer span.End()

	hash := "someString"
	for i := 0; i <= repetition; i++ {
		hash = fmt.Sprintf("%x", sha256.Sum256([]byte(hash)))
	}
	return hash
}

func main() {
	shutdown := initProvider()
	defer shutdown()

	serviceName, ok := os.LookupEnv("SERVICE_NAME")
	if !ok {
		log.Fatal("cannot get service name from env variables")
	}

	e := echo.New()
	e.Use(otelecho.Middleware(serviceName))

	propagator := otel.GetTextMapPropagator()

	e.POST("/work", func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		req := &ReceivedTraffic{}
		err := echoCtx.Bind(req)
		if err != nil {
			fmt.Printf("Error binding request to req: %+v\n", err)
			return err
		}

		result := GenerateLoad(ctx, req.Repetition)

		var forwardResponses []ForwardResponse

		client := &http.Client{}

		for _, forward := range req.Forwards {
			forwardJson, err := json.Marshal(forward)
			if err != nil {
				return err
			}

			url := fmt.Sprintf("http://%s", forward.Service)

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(forwardJson))
			if err != nil {
				fmt.Printf("err creating a request: %+v\n", err)
				return err
			}

			header := req.Header
			header.Set("Content-Type", "application/json")
			propagator.Inject(ctx, propagation.HeaderCarrier(header))

			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("err executing request: %+v\n", err)
				continue
			}

			svcSelfResponse := &SelfResponse{}
			err = json.NewDecoder(resp.Body).Decode(svcSelfResponse)
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

		return echoCtx.JSON(http.StatusOK, SelfResponse{
			Result:          result,
			Repetition:      req.Repetition,
			ForwardResponse: forwardResponses,
		})
	})

	e.Logger.Fatal(e.Start(":1234"))
}
