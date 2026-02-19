# progressbar (-go)

[![Go](https://github.com/hedzr/progressbar/actions/workflows/go.yml/badge.svg)](https://github.com/hedzr/progressbar/actions/workflows/go.yml)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/hedzr/progressbar.svg?label=release)](https://github.com/hedzr/progressbar/releases)
[![go.dev](https://img.shields.io/badge/go-dev-green)](https://pkg.go.dev/github.com/hedzr/progressbar)

An asynchronous, multitask console/terminal progressbar widget. The main look of default stepper is:

![stepper-0](https://github.com/hzimg/blog-pics/blob/master/Picsee/stepper-0.mov.webp?raw=true)

Its original sample is pip installing ui, or python [rich](https://github.com/Textualize/rich)-like progressbar.

## History

### V2

We're happy to annouce the new `v2` released.

This is a seamless upgrade for your legacy v1 codes.

But new `NewV2(opts...)` will take a fully-rewritten progressbar and a better accurate behavior to you.

In our old v1 releases, some tasks ended but its bar stunned at 99.x%, we guess that you'd like to see a 100% fully-completed bar, right? The little trouble is gone now.

In other sides, we keep the codes under the control of reusing and isolations, which improvements would bring a most stable progressbar for the repeatable, reuseable jobs of yours.

- See the [CHANGELOG](https://github.com/hedzr/progressbar/blob/master/CHANGELOG)

**V2 need go toolchain 1.25+ now (since v2.1.0)**.

### ~~V1~~

Since v1.2.5, the minimal toolchain upgraded to go1.23.7.

Since v1.2, we upgrade and rewrite a new implementation of GPB so that we can provide grouped progressbar with titles. It stay in unstabled state but it worked for me. Sometimes you can rollback to v1.1.x to keep the old progrmatic logics unchanged.

- See the [CHANGELOG](https://github.com/hedzr/progressbar/blob/master/CHANGELOG)

## Guide

`progressbar` provides a friendly interface to make things simple,
for creating the tasks with a terminal progressbar.

It assumes you're commonly running several asynchronous tasks with
rich terminal UI progressing display. The progressing UI can
be a bar (called `Stepper`) or a spinner.

A demo of ~~`multibar`~~ (deprecated) looks like:

![anim](https://github.com/hzimg/blog-pics/blob/master/Picsee/Screen%20Recording%202023-01-20%20at%2018.52.29.webp?raw=true)

`mpbv2` is a sample app to show you more advanced usages.

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

### Using `progressbar.v2`

Since v2, we enable `NewV2()` to take a stable, accurate CLI progressbar to you:

```bash
import "github.com/hedzr/progressbar/v2"

func downloadGroupsV2Worked() {
	// const mySchema = `{{.Indent}}{{.Prepend}} <font color="green">{{.Title}}</font> {{.Percent}} {{.Bar}} {{.Current}}/{{.Total}} {{.Speed}} {{.Elapsed}} {{.Append}}`
	// var versions = []string{"1.16.1", "1.17.1", "1.18.1", "1.19.1", "1.20.1", "1.21.1", "1.22.1", "1.23.1", "1.24.1"}
	var versions = []string{"1.24.1"}

	var mpb *progressbar.MPBV2
	if schema := os.Getenv("SCHEMA"); schema != "" {
		mpb = progressbar.NewV2(progressbar.WithSchema(schema))
	} else {
		mpb = progressbar.NewV2()
	}
	defer mpb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// define a counter job here
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	job := func(bar *progressbar.MPBV2, grp *progressbar.GroupV2, tsk *progressbar.TaskBar, progress int64, args ...any) (delta int64, err error) {
		time.Sleep(time.Duration(rng.Intn(60)+30) * time.Millisecond)
		delta += int64(rng.Intn(5) + 1)
		return
	}

	// define the downloading job adder here
	verIdx := 0
	addDownloadJob := func(bar *progressbar.MPBV2, i, j int) {
		ver := versions[verIdx]
		url1 := TitledURL("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		bar.AddDownloadingBar(
			"Group "+strconv.Itoa(i), "Task #"+strconv.Itoa(j),
			&progressbar.DownloadTask{
				Url:      url1.String(),
				Filename: url1.Title(),
				Title:    url1.Title(),
			},
			// more opts can be set here
			// progressbar.WithTaskBarStepper(whichStepper),
			// progressbar.WithTaskBarSpinner(whichSpinner),
		)
		verIdx++
	}

	total, num, numTasks := int64(100), 2, 3
	// we would add some progressing task groups,
	for i := range num {
		// in a single task group, we add some tasks,
		for j := range numTasks {
			// one of which is a downloading task.
			if (j == numTasks-1 || i == 0) && verIdx < len(versions) {
				addDownloadJob(mpb, i, j)
				continue
			}
			// and the rests are counter tasks.
			mpb.AddBar("Group "+strconv.Itoa(i), "Task #"+strconv.Itoa(j), 0, total, counterJob)
		}
	}

	// var wg sync.WaitGroup
	// wg.Add(num * numTasks)

	// so you will get a multi-group multi-task progress bar by Run it.
	mpb.Run(ctx)
}
```

To try the above sample codes, run under a terminal:

```bash
go run ./examples/mpbv2
```

The notable thing is, once you've upgraded to `progressbar.v2`, the legacy v1 code would keep work without any changes.

## ~~v1 docs (DEPRECATED)~~

For the users' legacy codes, we kept v1 codes and added new codes and upgrade to v2.
So both v1 and v2 codes in the current release.

But the legacy codes have some known issues.

So we will cleanup these deprecated codes at a day when we release v2.2.

> For more docs, see [README.v1.md](./README.v1.md).

## Credit

This repo is inspired from python3 install tui, and
[schollz/progressbar](https://github.com/schollz/progressbar), and more tui progress bars.

## LICENSE

Apache 2.0
