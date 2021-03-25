package main

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"google.golang.org/grpc"
	"log"
	"os"
	"time"
)

func initProvider() func() {
	ctx := context.Background()

	log.Printf("trying to get otel collector endpoint")
	otelCollectorAddr, ok := os.LookupEnv("OTEL_COLLECTOR_ENDPOINT")
	if !ok {
		log.Fatal("cannot get otel collector endpoint")
	}

	log.Printf("trying to create otel exporter")
	exp, err := otlp.NewExporter(
		ctx,
		otlpgrpc.NewDriver(
			otlpgrpc.WithInsecure(),
			otlpgrpc.WithEndpoint(otelCollectorAddr),
			otlpgrpc.WithDialOption(grpc.WithBlock()),
		),
	)
	handleErr(err, "failed to create exporter")

	log.Printf("trying to get service name")
	serviceName, ok := os.LookupEnv("SERVICE_NAME")
	if !ok {
		log.Fatal("cannot get service name from env variables")
	}

	log.Printf("trying to create otel resource")
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	handleErr(err, "failed to create resource")

	log.Printf("trying to create batch span processor")
	bsp := sdktrace.NewBatchSpanProcessor(exp)

	log.Printf("trying to create trace provider")
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
	)

	log.Printf("trying to create controller")
	cont := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exp,
		),
		controller.WithCollectPeriod(5*time.Second),
		controller.WithExporter(exp),
	)

	log.Printf("setting global parameters")
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)
	global.SetMeterProvider(cont.MeterProvider())
	handleErr(cont.Start(context.Background()), "failed to start metric controller")

	return func() {
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown provider")
		handleErr(cont.Stop(context.Background()), "failed to stop metrics controller") // pushes any last exports to the receiver
		handleErr(exp.Shutdown(ctx), "failed to stop exporter")
	}
}
