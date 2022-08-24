package domainfinder

import (
	"bytes"
	"context"
	"io"
	"log"
	"strings"

	"github.com/projectdiscovery/subfinder/v2/pkg/passive"
	"github.com/projectdiscovery/subfinder/v2/pkg/resolve"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
)

type DomainFinder struct {
	runner *runner.Runner
}

func NewSubfinder() *DomainFinder {
	runnerInstance, _ := runner.NewRunner(&runner.Options{
		Threads:            10,                              // Thread controls the number of threads to use for active enumerations
		Timeout:            30,                              // Timeout is the seconds to wait for sources to respond
		MaxEnumerationTime: 10,                              // MaxEnumerationTime is the maximum amount of time in mins to wait for enumeration
		Resolvers:          resolve.DefaultResolvers,        // Use the default list of resolvers by marshaling it to the config
		Sources:            passive.DefaultSources,          // Use the default list of passive sources
		AllSources:         passive.DefaultAllSources,       // Use the default list of all passive sources
		Recursive:          passive.DefaultRecursiveSources, // Use the default list of recursive sources
		Providers:          &runner.Providers{},             // Use empty api keys for all providers
	})
	return &DomainFinder{runner: runnerInstance}
}

func (r *DomainFinder) Search(domains []string, subdomainCh chan []string) {
	for _, domain := range domains {
		domain = strings.TrimPrefix(domain, "https://")
		buf := bytes.Buffer{}
		err := r.runner.EnumerateSingleDomain(context.Background(), domain, []io.Writer{&buf})
		if err != nil {
			log.Fatal(err)
		}

		data, err := io.ReadAll(&buf)
		if err != nil {
			log.Fatal(err)
		}

		out := strings.Split(string(data), "\n")
		for i, _ := range out {
			out[i] = "https://" + out[i]
		}

		if len(out) > 0 {
			out = out[:len(out)-1]
		}

		subdomainCh <- out
	}

	close(subdomainCh)
}
