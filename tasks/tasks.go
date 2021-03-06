package tasks

import (
	"fmt"
	"strings"

	"github.com/dnephin/dobi/config"
	"github.com/dnephin/dobi/execenv"
	"github.com/dnephin/dobi/logging"
	"github.com/dnephin/dobi/tasks/alias"
	"github.com/dnephin/dobi/tasks/common"
	"github.com/dnephin/dobi/tasks/compose"
	"github.com/dnephin/dobi/tasks/context"
	"github.com/dnephin/dobi/tasks/iface"
	"github.com/dnephin/dobi/tasks/image"
	"github.com/dnephin/dobi/tasks/mount"
	"github.com/dnephin/dobi/tasks/run"
	"github.com/dnephin/dobi/utils/stack"
	docker "github.com/fsouza/go-dockerclient"
)

// TaskCollection is a collection of Task objects
type TaskCollection struct {
	tasks []iface.Task
}

func (c *TaskCollection) add(task iface.Task, resource config.Resource) {
	c.tasks = append(c.tasks, task)
}

func (c *TaskCollection) contains(name string) bool {
	for _, task := range c.tasks {
		if task.Name() == name {
			return true
		}
	}
	return false
}

// All returns all the tasks in the dependency order
func (c *TaskCollection) All() []iface.Task {
	return c.tasks
}

// Reversed returns all the tasks in reversed dependency order
func (c *TaskCollection) Reversed() []iface.Task {
	tasks := []iface.Task{}
	for i := len(c.tasks) - 1; i >= 0; i-- {
		tasks = append(tasks, c.tasks[i])
	}
	return tasks
}

func newTaskCollection() *TaskCollection {
	return &TaskCollection{}
}

func collectTasks(options RunOptions, execEnv *execenv.ExecEnv) (*TaskCollection, error) {
	return collect(options, &collectionState{
		newTaskCollection(),
		stack.NewStringStack(),
		newResourceResolver(execEnv),
	})
}

type collectionState struct {
	tasks     *TaskCollection
	taskStack *stack.StringStack
	resolver  *ResourceResolver
}

func collect(options RunOptions, state *collectionState) (*TaskCollection, error) {
	for _, name := range options.Tasks {
		if state.tasks.contains(name) {
			continue
		}

		if state.taskStack.Contains(name) {
			return nil, fmt.Errorf(
				"Invalid dependency cycle: %s", strings.Join(state.taskStack.Items(), ", "))
		}

		name, action := common.SplitTaskActionName(name)
		resource, ok := options.Config.Resources[name]
		if !ok {
			return nil, fmt.Errorf("Resource %q does not exist", name)
		}

		resource, err := state.resolver.Resolve(name, resource)
		if err != nil {
			return nil, err
		}

		task, err := buildTaskFromResource(name, action, resource)
		if err != nil {
			return nil, err
		}
		state.taskStack.Push(name)
		options.Tasks = task.Dependencies()
		if _, err := collect(options, state); err != nil {
			return nil, err
		}
		state.tasks.add(task, resource)
		state.taskStack.Pop()
	}
	return state.tasks, nil
}

// ResourceResolver is used to resolve variables in a resource config, and cache
// the result of the resolution
type ResourceResolver struct {
	execEnv *execenv.ExecEnv
	cache   map[string]config.Resource
}

// Resolve calls Resolve on resources and caches the resolved resource
func (r *ResourceResolver) Resolve(name string, res config.Resource) (config.Resource, error) {
	var err error
	resolved, ok := r.cache[name]
	if ok {
		return resolved, nil
	}
	resolved, err = res.Resolve(r.execEnv)
	if err == nil {
		r.cache[name] = resolved
	}
	return resolved, err
}

func newResourceResolver(execEnv *execenv.ExecEnv) *ResourceResolver {
	return &ResourceResolver{execEnv: execEnv, cache: make(map[string]config.Resource)}
}

// TODO: some way to make this a registry
func buildTaskFromResource(name, action string, resource config.Resource) (iface.Task, error) {
	switch conf := resource.(type) {
	case *config.ImageConfig:
		return image.GetTask(name, action, conf)
	case *config.RunConfig:
		return run.GetTask(name, action, conf)
	case *config.MountConfig:
		return mount.GetTask(name, action, conf)
	case *config.AliasConfig:
		return alias.GetTask(name, action, conf)
	case *config.ComposeConfig:
		return compose.GetTask(name, action, conf)
	default:
		panic(fmt.Sprintf("Unexpected config type %T", conf))
	}

}

func executeTasks(ctx *context.ExecuteContext, tasks *TaskCollection) error {
	defer func() {
		logging.Log.Debug("stopping tasks")
		for _, task := range tasks.Reversed() {
			if err := task.Stop(ctx); err != nil {
				logging.Log.Warnf("Failed to stop task %q: %s", task.Name(), err)
			}
		}
	}()

	logging.Log.Debug("executing tasks")
	for _, task := range tasks.All() {
		if err := task.Run(ctx); err != nil {
			return fmt.Errorf("Failed to execute task %q: %s", task.Name(), err)
		}
	}
	return nil
}

// RunOptions are the options supported by Run
type RunOptions struct {
	Client *docker.Client
	Config *config.Config
	Tasks  []string
	Quiet  bool
}

func getTaskNames(options RunOptions) []string {
	if len(options.Tasks) > 0 {
		return options.Tasks
	}

	if options.Config.Meta.Default != "" {
		return []string{options.Config.Meta.Default}
	}

	return options.Tasks
}

// Run one or more tasks
func Run(options RunOptions) error {
	options.Tasks = getTaskNames(options)
	if len(options.Tasks) == 0 {
		return fmt.Errorf("No task to run, and no default task defined.")
	}

	execEnv, err := execenv.NewExecEnvFromConfig(
		options.Config.Meta.ExecID,
		options.Config.Meta.Project,
		options.Config.WorkingDir,
	)
	if err != nil {
		return err
	}

	tasks, err := collectTasks(options, execEnv)
	if err != nil {
		return err
	}

	ctx := context.NewExecuteContext(options.Config, options.Client, execEnv, options.Quiet)
	return executeTasks(ctx, tasks)
}
