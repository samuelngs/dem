package ext

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

// SetupTaskHandler for implementing installation or setup instructions
type SetupTaskHandler func(ProgressBar) error

// SetupTaskHandlers is the multiple setup task handlers type
type SetupTaskHandlers []SetupTaskHandler

// SetupTask type
type SetupTask struct {
	Status  string
	Handler SetupTaskHandler
	Options Options
}

// SetupTasks is the multiple setup task type
type SetupTasks []*SetupTask

// Setup to run extension setup tasks
func Setup(extensions ...Extension) error {

	// skip rendering progress view if all setup tasks are already completed
	var numBars int
	for _, extension := range extensions {
		if len(extension.SetupTasks()) > 0 {
			numBars++
		}
	}
	if numBars == 0 {
		return nil
	}

	// rendering progress view for setup tasks
	var (
		setupWg            = new(sync.WaitGroup)
		p                  = mpb.New(mpb.WithWidth(64), mpb.WithWaitGroup(setupWg))
		format             = " · %s  "
		taskLen, statusLen int
	)
	for _, extension := range extensions {
		if l := len(fmt.Sprintf(format, extension.String())); l > taskLen {
			taskLen = l
		}
		for _, task := range extension.SetupTasks() {
			if l := len(task.Status) + 1; l > statusLen {
				statusLen = l
			}
		}
	}
	for _, extension := range extensions {
		setupWg.Add(1)

		go func(extension Extension) {
			defer setupWg.Done()

			var (
				setupTasks   = extension.SetupTasks()
				numOfTasks   = len(setupTasks)
				progressBars = make([]*mpb.Bar, numOfTasks)
			)
			for j, setupTask := range setupTasks {
				var (
					task          = fmt.Sprintf(format, strings.ToLower(extension.String()))
					status        = setupTask.Status
					options       = make([]mpb.BarOption, 0)
					total   int64 = 100
				)

				if j > 0 {
					options = append(options, mpb.BarReplaceOnComplete(progressBars[j-1]))
				}

				if j == numOfTasks-1 {
					options = append(options, mpb.BarClearOnComplete())
					options = append(options, mpb.PrependDecorators(
						decor.Name(task, decor.WC{W: taskLen, C: decor.DidentRight}),
						decor.OnComplete(decor.Name(status, decor.WC{W: statusLen, C: decor.DidentRight}), setupTask.Options.CompleteMessage),
					))
					if setupTask.Options.ShowPercentage {
						options = append(options, mpb.AppendDecorators(
							decor.OnComplete(decor.Percentage(decor.WC{W: 5}), ""),
						))
					}
				} else {
					options = append(options, mpb.BarRemoveOnComplete())
					options = append(options, mpb.PrependDecorators(
						decor.Name(task, decor.WC{W: taskLen, C: decor.DidentRight}),
						decor.Name(status, decor.WC{W: statusLen, C: decor.DidentRight}),
					))
					if setupTask.Options.ShowPercentage {
						options = append(options, mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})))
					}
				}

				bar := p.AddBar(total, options...)
				progressBars[j] = bar

				err := setupTask.Handler(bar)
				for !bar.Completed() {
					bar.IncrBy(1)
				}

				if err != nil {
					repl := p.AddBar(1,
						mpb.BarClearOnComplete(),
						mpb.BarReplaceOnComplete(bar),
						mpb.PrependDecorators(
							decor.Name(task, decor.WC{W: taskLen, C: decor.DidentRight}),
							decor.OnComplete(decor.Name(status, decor.WC{W: statusLen, C: decor.DidentRight}), err.Error()),
						),
					)
					repl.IncrBy(1)
					break
				}
			}
		}(extension)
	}
	p.Wait()

	return nil
}

// Procedure creates new setup task
func Procedure(status string, handler SetupTaskHandler, opts ...Option) *SetupTask {
	task := &SetupTask{
		Status:  status,
		Handler: handler,
		Options: newOptions(opts...),
	}
	return task
}
