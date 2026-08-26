package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/LYH2263/go-ratelimit/dashboard"
	"github.com/LYH2263/go-ratelimit/ratelimit"
)

func main() {
	addr := flag.String("addr", ":8224", "listen address")
	shards := flag.Int("shards", 32, "store shard count")
	flag.Parse()

	store := ratelimit.NewShardedStore(*shards)
	lim, err := ratelimit.New(ratelimit.Options{Store: store, Shards: *shards})
	if err != nil {
		log.Fatal(err)
	}
	defer lim.Close()

	api := &dashboard.API{Limiter: lim}
	srv := dashboard.New(api)
	log.Printf("ratelimit dashboard listening on %s", *addr)
	if err := http.ListenAndServe(*addr, srv.Handler); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
