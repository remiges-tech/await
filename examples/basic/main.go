package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/remiges-tech/await"
	"github.com/remiges-tech/await/retry"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Await Library Examples ===")

	// Example 1: All - Wait for all tasks to complete
	fmt.Println("1. All() - Fetch multiple APIs concurrently:")
	apiTasks := []await.Task[string]{
		func(ctx context.Context) (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "API 1 response", nil
		},
		func(ctx context.Context) (string, error) {
			time.Sleep(150 * time.Millisecond)
			return "API 2 response", nil
		},
		func(ctx context.Context) (string, error) {
			time.Sleep(80 * time.Millisecond)
			return "API 3 response", nil
		},
	}

	results, err := await.All(ctx, apiTasks...)
	if err != nil {
		log.Printf("Function error in All: %v", err)
	} else {
		// Check for task errors and extract values
		var values []string
		for i, result := range results {
			if result.Err != nil {
				fmt.Printf("   Task %d error: %v\n", i+1, result.Err)
			} else {
				values = append(values, result.Value)
			}
		}
		fmt.Printf("   Successful results: %v\n", values)
	}

	// Example 2: Any - Return first successful result
	fmt.Println("\n2. Any() - Try multiple endpoints until one succeeds:")
	endpoints := []await.Task[string]{
		func(ctx context.Context) (string, error) {
			time.Sleep(200 * time.Millisecond)
			return "", fmt.Errorf("endpoint 1 failed")
		},
		func(ctx context.Context) (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "Endpoint 2 succeeded", nil
		},
		func(ctx context.Context) (string, error) {
			time.Sleep(300 * time.Millisecond)
			return "Endpoint 3 succeeded", nil
		},
	}

	result, err := await.Any(ctx, endpoints...)
	if err != nil {
		log.Printf("All endpoints failed: %v", err)
	} else {
		fmt.Printf("   First success: %s\n", result)
	}

	// Example 3: Race - Get the fastest response
	fmt.Println("\n3. Race() - Get the fastest response:")
	raceTasks := []await.Task[int]{
		func(ctx context.Context) (int, error) {
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
			return 1, nil
		},
		func(ctx context.Context) (int, error) {
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
			return 2, nil
		},
		func(ctx context.Context) (int, error) {
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
			return 3, nil
		},
	}

	winner, err := await.Race(ctx, raceTasks...)
	if err != nil {
		log.Printf("Race error: %v", err)
	} else {
		fmt.Printf("   Winner: Task %d\n", winner)
	}

	// Example 4: All with mixed success/failure - Get all results
	fmt.Println("\n4. All() with mixed results - Get all results with errors:")
	mixedTasks := []await.Task[int]{
		func(ctx context.Context) (int, error) {
			return 1, nil
		},
		func(ctx context.Context) (int, error) {
			return 0, fmt.Errorf("task 2 failed")
		},
		func(ctx context.Context) (int, error) {
			return 3, nil
		},
	}

	settled, err := await.All(ctx, mixedTasks...)
	if err != nil {
		log.Printf("Function error: %v", err)
	} else {
		for i, result := range settled {
			if result.Err != nil {
				fmt.Printf("   Task %d: Error - %v\n", i+1, result.Err)
			} else {
				fmt.Printf("   Task %d: Success - %d\n", i+1, result.Value)
			}
		}
	}

	// Example 5: Retry with exponential backoff
	fmt.Println("\n5. retry.Do() - Retry failing operations:")
	attempts := 0
	flakeyTask := await.Task[string](func(ctx context.Context) (string, error) {
		attempts++
		fmt.Printf("   Attempt %d...\n", attempts)
		if attempts < 3 {
			return "", fmt.Errorf("temporary failure")
		}
		return "Success after retries!", nil
	})

	retryResult, err := retry.Do(ctx, flakeyTask, retry.Options{
		MaxAttempts: 5,
		Strategy: &retry.ExponentialBackoff{
			InitialDelay: 100 * time.Millisecond,
			Multiplier:   2,
			MaxDelay:     1 * time.Second,
		},
	})
	if err != nil {
		log.Printf("Retry failed: %v", err)
	} else {
		fmt.Printf("   Result: %s\n", retryResult)
	}

	fmt.Println("\n=== Examples Complete ===")
}
