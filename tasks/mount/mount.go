package mount

import (
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/dnephin/dobi/config"
	"github.com/dnephin/dobi/logging"
	"github.com/dnephin/dobi/tasks/context"
)

// CreateTask is a task which creates a directory on the host
type CreateTask struct {
	name   string
	config *config.MountConfig
}

// NewCreateTask creates a new CreateTask object
func NewCreateTask(name string, conf *config.MountConfig) *CreateTask {
	return &CreateTask{name: name, config: conf}
}

// Name returns the name of the task
func (t *CreateTask) Name() string {
	return t.name
}

func (t *CreateTask) logger() *log.Entry {
	return logging.Log.WithFields(log.Fields{"task": t})
}

// Repr formats the task for logging
func (t *CreateTask) Repr() string {
	return fmt.Sprintf("[mount:create %s] %s (%s)", t.name, t.config.Bind, t.config.Mode)
}

// Run creates the host path if it doesn't already exist
func (t *CreateTask) Run(ctx *context.ExecuteContext) error {
	t.logger().Debug("Run")

	if t.exists(ctx) {
		t.logger().Debug("is fresh")
		return nil
	}

	if err := t.create(ctx); err != nil {
		return err
	}
	ctx.SetModified(t.name)
	t.logger().Info("Created")
	return nil
}

func (t *CreateTask) create(ctx *context.ExecuteContext) error {
	path := AbsBindPath(t.config, ctx.WorkingDir)
	mode := t.config.Mode

	switch t.config.File {
	case true:
		if mode == 0 {
			mode = 0644
		}
		return ioutil.WriteFile(path, []byte{}, os.FileMode(mode))
	default:
		if mode == 0 {
			mode = 0777
		}
		return os.MkdirAll(path, os.FileMode(mode))
	}
}

// Dependencies returns the list of dependencies
func (t *CreateTask) Dependencies() []string {
	return t.config.Dependencies()
}

func (t *CreateTask) exists(ctx *context.ExecuteContext) bool {
	_, err := os.Stat(AbsBindPath(t.config, ctx.WorkingDir))
	if err != nil {
		return false
	}

	return true
}

// Stop the task
func (t *CreateTask) Stop(ctx *context.ExecuteContext) error {
	return nil
}
