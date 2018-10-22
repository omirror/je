package je

// Queue describes an interface for publishing jobs and subscribing to job
// state changes such as stopped, killed or errored. Implementations are
// expected to publish the new job to appropriate topics on a queue and
// setup subscriptions to listen for job state changes
type Queue interface {
	Publish(job *Job) error
	Subscribe(job *Job) error
	Close() error
}
