package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/remiges-tech/await/retry"
)

// Pattern 1: Function with multiple inputs - use closure
func multiInputExample() {
	ctx := context.Background()

	// Original function with multiple parameters
	add := func(a, b, c int) (int, error) {
		if a < 0 || b < 0 || c < 0 {
			return 0, errors.New("negative numbers not allowed")
		}
		return a + b + c, nil
	}

	// Wrap in closure to capture parameters
	a, b, c := 1, 2, 3
	result, err := retry.Do(ctx, func(ctx context.Context) (int, error) {
		return add(a, b, c)
	}, retry.WithMaxAttempts(3))

	if err != nil {
		log.Printf("Failed: %v", err)
	} else {
		log.Printf("Sum of %d + %d + %d = %d", a, b, c, result)
	}
}

// Pattern 2: Function with multiple outputs - use struct
func multiOutputWithStruct() {
	ctx := context.Background()

	// Function that returns multiple values
	divide := func(a, b float64) (quotient float64, remainder float64, err error) {
		if b == 0 {
			return 0, 0, errors.New("division by zero")
		}
		quotient = float64(int(a / b))
		remainder = a - (quotient * b)
		return quotient, remainder, nil
	}

	// Define a result struct
	type DivResult struct {
		Quotient  float64
		Remainder float64
	}

	// Wrap to return struct
	result, err := retry.Do(ctx, func(ctx context.Context) (DivResult, error) {
		q, r, err := divide(10, 3)
		if err != nil {
			return DivResult{}, err
		}
		return DivResult{Quotient: q, Remainder: r}, nil
	}, retry.WithMaxAttempts(3))

	if err != nil {
		log.Printf("Failed: %v", err)
	} else {
		log.Printf("10 ÷ 3 = %g remainder %g", result.Quotient, result.Remainder)
	}
}

// Pattern 3: Function with multiple outputs - use closure variables
func multiOutputWithClosure() {
	ctx := context.Background()

	// Variables to capture outputs
	var name string
	var age int
	var active bool

	// Wrap function to capture outputs in closure
	_, err := retry.Do(ctx, func(ctx context.Context) (struct{}, error) {
		// Simulate getting user info that returns multiple values
		n, a, act, err := getUserInfo("user123")
		if err != nil {
			return struct{}{}, err
		}
		// Capture the results
		name = n
		age = a
		active = act
		return struct{}{}, nil
	}, retry.WithMaxAttempts(3))

	if err != nil {
		log.Printf("Failed: %v", err)
	} else {
		log.Printf("User: %s, Age: %d, Active: %v", name, age, active)
	}
}

// Simulate a function that returns multiple values
func getUserInfo(id string) (string, int, bool, error) {
	if id == "" {
		return "", 0, false, errors.New("empty ID")
	}
	return "John Doe", 30, true, nil
}

// Pattern 4: Create a generic helper
func retryWith2Returns[T, U any](
	ctx context.Context,
	fn func() (T, U, error),
	opts retry.Options,
) (T, U, error) {
	var result1 T
	var result2 U

	_, err := retry.Do(ctx, func(ctx context.Context) (struct{}, error) {
		r1, r2, err := fn()
		if err != nil {
			return struct{}{}, err
		}
		result1 = r1
		result2 = r2
		return struct{}{}, nil
	}, opts)

	return result1, result2, err
}

func genericHelperExample() {
	ctx := context.Background()

	// Use the generic helper
	str, num, err := retryWith2Returns(ctx, func() (string, int, error) {
		// Your function that returns 2 values + error
		return "hello", 42, nil
	}, retry.WithMaxAttempts(3))

	if err != nil {
		log.Printf("Failed: %v", err)
	} else {
		log.Printf("Got: %s and %d", str, num)
	}
}

func main() {
	fmt.Println("=== Pattern 1: Multiple Inputs ===")
	multiInputExample()

	fmt.Println("\n=== Pattern 2: Multiple Outputs with Struct ===")
	multiOutputWithStruct()

	fmt.Println("\n=== Pattern 3: Multiple Outputs with Closure ===")
	multiOutputWithClosure()

	fmt.Println("\n=== Pattern 4: Generic Helper ===")
	genericHelperExample()
}
