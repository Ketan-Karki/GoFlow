package worker

import (
	"context"
	"encoding/json"
	"time"

	"goflow/internal/model"
)

// Processor runs a single job type and returns result JSON or error.
type Processor func(ctx context.Context, payload json.RawMessage) (result []byte, err error)

// Processors returns a map of job type -> processor (report, image, email, heavy_task).
func Processors() map[model.JobType]Processor {
	return map[model.JobType]Processor{
		model.TypeReport:    processReport,
		model.TypeImage:     processImage,
		model.TypeEmail:     processEmail,
		model.TypeHeavyTask: processHeavyTask,
	}
}

func processReport(ctx context.Context, payload json.RawMessage) ([]byte, error) {
	// Simulate report generation
	time.Sleep(100 * time.Millisecond)
	return json.Marshal(map[string]string{"report": "generated", "payload": string(payload)})
}

func processImage(ctx context.Context, payload json.RawMessage) ([]byte, error) {
	time.Sleep(50 * time.Millisecond)
	return json.Marshal(map[string]string{"image": "processed", "payload": string(payload)})
}

func processEmail(ctx context.Context, payload json.RawMessage) ([]byte, error) {
	time.Sleep(80 * time.Millisecond)
	return json.Marshal(map[string]string{"email": "sent", "payload": string(payload)})
}

func processHeavyTask(ctx context.Context, payload json.RawMessage) ([]byte, error) {
	// Simulate heavy work
	time.Sleep(500 * time.Millisecond)
	return json.Marshal(map[string]string{"heavy_task": "done", "payload": string(payload)})
}
