package builder

import (
	"errors"
	"time"
	"ubbagent/aggregator"
	"ubbagent/config"
	"ubbagent/endpoint"
	"ubbagent/endpoint/disk"
	"ubbagent/persistence"
	"ubbagent/pipeline"
	"ubbagent/sender"
)

// Build builds pipeline containing a configured Aggregator and all of the resources
// (persistence, endpoints) behind it. It returns the pipeline.Head.
func Build(cfg *config.Config, p persistence.Persistence) (pipeline.Head, error) {
	endpoints, err := createEndpoints(cfg)
	if err != nil {
		return nil, err
	}
	senders := make([]sender.Sender, len(endpoints))
	for i := range endpoints {
		senders[i] = sender.NewRetryingSender(endpoints[i], p)
	}
	d := sender.NewDispatcher(senders)

	return aggregator.NewAggregator(cfg.Metrics, d, p), nil
}

func createEndpoints(config *config.Config) ([]endpoint.Endpoint, error) {
	var eps []endpoint.Endpoint
	for _, cfgep := range config.Endpoints {
		ep, err := createEndpoint(cfgep)
		if err != nil {
			// TODO(volkman): close already-created endpoints in event of error?
			return nil, err
		}
		eps = append(eps, ep)
	}
	return eps, nil
}

func createEndpoint(cfgep config.Endpoint) (endpoint.Endpoint, error) {
	if cfgep.Disk != nil {
		return disk.NewDiskEndpoint(cfgep.Name, cfgep.Disk.ReportDir, time.Duration(cfgep.Disk.ExpireSeconds)*time.Second), nil
	}
	// TODO(volkman): support servicecontrol and pubsub
	return nil, errors.New("unsupported endpoint")
}