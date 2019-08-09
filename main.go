package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/spf13/cobra"
)

// Serverspec contains the full test report.
type Serverspec struct {
	Version     string    `json:"version"`
	Examples    []Example `json:"examples"`
	Summary     Summary   `json:"summary"`
	SummaryLine string    `json:"summary_line"`
}

// Example contains information about a single test.
type Example struct {
	ID              string    `json:"id"`
	Description     string    `json:"description"`
	FullDescription string    `json:"full_description"`
	Status          string    `json:"status"`
	FilePath        string    `json:"file_path"`
	LineNumber      int       `json:"line_number"`
	RunTime         float32   `json:"run_time"`
	PendingMessage  string    `json:"pending_message"`
	Exception       Exception `json:"exception"`
}

// Exception contains any errors that occurred in the test.
type Exception struct {
	Class     string   `json:"class"`
	Message   string   `json:"message"`
	Backtrace []string `json:"backtrace"`
}

// Summary is a summary of the full test report.
type Summary struct {
	Duration                    float32 `json:"duration"`
	ExampleCount                int     `json:"example_count"`
	FailureCount                int     `json:"failure_count"`
	PendingCount                int     `json:"pending_count"`
	ErrorsOutsideOfExampleCount int     `json:"errors_outside_of_examples_count"`
}

var (
	token     string
	namespace string
	url       string
	handlers  []string
	stdin     *os.File
)

func main() {
	rootCmd := configureRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sensu-serverspec-events",
		Short: "a serverspec json handler to use with sensu",
		RunE:  run,
	}

	cmd.Flags().StringSliceVarP(&handlers,
		"handlers",
		"",
		[]string{},
		"sensu handlers that the new serverspec events will be handled by")
	cmd.Flags().StringVarP(&namespace,
		"namespace",
		"n",
		"default",
		"sensu namespace that the new serverspec events will be created in")
	cmd.Flags().StringVarP(&token,
		"token",
		"t",
		os.Getenv("SENSU_API_TOKEN"),
		"sensu api token, (default is the value of the SENSU_API_TOKEN environment variable)")
	cmd.Flags().StringVarP(&url,
		"url",
		"u",
		"http://127.0.0.1:8080",
		"sensu api url")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		_ = cmd.Help()
		return fmt.Errorf("invalid argument(s) received")
	}

	if token == "" {
		_ = cmd.Help()
		return fmt.Errorf("sensu api token is required")
	}

	if stdin == nil {
		stdin = os.Stdin
	}

	b, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err)
	}

	event := &corev2.Event{}
	err = json.Unmarshal(b, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", err)
	}

	if event.HasMetrics() {
		return fmt.Errorf("event should not contain metrics")
	}

	serverspec := &Serverspec{}
	err = json.Unmarshal([]byte(strings.ReplaceAll(event.Check.Output, "\n", "")), serverspec)
	if err != nil {
		return fmt.Errorf("failed to unmarshal serverspec data: %s", err)
	}

	if len(serverspec.Examples) == 0 {
		return fmt.Errorf("no serverspec examples in check output")
	}

	for _, e := range serverspec.Examples {
		newEvent := event
		newEvent.Check.Name = cleanExampleID(e.ID)
		newEvent.Check.Namespace = namespace
		newEvent.Entity.Namespace = namespace
		newEvent.Check.Handlers = handlers
		newEvent.Check.Output = fmt.Sprint(e)
		newEvent.Timestamp = time.Now().Unix()
		switch e.Status {
		case "passed":
			newEvent.Check.State = corev2.EventPassingState
			newEvent.Check.Status = 0
		case "failed":
			newEvent.Check.State = corev2.EventFailingState
			newEvent.Check.Status = 1
		case "unknown":
			newEvent.Check.State = corev2.EventFailingState
			newEvent.Check.Status = 2
		case "pending":
			newEvent.Check.State = corev2.EventFailingState
			newEvent.Check.Status = 2
		default:
			return fmt.Errorf("unknown serverspec status")
		}

		if err = SendEventToAPI(newEvent); err != nil {
			return err
		}
	}
	return nil
}

func cleanExampleID(original string) string {
	id := strings.Replace(original, "/", "-", -1)
	id = strings.Replace(id, "[", "-", -1)
	id = strings.Replace(id, "]", "-", -1)
	id = strings.Replace(id, ":", "-", -1)
	id = strings.TrimPrefix(id, ".")
	id = strings.TrimPrefix(id, "-")
	id = strings.TrimSuffix(id, "-")
	return id
}

// SendEventToAPI sends a new event to the Sensu Events API.
func SendEventToAPI(newEvent *corev2.Event) error {
	eventBytes, err := json.Marshal(newEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %s", err)
	}

	api := fmt.Sprintf("%s/api/core/v2/namespaces/%s/events/%s/%s", url, namespace, newEvent.Entity.Name, newEvent.Check.Name)
	req, err := http.NewRequest("PUT", api, bytes.NewBuffer(eventBytes))
	if err != nil {
		return fmt.Errorf("failed to create http request: %s", err)
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send event data: %s", err)
	}
	defer resp.Body.Close()

	fmt.Printf("sent sensu event to api %s\n", api)
	return nil
}
