// Copyright 2017, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command prometheus is an example program that collects data for
// video size over a time window. Collected data is exported to Prometheus.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"go.opencensus.io/exporter/stats/prometheus"
	"go.opencensus.io/stats"
)

func main() {
	ctx := context.Background()

	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatal(err)
	}
	stats.RegisterExporter(exporter)

	// Create measures. The program will record measures for the size of
	// processed videos and the nubmer of videos marked as spam.
	videoCount, err := stats.NewMeasureInt64("my.org/measures/video_count", "number of processed videos", "")
	if err != nil {
		log.Fatalf("Video size measure not created: %v", err)
	}

	// Create view to see the processed video size cumulatively.
	view, err := stats.NewView(
		"video_count",
		"processed video size over time",
		nil,
		videoCount,
		stats.CountAggregation{},
		stats.Cumulative{},
	)
	if err != nil {
		log.Fatalf("Cannot create view: %v", err)
	}

	// Set reporting period to report data at every second.
	stats.SetReportingPeriod(1 * time.Second)

	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	if err := view.Subscribe(); err != nil {
		log.Fatalf("Cannot subscribe to the view: %v", err)
	}

	go func() {
		for {
			// Record some data points.
			stats.Record(ctx, videoCount.M(1))
			<-time.After(time.Millisecond * time.Duration(1+rand.Intn(400)))
		}
	}()

	// Wait for a duration longer than reporting duration to ensure the stats
	// library reports the collected data.
	fmt.Println("Wait longer than the reporting duration...")

	http.Handle("/metrics", exporter)
	log.Fatal(http.ListenAndServe(":9999", nil))
}