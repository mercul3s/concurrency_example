package main

import (
	"context"
	"fmt"
	"log"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
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
	client, err := v3.New(
		v3.Config{
			Endpoints:   []string{"localhost:2379"},
			DialTimeout: 5 * time.Second,
		})
	if err != nil {
		fmt.Println("Error establishing client ", err)
	}

	// create two separate sessions for lock competition
	s1, err := concurrency.NewSession(client)
	if err != nil {
		log.Fatal(err)
	}
	defer s1.Close()
	m1 := concurrency.NewMutex(s1, "/my-lock/")

	s2, err := concurrency.NewSession(client)
	if err != nil {
		log.Fatal(err)
	}
	defer s2.Close()
	m2 := concurrency.NewMutex(s2, "/my-lock/")

	// acquire lock for s1
	if err := m1.Lock(context.TODO()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("acquired lock for s1")

	m2Locked := make(chan struct{})
	go func() {
		defer close(m2Locked)
		// wait until s1 is locks /my-lock/
		if err := m2.Lock(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	if err := m1.Unlock(context.TODO()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("released lock for s1")

	<-m2Locked
	fmt.Println("acquired lock for s2")

}
