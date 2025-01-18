# multibar2

For enabling the new feature `Groupable`, `multibar2` app demonstrates how to coding with `MultiPB2`.

## What's Groupable

This new feature added a title to the front of each group, and allowes groups (jobs) running asynchronously or synchronously.

## How To

The new feature (`Groupable with title`) needs a type assertion generally:

```go
mpb := progressbar.New().(progressbar.GroupedPB)
defer mpb.Close() // cleanup
```

`GroupedPB` enables `AddToGroup(groupTitle, totalTicks, taskTitle, opts...)` to allow adding group title and task.

`doEachGroupWithTasks` is a common function to show how to do it: 

```go
func doEachGroupWithTasks(mpb progressbar.GroupedPB, wg *sync.WaitGroup, group groupedJobs) {
	if mpb == nil {
		mpb := progressbar.New().(progressbar.GroupedPB)
		defer mpb.Close() // cleanup
	}
	if wg == nil {
		wg = &sync.WaitGroup{}
		defer wg.Wait() // waiting for all tasks done.
	}

	var jobs []*Job

	for _, ver := range group.group {
		url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
		job := &Job{Url: url1, mpb: mpb, wg: wg}

		job.totalTicks = int64(3500 + int(rand.Int31n(2000))) // 3500ms
		job.index = mpb.AddToGroup(
			group.Title(),
			job.totalTicks,
			url1.Title(),
			progressbar.WithBarOnStart(job.onStart),
			progressbar.WithBarWorker(job.doWorker),
			progressbar.WithBarOnCompleted(job.onCompleted),
			progressbar.WithBarStepper(whichStepper),
			progressbar.WithBarStepperPostInit(func(bar progressbar.BarT) {
				bar.SetHighlightColor(tool.FgDarkGray)
			}),
		)

		wg.Add(1)
		jobs = append(jobs, job)
		job.Start()
	}
}

type groupedJobs struct {
group []string
title string
}

func (s groupedJobs) Title() string { return s.title }
```

In a higher layer, there are several ways to add your tasks. The differences of them are how to organize the task detail and how to sent it into MPBar.

> Both the three ways will call into `doEachGroupWithTasks` to commit each tasks.

### Way 1

```go
func downloadGroups1Worked() {
	mpb := progressbar.NewGPB()
	defer mpb.Close() // cleanup

	var wg sync.WaitGroup
	defer wg.Wait() // waiting for all tasks done.

	for _, group := range []groupedJobs{
		{[]string{"1.14.1", "1.15.1"}, "AAA"},
		{[]string{"1.16.1", "1.17.1", "1.18.1"}, "BBB"},
	} {
		doEachGroupWithTasks(mpb, &wg, group)
	}

	for _, group := range []groupedJobs{
		{[]string{"1.20.1", "1.21.1"}, "CCC"},
		{[]string{"1.22.1"}, "DDD"},
	} {
		doEachGroupWithTasks(mpb, &wg, group)
	}

	// mpb.Close() // cleanup
}
```

### Way 2

You can add each group separately, but the running effect has no difference. This is caused because we want a backward compatibility.

```go
func downloadGroups2Worked() {
	mpb := progressbar.New().(progressbar.GroupedPB)

	for _, group := range []groupedJobs{
		{[]string{"1.14.1", "1.15.1"}, "AAA"},
		{[]string{"1.16.1", "1.17.1", "1.18.1"}, "BBB"},
	} {
		doEachGroupWithTasks(mpb, nil, group)
	}

	mpb.Close() // cleanup
	mpb = nil

	mpb = progressbar.New().(progressbar.GroupedPB)

	for _, group := range []groupedJobs{
		{[]string{"1.20.1", "1.21.1"}, "CCC"},
		{[]string{"1.22.1"}, "DDD"},
	} {
		doEachGroupWithTasks(mpb, nil, group)
	}

	mpb.Close() // cleanup
}
```

### Way 3

Way 3 is a concise approach to add grouped task with titles: there's no need to initiate a pb object explicitly because `doEachGroupWithTasks()` could handle the case.

```go
func downloadGroups3Worked() {
	for _, group := range []groupedJobs{
		{[]string{"1.14.1", "1.15.1"}, "AAA"},
		{[]string{"1.16.1", "1.17.1", "1.18.1"}, "BBB"},
	} {
		doEachGroupWithTasks(nil, nil, group)
	}

	for _, group := range []groupedJobs{
		{[]string{"1.20.1", "1.21.1"}, "CCC"},
		{[]string{"1.22.1"}, "DDD"},
	} {
		doEachGroupWithTasks(nil, nil, group)
	}
}
```

## Post Words

The api is unstable now.

I'm still looking for a better api to help you manage terminal task such as downloading from a url, or invoking a background task/program with long time.

Please issue me or suggest to me for it.
