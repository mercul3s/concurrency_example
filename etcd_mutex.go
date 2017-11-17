package main

import (
	"context"
	"fmt"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	c "github.com/coreos/etcd/clientv3/concurrency"
)

// Only one client session is needed. It should be reused because it has
// internal state, and is safe to use with multiple goroutines.
// Failure cases:
// 	- attempt a deadlock
// 	- attempt an unlock based on a lease expiration

const (
	mutexRoot = "/lockpfx"
)

func main() {
	ctx := context.TODO()
	client, err := v3.New(
		v3.Config{
			Endpoints:   []string{"localhost:2379"},
			DialTimeout: 5 * time.Second,
		})
	if err != nil {
		fmt.Println("Error establishing client ", err)
	}

	defer client.Close()
	session, err := c.NewSession(client)
	if err != nil {
		fmt.Println("error creating session ", err)
	}

	mutex := c.NewMutex(session, mutexRoot)
	if err = mutex.Lock(ctx); err != nil {
		fmt.Println("error locking mutex ", err)
	}

	fmt.Println("Mutex locked")
	if err = mutex.Unlock(ctx); err != nil {
		fmt.Println("error unlocking mutex", err)
	}
	fmt.Println("Mutex unlocked")
}
