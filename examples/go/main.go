package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:9090",
	})
	defer rdb.Close()

	// PING
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("PING failed: %v", err)
	}
	fmt.Printf("PING: %s\n", pong)

	// SET
	err = rdb.Set(ctx, "name", "nexo", 0).Err()
	if err != nil {
		log.Fatalf("SET failed: %v", err)
	}
	fmt.Println("SET name nexo: OK")

	// GET
	val, err := rdb.Get(ctx, "name").Result()
	if err != nil {
		log.Fatalf("GET failed: %v", err)
	}
	fmt.Printf("GET name: %s\n", val)

	// SET with EX (seconds)
	err = rdb.Set(ctx, "temp", "expires-in-5s", 5*time.Second).Err()
	if err != nil {
		log.Fatalf("SET EX failed: %v", err)
	}
	fmt.Println("SET temp expires-in-5s EX 5: OK")

	// GET before expiry
	val, err = rdb.Get(ctx, "temp").Result()
	if err != nil {
		log.Fatalf("GET temp failed: %v", err)
	}
	fmt.Printf("GET temp (before expiry): %s\n", val)

	// EXPIRE
	err = rdb.Expire(ctx, "name", 10*time.Second).Err()
	if err != nil {
		log.Fatalf("EXPIRE failed: %v", err)
	}
	fmt.Println("EXPIRE name 10: OK")

	// DEL
	err = rdb.Del(ctx, "temp").Err()
	if err != nil {
		log.Fatalf("DEL failed: %v", err)
	}
	fmt.Println("DEL temp: OK")

	// GET after DEL
	_, err = rdb.Get(ctx, "temp").Result()
	if err == redis.Nil {
		fmt.Println("GET temp (after DEL): key not found (expected)")
	} else if err != nil {
		log.Fatalf("GET temp after DEL failed: %v", err)
	}

	fmt.Println("\nAll examples completed successfully!")
}
