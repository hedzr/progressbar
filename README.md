# progressbar (-go)

[![Go](https://github.com/hedzr/progressbar/actions/workflows/go.yml/badge.svg)](https://github.com/hedzr/progressbar/actions/workflows/go.yml)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/hedzr/progressbar.svg?label=release)](https://github.com/hedzr/progressbar/releases)
[![go.dev](https://img.shields.io/badge/go-dev-green)](https://pkg.go.dev/github.com/hedzr/progressbar)

An asynchronous, multiple console/terminal progressbar widget. The main look of default stepper is:

![stepper-0](https://github.com/hzimg/blog-pics/blob/master/Picsee/stepper-0.mov.webp?raw=true)

Its original sample is pip installing ui.

## Guide

`progressbar` provides a friendly interface to make things simple,
for creating the tasks with a terminal progressbar.

It assumes you're running several asynchronous tasks with
rich terminal UI progressing display. The progressing UI can
be a bar (called `Stepper`) or a spinner.

A demo of `multibar` looks like:

![anim](https://github.com/hzimg/blog-pics/blob/master/Picsee/Screen%20Recording%202023-01-20%20at%2018.52.29.webp?raw=true)

### What's Steppers

Stepper style like a horizontal bar with progressing tick.

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

### Tasks & With Multiple groups

#### Using Tasks

By using `progressbar.NewTasks()`, you can add new task with a bundled progressbar.

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
				progressbar.WithBarTextSchema(schema),
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

	time.Sleep(5 * time.Millisecond)
	tasks.Wait()
}
```

To have a see to run:

```bash
go run ./examples/tasks
```

#### Multiple Bars

For using `Stepper` instead of `Spinner`, these fragments can be applied:

```go
tasks.Add(url, fn,
	progressbar.WithBarStepper(whichStepper),
)
```

To start many groups of tasks like `docker pull` to get the layers, just add them:

```go
func doEachGroup(group []string) {
	tasks := progressbar.NewDownloadTasks(progressbar.New())
	defer tasks.Close()

	for _, ver := range group {
		url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		fn := fmt.Sprintf("go%v.src.tar.gz", ver)

		tasks.Add(url, fn,
			progressbar.WithBarStepper(whichStepper),
		)
	}

	time.Sleep(5 * time.Millisecond)
	tasks.Wait()
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

Run it:

```bash
go run ./examples/multibar
go run ./examples/multibar 3 # to select a stepper

# Or using spinner style
go run ./examples/multibar_spinner
go run ./examples/multibar_spinner 3 # to select a stepper
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

The API to change a spinner's display layout is same to above.

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
