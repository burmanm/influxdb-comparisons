package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hawkular/hawkular-client-go/metrics"
)

type HawkularPusher struct {
	Client    *metrics.Client
	tagsCache map[string]bool
}

func main() {
	p := metrics.Parameters{Tenant: "measurements", Url: "http://localhost:8080", AdminToken: "secret"}
	c, err := metrics.NewHawkularClient(p)
	if err != nil {
		fmt.Printf("Unable to acquire Hawkular-Metrics connection\n")
		fmt.Printf(err.Error())
	}
	pusher := HawkularPusher{
		Client: c,
	}
	err = pusher.Scan()
	if err != nil {
		fmt.Printf("Failed to read: %s\n", err.Error())
	}
}

func (p HawkularPusher) Scan() error {
	p.tagsCache = make(map[string]bool, 0)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		mH := metrics.MetricHeader{}
		err := json.Unmarshal(scanner.Bytes(), &mH)
		if err != nil {
			return err
		}
		mH.Type = metrics.Gauge

		if found := p.tagsCache[mH.ID]; !found {
			tags := mH.Data[0].Tags
			err = p.Client.UpdateTags(mH.Type, mH.ID, tags)
			if err != nil {
				return err
			}
			p.tagsCache[mH.ID] = true
		}

		// Remove tags from the datapoint
		mH.Data[0].Tags = nil

		err = p.Client.Write([]metrics.MetricHeader{mH})
		if err != nil {
			return err
		}
	}
	return nil
}
