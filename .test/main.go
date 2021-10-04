package main

import (
	"encoding/csv"
	"github.com/zalando/skipper/scheduler"
	"log"
	"os"
	"strconv"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	f, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}

	bm := scheduler.NewBucketManager()

	r := csv.NewReader(f)
	w := csv.NewWriter(os.Stdout)
	records, err := r.ReadAll()
	if err != nil {
		return err
	}
	oldConcurrency := 50.0
	for _, record := range records {
		ff, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return err
		}
		bm.Collect(ff)
		newConcurrency := scheduler.CalculateConcurrency(bm, oldConcurrency)
		if newConcurrency >= 50 {
			newConcurrency = 50
			bm.Next()
		} else {
			bm.ResetCurrent()
		}
		oldConcurrency = newConcurrency
		_ = w.Write([]string{record[0], record[1], strconv.Itoa(int(newConcurrency))})
	}
	w.Flush()
	return nil
}
