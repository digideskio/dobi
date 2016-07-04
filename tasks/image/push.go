package image

import (
	"fmt"
	"io"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/dnephin/dobi/config"
	"github.com/dnephin/dobi/logging"
	"github.com/dnephin/dobi/tasks/context"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	docker "github.com/fsouza/go-dockerclient"
)

// PushTask creates a Docker image
type PushTask struct {
	name   string
	config *config.ImageConfig
}

// NewPushTask creates a new PushTask object
func NewPushTask(name string, conf *config.ImageConfig) *PushTask {
	return &PushTask{name: name, config: conf}
}

// Name returns the name of the task
func (t *PushTask) Name() string {
	return t.name
}

func (t *PushTask) String() string {
	return fmt.Sprintf("image.PushTask(name=%s, config=%s)", t.name, t.config)
}

func (t *PushTask) logger() *log.Entry {
	return logging.Log.WithFields(log.Fields{"task": t})
}

// Repr formats the task for logging
func (t *PushTask) Repr() string {
	return fmt.Sprintf("[image:push %s] %s", t.name, t.config.Image)
}

// Run builds or pulls an image if it is out of date
func (t *PushTask) Run(ctx *context.ExecuteContext) error {
	t.logger().Debug("Run")

	t.logger().Info("Pushing")
	if err := t.pushTags(ctx); err != nil {
		return err
	}
	ctx.SetModified(t.name)
	t.logger().Info("Pushed")
	return nil
}

// TODO: can more code be shared with build() ?
func (t *PushTask) push(ctx *context.ExecuteContext, tag string) error {
	out := os.Stdout
	outFd, isTTY := term.GetFdInfo(out)
	rpipe, wpipe := io.Pipe()
	defer rpipe.Close()

	errChan := make(chan error)

	go func() {
		err := jsonmessage.DisplayJSONMessagesStream(rpipe, out, outFd, isTTY, nil)
		errChan <- err
	}()

	// TODO: support auth creds
	err := ctx.Client.PushImage(docker.PushImageOptions{
		Name:          t.config.Image,
		Tag:           tag,
		OutputStream:  wpipe,
		RawJSONStream: true,
		// TODO: do I need to set the registry?
		// TODO: timeout
	}, docker.AuthConfiguration{})
	wpipe.Close()
	if err != nil {
		<-errChan
		return err
	}

	return <-errChan
}

func (t *PushTask) pushTags(ctx *context.ExecuteContext) error {
	for _, tag := range t.config.Tags {
		if err := t.push(ctx, tag); err != nil {
			return err
		}
	}
	return nil
}

// Prepare the task
// TODO: move this to common function or maybe prepare should be based on the
// resource?
func (t *PushTask) Prepare(ctx *context.ExecuteContext) error {
	for _, tag := range t.config.Tags {
		if _, err := ctx.Env.Resolve(tag); err != nil {
			return err
		}
	}
	return nil
}

// Stop the task
func (t *PushTask) Stop(ctx *context.ExecuteContext) error {
	return nil
}
