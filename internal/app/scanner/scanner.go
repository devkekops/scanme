package scanner

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/logrusorgru/aurora"
	"go.uber.org/ratelimit"

	"github.com/projectdiscovery/nuclei/v2/pkg/catalog"
	"github.com/projectdiscovery/nuclei/v2/pkg/catalog/config"
	"github.com/projectdiscovery/nuclei/v2/pkg/catalog/loader"
	"github.com/projectdiscovery/nuclei/v2/pkg/core"
	"github.com/projectdiscovery/nuclei/v2/pkg/core/inputs"
	"github.com/projectdiscovery/nuclei/v2/pkg/output"
	"github.com/projectdiscovery/nuclei/v2/pkg/parsers"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols/common/hosterrorscache"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols/common/interactsh"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols/common/protocolinit"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols/common/protocolstate"
	"github.com/projectdiscovery/nuclei/v2/pkg/reporting"
	"github.com/projectdiscovery/nuclei/v2/pkg/testutils"
	"github.com/projectdiscovery/nuclei/v2/pkg/types"
)

func copyTemplates(templates []string) string {
	id := uuid.New()
	tmpDirName := "tmp_" + id.String()
	if err := os.Mkdir(tmpDirName, 0777); err != nil {
		log.Fatal(err)
	}

	cwd, _ := os.Getwd()
	tmpDirPath := filepath.Join(cwd, tmpDirName)
	allTemplatesPath := filepath.Join(cwd, "/templates")

	for _, template := range templates {
		from, err := os.Open(filepath.Join(allTemplatesPath, template))
		if err != nil {
			log.Fatal(err)
		}
		defer from.Close()

		to, err := os.OpenFile(filepath.Join(tmpDirPath, template), os.O_RDWR|os.O_CREATE, 0777)
		if err != nil {
			log.Fatal(err)
		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			log.Fatal(err)
		}
	}

	return tmpDirPath
}

func Scan(domains []string, templates []string, resultCh chan *output.ResultEvent) {
	fmt.Println("start scan")
	tmpDirPath := copyTemplates(templates)
	defer os.RemoveAll(tmpDirPath)

	cache := hosterrorscache.New(30, hosterrorscache.DefaultMaxHostsCount)
	defer cache.Close()

	mockProgress := &testutils.MockProgressClient{}
	reportingClient, _ := reporting.New(&reporting.Options{}, "")
	defer reportingClient.Close()

	outputWriter := testutils.NewMockOutputWriter()
	outputWriter.WriteCallback = func(event *output.ResultEvent) {
		event.Request = ""
		event.Response = ""
		//fmt.Printf("Got Result: %v\n", event)
		//events = append(events, event)
		resultCh <- event
	}

	defaultOpts := types.DefaultOptions()
	protocolstate.Init(defaultOpts)
	protocolinit.Init(defaultOpts)

	//defaultOpts.Templates = goflags.FileOriginalNormalizedStringSlice{"dns/cname-service-detection.yaml"}
	//defaultOpts.ExcludeTags = config.ReadIgnoreFile().Tags

	interactOpts := interactsh.NewDefaultOptions(outputWriter, reportingClient, mockProgress)
	interactClient, err := interactsh.New(interactOpts)
	if err != nil {
		log.Fatalf("Could not create interact client: %s\n", err)
	}
	defer interactClient.Close()

	home, _ := os.UserHomeDir()
	catalog := catalog.New(path.Join(home, "nuclei-templates"))
	//catalog := catalog.New("../../internal/app/scanner/templates")
	executerOpts := protocols.ExecuterOptions{
		Output:          outputWriter,
		Options:         defaultOpts,
		Progress:        mockProgress,
		Catalog:         catalog,
		IssuesClient:    reportingClient,
		RateLimiter:     ratelimit.New(150),
		Interactsh:      interactClient,
		HostErrorsCache: cache,
		Colorizer:       aurora.NewAurora(true),
		ResumeCfg:       types.NewResumeCfg(),
	}
	engine := core.New(defaultOpts)
	engine.SetExecuterOptions(executerOpts)

	workflowLoader, err := parsers.NewLoader(&executerOpts)
	if err != nil {
		log.Fatalf("Could not create workflow loader: %s\n", err)
	}
	executerOpts.WorkflowLoader = workflowLoader

	configObject, err := config.ReadConfiguration()
	if err != nil {
		log.Fatalf("Could not read config: %s\n", err)
	}

	configObject.TemplatesDirectory = tmpDirPath

	store, err := loader.New(loader.NewConfig(defaultOpts, configObject, catalog, executerOpts))
	if err != nil {
		log.Fatalf("Could not create loader client: %s\n", err)
	}
	store.Load()

	input := &inputs.SimpleInputProvider{Inputs: domains}
	_ = engine.Execute(store.Templates(), input)
	engine.WorkPool().Wait() // Wait for the scan to finish

	fmt.Println("finish")
	close(resultCh)
}
