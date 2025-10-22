package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Config struct {
	URL         string
	Requests    int
	Concurrency int
}

type Result struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

type Report struct {
	TotalTime       time.Duration
	TotalRequests   int
	SuccessRequests int
	StatusCodes     map[int]int
}

func parseFlags() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.URL, "url", "", "URL do serviço a ser testado")
	flag.IntVar(&config.Requests, "requests", 0, "Número total de requests")
	flag.IntVar(&config.Concurrency, "concurrency", 0, "Número de chamadas simultâneas")
	flag.Parse()

	if config.URL == "" {
		return nil, fmt.Errorf("parâmetro --url é obrigatório")
	}
	if config.Requests <= 0 {
		return nil, fmt.Errorf("parâmetro --requests deve ser maior que 0")
	}
	if config.Concurrency <= 0 {
		return nil, fmt.Errorf("parâmetro --concurrency deve ser maior que 0")
	}
	if config.Concurrency > config.Requests {
		config.Concurrency = config.Requests
	}

	return config, nil
}

func worker(ctx context.Context, client *http.Client, url string, jobs <-chan int, results chan<- Result) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-jobs:
			if !ok {
				return
			}

			startTime := time.Now()
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				results <- Result{Error: err, Duration: time.Since(startTime)}
				continue
			}

			resp, err := client.Do(req)
			duration := time.Since(startTime)

			if err != nil {
				results <- Result{Error: err, Duration: duration}
				continue
			}

			resp.Body.Close()
			results <- Result{
				StatusCode: resp.StatusCode,
				Duration:   duration,
			}

		}
	}
}

func runLoadTest(config *Config) *Report {
	fmt.Printf("Iniciando teste de carga...\n")
	fmt.Printf("URL: %s\n", config.URL)
	fmt.Printf("Total de requests: %d\n", config.Requests)
	fmt.Printf("Concorrência: %d\n", config.Concurrency)
	fmt.Println()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan int, config.Requests)
	results := make(chan Result, config.Requests)

	var wg sync.WaitGroup
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(ctx, client, config.URL, jobs, results)
		}()
	}

	startTime := time.Now()
	go func() {
		defer close(jobs)
		for i := 0; i < config.Requests; i++ {
			jobs <- i
		}
	}()

	report := &Report{
		StatusCodes: make(map[int]int),
	}

	for i := 0; i < config.Requests; i++ {
		result := <-results
		report.TotalRequests++

		if result.Error != nil {
			report.StatusCodes[0]++
		} else {
			report.StatusCodes[result.StatusCode]++
			if result.StatusCode == 200 {
				report.SuccessRequests++
			}
		}

		if (i+1)%100 == 0 || i+1 == config.Requests {
			fmt.Printf("Progress: %d/%d requests completed\n", i+1, config.Requests)
		}
	}

	report.TotalTime = time.Since(startTime)

	cancel()
	wg.Wait()
	close(results)

	return report
}

func printReport(report *Report) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("RELATÓRIO DE TESTE DE CARGA")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("Tempo total de execução: %v\n", report.TotalTime)
	fmt.Printf("Total de requests realizados: %d\n", report.TotalRequests)
	fmt.Printf("Requests com status 200: %d\n", report.SuccessRequests)

	successRate := float64(report.SuccessRequests) / float64(report.TotalRequests) * 100
	fmt.Printf("Taxa de sucesso: %.2f%%\n", successRate)

	requestsPerSecond := float64(report.TotalRequests) / report.TotalTime.Seconds()
	fmt.Printf("Requests por segundo: %.2f\n", requestsPerSecond)

	fmt.Println("\nDistribuição de códigos de status:")
	for statusCode, count := range report.StatusCodes {
		percentage := float64(count) / float64(report.TotalRequests) * 100
		if statusCode == 0 {
			fmt.Printf("  Errors: %d (%.2f%%)\n", count, percentage)
		} else {
			fmt.Printf("  %d: %d (%.2f%%)\n", statusCode, count, percentage)
		}
	}
	fmt.Println(strings.Repeat("=", 50))
}

func main() {
	config, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nUso: %s --url=<URL> --requests=<NUM> --concurrency=<NUM>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	report := runLoadTest(config)
	printReport(report)
}
