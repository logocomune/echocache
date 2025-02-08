package main

import (
	"context"
	"github.com/logocomune/echocache"
	"github.com/logocomune/echocache/simple"

	"fmt"
	"time"
)

func main() {
	cache := simple.NewSimpleCache[int]()
	var rf echocache.RefreshFunc[int] = func(ctx context.Context) (int, error) {
		time.Sleep(time.Second * 1)
		return int(time.Now().Unix()), nil
	}
	e := echocache.New[int](cache, rf, time.Hour)
	start := time.Now()
	val, ok, err := e.Get(context.Background(), "key")
	fmt.Println(val, ok, err, time.Since(start))

	start = time.Now()
	val, ok, err = e.Get(context.Background(), "key")
	fmt.Println(val, ok, err, time.Since(start))
	start = time.Now()
	e.Refresh("")
	val, ok, err = e.Get(context.Background(), "key")
	fmt.Println(val, ok, err, time.Since(start))

	time.Sleep(time.Second * 2)
	start = time.Now()
	val, ok, err = e.Get(context.Background(), "key")
	fmt.Println(val, ok, err, time.Since(start))
}
