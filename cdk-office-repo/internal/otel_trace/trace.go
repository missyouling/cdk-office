/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package otel_trace

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var (
	// TracerProvider 全局追踪提供者
	TracerProvider *trace.TracerProvider
)

// Init 初始化链路追踪
func Init(serviceName string) {
	// 创建Jaeger导出器
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	if err != nil {
		log.Printf("[OTEL] failed to create Jaeger exporter: %v", err)
		return
	}
	
	// 创建追踪提供者
	TracerProvider = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	
	// 设置全局追踪提供者
	otel.SetTracerProvider(TracerProvider)
	
	log.Println("[OTEL] OpenTelemetry tracing initialized")
}

// GetTracer 获取追踪器
func GetTracer() trace.Tracer {
	if TracerProvider == nil {
		return nil
	}
	return TracerProvider.Tracer("cdk-office")
}

// Shutdown 关闭链路追踪
func Shutdown(ctx context.Context) {
	if TracerProvider != nil {
		if err := TracerProvider.Shutdown(ctx); err != nil {
			log.Printf("[OTEL] failed to shutdown tracer provider: %v", err)
		}
	}
	log.Println("[OTEL] OpenTelemetry tracing shutdown")
}