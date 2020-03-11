package job
// Worker defines a job.
type Job interface {
	Get() error
	Run() error
	Terminate() error
    Finish() error
}
