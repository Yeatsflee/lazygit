package tasks

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func getCounter() (func(), func() int) {
	counter := 0
	return func() { counter++ }, func() int { return counter }
}

type observedTask struct {
	*gocui.FakeTask
	done chan struct{}
}

func (task *observedTask) Done() {
	task.FakeTask.Done()
	close(task.done)
}

func newObservedManager() (*ViewBufferManager, <-chan *observedTask) {
	tasks := make(chan *observedTask, 4)
	return NewViewBufferManager(
		utils.NewDummyLog(), io.Discard,
		func() {}, func() {}, func() {}, func() {},
		func() gocui.Task {
			task := &observedTask{FakeTask: gocui.NewFakeTask(), done: make(chan struct{})}
			tasks <- task
			return task
		},
	), tasks
}

func awaitTaskSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(5 * time.Second):
		assert.FailNow(t, "timed out waiting for task")
	}
}

func TestInitialContentCallback(t *testing.T) {
	for _, explicitReady := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit ready %v", explicitReady), func(t *testing.T) {
			manager, tasks := newObservedManager()
			proceed := make(chan struct{})
			var task *observedTask
			callbackCount := 0
			content := ""
			assert.NoError(t, manager.NewTask(func(opts TaskOpts) error {
				<-proceed
				content = "ready"
				if explicitReady {
					opts.InitialContentLoaded()
					opts.InitialContentLoaded()
				}
				return nil
			}, "content", func() {
				callbackCount++
				assert.Equal(t, "ready", content)
				assert.Equal(t, gocui.TaskStatusBusy, task.Status())
			}))
			task = <-tasks
			close(proceed)
			awaitTaskSignal(t, task.done)
			manager.Close()
			assert.Equal(t, 1, callbackCount)
			assert.Equal(t, gocui.TaskStatusDone, task.Status())
		})
	}
}

func TestSupersededTaskDoesNotReportInitialContent(t *testing.T) {
	manager, tasks := newObservedManager()
	started := make(chan struct{})
	finish := make(chan struct{})
	var callbacks atomic.Int32

	assert.NoError(t, manager.NewTask(func(opts TaskOpts) error {
		close(started)
		<-finish
		// The replacement has been requested but has not stopped this task yet.
		opts.InitialContentLoaded()
		return nil
	}, "old", func() { callbacks.Add(1) }))
	oldTask := <-tasks
	awaitTaskSignal(t, started)

	manager.waitingMutex.Lock()
	assert.NoError(t, manager.NewTask(func(TaskOpts) error { return nil }, "new", func() {
		callbacks.Add(10)
	}))
	newTask := <-tasks
	close(finish)
	awaitTaskSignal(t, oldTask.done)
	manager.waitingMutex.Unlock()
	awaitTaskSignal(t, newTask.done)
	manager.Close()
	assert.Equal(t, int32(10), callbacks.Load())
}

func TestQueuedTasksRespectRequestOrder(t *testing.T) {
	manager, tasks := newObservedManager()
	var callbacks atomic.Int32
	manager.waitingMutex.Lock()
	assert.NoError(t, manager.NewTask(func(TaskOpts) error {
		assert.Fail(t, "superseded task must not start")
		return nil
	}, "old", func() { callbacks.Add(1) }))
	oldTask := <-tasks
	assert.NoError(t, manager.NewTask(func(TaskOpts) error { return nil }, "new", func() {
		callbacks.Add(10)
	}))
	newTask := <-tasks
	manager.waitingMutex.Unlock()
	awaitTaskSignal(t, oldTask.done)
	awaitTaskSignal(t, newTask.done)
	manager.Close()
	assert.Equal(t, int32(10), callbacks.Load())
}

func TestClosedTaskDoesNotReportInitialContent(t *testing.T) {
	manager, tasks := newObservedManager()
	started := make(chan struct{})
	var callbacks atomic.Int32
	assert.NoError(t, manager.NewTask(func(opts TaskOpts) error {
		close(started)
		<-opts.Stop
		opts.InitialContentLoaded()
		return nil
	}, "content", func() { callbacks.Add(1) }))
	task := <-tasks
	awaitTaskSignal(t, started)
	manager.Close()
	awaitTaskSignal(t, task.done)
	assert.Zero(t, callbacks.Load())
}

func TestFailedTaskDoesNotReportInitialContent(t *testing.T) {
	manager, tasks := newObservedManager()
	var callbacks atomic.Int32
	assert.NoError(t, manager.NewTask(func(TaskOpts) error {
		return errors.New("failed to load content")
	}, "content", func() { callbacks.Add(1) }))
	task := <-tasks
	awaitTaskSignal(t, task.done)
	manager.Close()
	assert.Zero(t, callbacks.Load())
}

func TestNewCmdTaskInstantStop(t *testing.T) {
	writer := bytes.NewBuffer(nil)
	beforeStart, getBeforeStartCallCount := getCounter()
	refreshView, getRefreshViewCallCount := getCounter()
	onEndOfInput, getOnEndOfInputCallCount := getCounter()
	onNewKey, getOnNewKeyCallCount := getCounter()
	onDone, getOnDoneCallCount := getCounter()
	task := gocui.NewFakeTask()
	newTask := func() gocui.Task {
		return task
	}

	manager := NewViewBufferManager(
		utils.NewDummyLog(),
		writer,
		beforeStart,
		refreshView,
		onEndOfInput,
		onNewKey,
		newTask,
	)

	stop := make(chan struct{})
	reader := bytes.NewBufferString("test")
	start := func() (*exec.Cmd, io.Reader) {
		// not actually starting this because it's not necessary
		cmd := exec.Command("blah")

		close(stop)

		return cmd, reader
	}

	fn := manager.NewCmdTask(start, "prefix\n", LinesToRead{20, -1, nil}, onDone)

	_ = fn(TaskOpts{Stop: stop, InitialContentLoaded: func() { task.Done() }})

	callCountExpectations := []struct {
		expected int
		actual   int
		name     string
	}{
		{0, getBeforeStartCallCount(), "beforeStart"},
		{1, getRefreshViewCallCount(), "refreshView"},
		{0, getOnEndOfInputCallCount(), "onEndOfInput"},
		{0, getOnNewKeyCallCount(), "onNewKey"},
		{1, getOnDoneCallCount(), "onDone"},
	}
	for _, expectation := range callCountExpectations {
		if expectation.actual != expectation.expected {
			t.Errorf("expected %s to be called %d times, got %d", expectation.name, expectation.expected, expectation.actual)
		}
	}

	if task.Status() != gocui.TaskStatusDone {
		t.Errorf("expected task status to be 'done', got '%s'", task.FormatStatus())
	}

	expectedContent := ""
	actualContent := writer.String()
	if actualContent != expectedContent {
		t.Errorf("expected writer to receive the following content: \n%s\n. But instead it received: %s", expectedContent, actualContent)
	}
}

func TestNewCmdTask(t *testing.T) {
	writer := bytes.NewBuffer(nil)
	beforeStart, getBeforeStartCallCount := getCounter()
	refreshView, getRefreshViewCallCount := getCounter()
	onEndOfInput, getOnEndOfInputCallCount := getCounter()
	onNewKey, getOnNewKeyCallCount := getCounter()
	onDone, getOnDoneCallCount := getCounter()
	task := gocui.NewFakeTask()
	newTask := func() gocui.Task {
		return task
	}

	manager := NewViewBufferManager(
		utils.NewDummyLog(),
		writer,
		beforeStart,
		refreshView,
		onEndOfInput,
		onNewKey,
		newTask,
	)

	stop := make(chan struct{})
	reader := bytes.NewBufferString("test")
	start := func() (*exec.Cmd, io.Reader) {
		// not actually starting this because it's not necessary
		cmd := exec.Command("blah")

		return cmd, reader
	}

	fn := manager.NewCmdTask(start, "prefix\n", LinesToRead{20, -1, nil}, onDone)
	wg := sync.WaitGroup{}
	wg.Go(func() {
		time.Sleep(100 * time.Millisecond)
		close(stop)
	})
	_ = fn(TaskOpts{Stop: stop, InitialContentLoaded: func() { task.Done() }})

	wg.Wait()

	callCountExpectations := []struct {
		expected int
		actual   int
		name     string
	}{
		{1, getBeforeStartCallCount(), "beforeStart"},
		{1, getRefreshViewCallCount(), "refreshView"},
		{1, getOnEndOfInputCallCount(), "onEndOfInput"},
		{0, getOnNewKeyCallCount(), "onNewKey"},
		{1, getOnDoneCallCount(), "onDone"},
	}
	for _, expectation := range callCountExpectations {
		if expectation.actual != expectation.expected {
			t.Errorf("expected %s to be called %d times, got %d", expectation.name, expectation.expected, expectation.actual)
		}
	}

	if task.Status() != gocui.TaskStatusDone {
		t.Errorf("expected task status to be 'done', got '%s'", task.FormatStatus())
	}

	expectedContent := "prefix\ntest\n"
	actualContent := writer.String()
	if actualContent != expectedContent {
		t.Errorf("expected writer to receive the following content: \n%s\n. But instead it received: %s", expectedContent, actualContent)
	}
}

// A dummy reader that simply yields as many blank lines as requested. The only
// thing we want to do with the output is count the number of lines.
type BlankLineReader struct {
	totalLinesToYield int
	linesYielded      int
}

func (d *BlankLineReader) Read(p []byte) (n int, err error) {
	if d.totalLinesToYield == d.linesYielded {
		return 0, io.EOF
	}

	d.linesYielded++
	p[0] = '\n'
	return 1, nil
}

func TestNewCmdTaskRefresh(t *testing.T) {
	type scenario struct {
		name                        string
		totalTaskLines              int
		linesToRead                 LinesToRead
		expectedLineCountsOnRefresh []int
	}

	scenarios := []scenario{
		{
			"total < initialRefreshAfter",
			150,
			LinesToRead{100, 120, nil},
			[]int{100},
		},
		{
			"total == initialRefreshAfter",
			150,
			LinesToRead{100, 100, nil},
			[]int{100},
		},
		{
			"total > initialRefreshAfter",
			150,
			LinesToRead{100, 50, nil},
			[]int{50, 100},
		},
		{
			"initialRefreshAfter == -1",
			150,
			LinesToRead{100, -1, nil},
			[]int{100},
		},
		{
			"totalTaskLines < initialRefreshAfter",
			25,
			LinesToRead{100, 50, nil},
			[]int{25},
		},
		{
			"totalTaskLines between total and initialRefreshAfter",
			75,
			LinesToRead{100, 50, nil},
			[]int{50, 75},
		},
	}

	for _, s := range scenarios {
		writer := bytes.NewBuffer(nil)
		lineCountsOnRefresh := []int{}
		refreshView := func() {
			lineCountsOnRefresh = append(lineCountsOnRefresh, strings.Count(writer.String(), "\n"))
		}

		task := gocui.NewFakeTask()
		newTask := func() gocui.Task {
			return task
		}

		manager := NewViewBufferManager(
			utils.NewDummyLog(),
			writer,
			func() {},
			refreshView,
			func() {},
			func() {},
			newTask,
		)

		stop := make(chan struct{})
		reader := BlankLineReader{totalLinesToYield: s.totalTaskLines}
		start := func() (*exec.Cmd, io.Reader) {
			// not actually starting this because it's not necessary
			cmd := exec.Command("blah")

			return cmd, &reader
		}

		fn := manager.NewCmdTask(start, "", s.linesToRead, func() {})
		wg := sync.WaitGroup{}
		wg.Go(func() {
			time.Sleep(100 * time.Millisecond)
			close(stop)
		})
		_ = fn(TaskOpts{Stop: stop, InitialContentLoaded: func() { task.Done() }})

		wg.Wait()

		if !reflect.DeepEqual(lineCountsOnRefresh, s.expectedLineCountsOnRefresh) {
			t.Errorf("%s: expected line counts on refresh: %v, got %v",
				s.name, s.expectedLineCountsOnRefresh, lineCountsOnRefresh)
		}
	}
}
