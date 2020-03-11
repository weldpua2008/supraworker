package worker
// Worker defines a worker.
type Worker interface {
	Init() error
	Run() error
	Terminate() error
}
