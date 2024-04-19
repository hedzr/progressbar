# progressbar (-go)

[![Go](https://github.com/hedzr/progressbar/actions/workflows/go.yml/badge.svg)](https://github.com/hedzr/progressbar/actions/workflows/go.yml)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/hedzr/progressbar.svg?label=release)](https://github.com/hedzr/progressbar/releases)
[![go.dev](https://img.shields.io/badge/go-dev-green)](https://pkg.go.dev/github.com/hedzr/progressbar)

An asynchronous, multitask console/terminal progressbar widget. The main look of default stepper is:

![stepper-0](https://github.com/hzimg/blog-pics/blob/master/Picsee/stepper-0.mov.webp?raw=true)

Its original sample is pip installing ui, or python [rich](https://github.com/Textualize/rich)-like progressbar.

> To simplify our maintaining jobs, this repo was only tested at go1.18+.

## History

- v1.1.8
  - security patch: upgrade golang.org/x/net to 0.23.0

- v1.1.7
  - security patch: upgrade golang.org/x/crypto to 0.17.0

- v1.1.6
  - security patch: upgrade golang.org/x/net to 0.17.0

- v1.1.5
  - security patch: upgrade deps for vuln report on golang.org/x/net

- v1.1.3
  - improving coding style, and more docs
  - allow user-defined data packaged and applied to bar building: `SchemaData.Data any` added

- v1.1.1
  - fixed the minor display matters
  - added `WithBarIndentChars(s)`, `WithBarAppendText(s)`, `WithBarPrependText(s)`, and `WithBarExtraTailSpaces(howMany)`
  - added `WithBarOnDataPrepared(cb)` so you can observe and postprocess the data provided to bar layout template.

- v1.1.0
  - fixed possible broken output in escape sequences
  - fixed formatting and calculating when i made it public
  - fixed setting schema when i made it public
  - fixed data race posibility when using shared CPT tool
  - added `schema` sample app to show you how to customize me

- v1.0.0
  - first release

## Guide

`progressbar` provides a friendly interface to make things simple,
for creating the tasks with a terminal progressbar.

It assumes you're commonly running several asynchronous tasks with
rich terminal UI progressing display. The progressing UI can
be a bar (called `Stepper`) or a spinner.

A demo of `multibar` looks like:

![anim](https://github.com/hzimg/blog-pics/blob/master/Picsee/Screen%20Recording%202023-01-20%20at%2018.52.29.webp?raw=true)

### What's Steppers

Stepper style is like a horizontal bar with progressing tick(s).

```bash
go run ./examples/steppers
go run ./examples/steppers 0 # can be 0..3 (=progressbar.MaxSteppers())
```

### What's Spinners

Spinner style is a rotating icon/text in generally.

```bash
go run ./examples/spinners
go run ./examples/spinners 0 # can be 0..75 (=progressbar.MaxSpinners())
```

### Tasks & With groups

#### Using Tasks

By using `progressbar.NewTasks()`, you can add new task
bundled with a progressbar.

```go
func forAllSpinners() {
	tasks := progressbar.NewTasks(progressbar.New())
	defer tasks.Close()

	for i := whichSpinner; i < whichSpinner+5; i++ {
		tasks.Add(
			progressbar.WithTaskAddBarOptions(
				progressbar.WithBarSpinner(i),
				progressbar.WithBarUpperBound(100),
				progressbar.WithBarWidth(8),
				// progressbar.WithBarTextSchema(schema),
			),
			progressbar.WithTaskAddBarTitle(fmt.Sprintf("Task %v", i)),
			progressbar.WithTaskAddOnTaskProgressing(func(bar progressbar.PB, exitCh <-chan struct{}) {
				for max, ix := bar.UpperBound(), int64(0); ix < max; ix++ {
					ms := time.Duration(200 + rand.Intn(1800)) //nolint:gosec //just a demo
					time.Sleep(time.Millisecond * ms)
					bar.Step(1)
				}
			}),
		)
	}

	tasks.Wait() // start waiting for all tasks completed gracefully
}
```

To have a see to run:

```bash
go run ./examples/tasks
```

#### Write Your Own Tasks With `MultiPB` and `PB`

The above sample shows you how a `Task` could be encouraged by
`progressbar.WithTaskAddOnTaskProgressing`, `WithTaskAddOnTaskInitializing`
and `WithTaskAddOnTaskCompleted`.

You can write your `Task` and feedback the progress to multi-pbar
(`MultiPB`) or pbar (`PB`), see the source code `taskdownload.go`.

The key point is, wrapping your task runner, maybe called as worker,
as a `PB.Worker`, and add it with `WithBarWorker`.

<details>
<summary> Expand to get implementations </summary>

```go
func (s *DownloadTasks) Add(url, filename string, opts ...Opt) {
	task := new(aTask)
	task.wg = &s.wg
	task.url = url
	task.fn = filename

	var o []Opt
	o = append(o,
		WithBarWorker(task.doWorker),
		WithBarOnCompleted(task.onCompleted),
		WithBarOnStart(task.onStart),
	)
	o = append(o, opts...)

	s.bar.Add(
		100,
		task.fn, // fmt.Sprintf("downloading %v", s.fn),
		// // WithBarSpinner(14),
		// // WithBarStepper(3),
		// WithBarStepper(0),
		// WithBarWorker(s.doWorker),
		// WithBarOnCompleted(s.onCompleted),
		// WithBarOnStart(s.onStart),
		o...,
	)

	s.wg.Add(1)
}

func (s *aTask) doWorker(bar PB, exitCh <-chan struct{}) {
	// _, _ = io.Copy(s.w, s.resp.Body)

	for {
		n, err := s.resp.Body.Read(s.buf)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("Error: %v", err)
			return
		}
		if n == 0 {
			break
		}

		if _, err = s.w.Write(s.buf[:n]); err != nil {
			log.Printf("Error: %v", err)
			return
		}

		select {
		case <-exitCh:
			return
		default: // avoid block at <-exitCh
		}

		// time.Sleep(time.Millisecond * 100)
	}
}

func (s *aTask) onCompleted(bar PB) {
	wg := s.wg
	s.wg = nil
	wg.Done()
	atomic.AddInt32(&s.doneCount, 1)
}

func (s *aTask) onStart(bar PB) {
	if s.req == nil {
		var err error
		s.req, err = http.NewRequest("GET", s.url, nil) //nolint:gocritic
		if err != nil {
			log.Printf("Error: %v", err)
		}
		s.f, err = os.OpenFile(s.fn, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		s.resp, err = http.DefaultClient.Do(s.req)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		bar.UpdateRange(0, s.resp.ContentLength)

		s.w = io.MultiWriter(s.f, bar)

		const BUFFERSIZE = 4096
		s.buf = make([]byte, BUFFERSIZE)
	}
}
```

</details>

#### Multiple Bars (and Multiple groups)

For using `Stepper` instead of `Spinner`, these fragments can be applied:

```go
tasks.Add(url, fn,
	progressbar.WithBarStepper(whichStepper),
)
```

If you're looking for a downloader with progress bar, our
`progressbar.NewDownloadTasks` is better choice because it
had wrapped all things in one.

To start many groups of tasks like `docker pull` to get the
layers, just add them:

```go
func doEachGroup(group []string) {
	tasks := progressbar.NewDownloadTasks(progressbar.New())
	defer tasks.Close()

	for _, ver := range group {
		url := "https://dl.google.com/go/go" + ver + ".src.tar.gz" // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		fn := "go" + ver + ".src.tar.gz"                           // fn := fmt.Sprintf("go%v.src.tar.gz", ver)
		tasks.Add(url, fn,
			progressbar.WithBarStepper(whichStepper),
		)
	}
	tasks.Wait() // start waiting for all tasks completed gracefully
}

func downloadGroups() {
	for _, group := range [][]string{
		{"1.14.2", "1.15.1"},           # first group,
		{"1.16.1", "1.17.1", "1.18.3"}, # and the second one,
	} {
		doEachGroup(group)
	}
}
```

Run it(s):

```bash
go run ./examples/multibar
go run ./examples/multibar 3 # to select a stepper

# Or using spinner style
go run ./examples/multibar_spinner
go run ./examples/multibar_spinner 7 # to select a spinners
```

### Customize the bar layout

The default bar layout of a stepper is

```go
// see it in stepper.go
var defaultSchema = `{{.Indent}}{{.Prepend}} {{.Bar}} {{.Percent}} | <font color="green">{{.Title}}</font> | {{.Current}}/{{.Total}} {{.Speed}} {{.Elapsed}} {{.Append}}`
```

But you can always replace it with your own. A sample is `examples/tasks`. The demo app shows the real way:

```go
package main

const schema = `{{.Indent}}{{.Prepend}} {{.Bar}} {{.Percent}} | <b><font color="green">{{.Title}}</font></b> {{.Append}}`

tasks.Add(
  progressbar.WithTaskAddBarOptions(
    progressbar.WithBarUpperBound(100),
    //progressbar.WithBarSpinner(i),       // if you're looking for a spinner instead stepper
    //progressbar.WithBarWidth(8),
    progressbar.WithBarStepper(0),
    progressbar.WithBarTextSchema(schema), // change the bar layout here
  ),
  // ...
  progressbar.WithTaskAddBarTitle(fmt.Sprintf("Task %v", i)),
  progressbar.WithTaskAddOnTaskProgressing(func(bar progressbar.PB, exitCh <-chan struct{}) {
    for max, ix := bar.UpperBound(), int64(0); ix < max; ix++ {
      ms := time.Duration(20 + rand.Intn(500)) //nolint:gosec //just a demo
      time.Sleep(time.Millisecond * ms)
      bar.Step(1)
    }
  }),
)
```

Simple html tags (b, i, u, font, strong, em, cite, mark, del, kbd, code, html, head, body) can be embedded if ANSI Escaped Color codes is hard to use.

The predefined named colors are also available:

```go
// These color names can be used in <font color=''> html tag:
cptCM = map[string]int{
	"black":     FgBlack,
	"red":       FgRed,
	"green":     FgGreen,
	"yellow":    FgYellow,
	"blue":      FgBlue,
	"magenta":   FgMagenta,
	"cyan":      FgCyan,
	"lightgray": FgLightGray, "light-gray": FgLightGray,
	"darkgray": FgDarkGray, "dark-gray": FgDarkGray,
	"lightred": FgLightRed, "light-red": FgLightRed,
	"lightgreen": FgLightGreen, "light-green": FgLightGreen,
	"lightyellow": FgLightYellow, "light-yellow": FgLightYellow,
	"lightblue": FgLightBlue, "light-blue": FgLightBlue,
	"lightmagenta": FgLightMagenta, "light-magenta": FgLightMagenta,
	"lightcyan": FgLightCyan, "light-cyan": FgLightCyan,
	"white": FgWhite,
}
```

> `tool.GetCPT()` returns a `ColorTranslater` to help you strips the
> basic HTML tags and render them with ANSI escape sequences.

If you wanna build a better Percent or Elapsed, try formatting with `PercentFloat` and `ElapsedTime` field:

```go
const schema = `{{.PercentFloat|printf "%3.1f%%" }},  {{.ElapsedTime}}`
```

To observe the supplied data to the schema, try `WithBarOnDataPrepared(cb)`:

```go
tasks.Add(
	progressbar.WithTaskAddBarOptions(
		progressbar.WithBarStepper(i),
		progressbar.WithBarUpperBound(100),
		progressbar.WithBarWidth(32),
		progressbar.WithBarTextSchema(schema),
		progressbar.WithBarExtraTailSpaces(16),
		progressbar.WithBarPrependText("[[[x]]]"),
		progressbar.WithBarAppendText("[[[z]]]"),
		progressbar.WithBarOnDataPrepared(func(bar progressbar.PB, data *progressbar.SchemaData) {
			data.ElapsedTime *= 2
		}),
	),
	progressbar.WithTaskAddBarTitle("Task "+strconv.Itoa(i)), // fmt.Sprintf("Task %v", i)),
	progressbar.WithTaskAddOnTaskProgressing(func(bar progressbar.PB, exitCh <-chan struct{}) {
		for max, ix := bar.UpperBound(), int64(0); ix < max; ix++ {
			ms := time.Duration(10 + rand.Intn(300)) //nolint:gosec //just a demo
			time.Sleep(time.Millisecond * ms)
			bar.Step(1)
		}
	}),
)
```

> The API to change a spinner's display layout is same to above.

## Using `cursor` lib

There is a tiny terminal cursor operating subpackage, `cursor`. It's cross-platforms to `show and hide cursor`, `move cursor up, left` with/out wipe out the characters. Notes that is not a `TUI` cursor controlling library.

## Tips

To review all possible looks, try our samples:

```bash
# To run all stocked steppers in a screen
go run ./examples/steppers
# To run certain a stepper
go run ./examples/steppers 0

# To run all stocked spinners in a screen
go run ./examples/spinners
# To run certain a stepper
go run ./examples/spinners 0
```

## Credit

This repo is inspired from python3 install tui, and
[schollz/progressbar](https://github.com/schollz/progressbar), and more tui progress bars.

## LICENSE

Apache 2.0
