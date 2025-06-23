package main

import (
	"fmt"
	"os"

	"github.com/Lyra-poing-serre/gator/cmd/aggregator"
	"github.com/Lyra-poing-serre/gator/internal/settings"

	_ "github.com/lib/pq"
)

func main() {
	conf, err := settings.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	aggregator.LaunchAggregator(&conf)
}
