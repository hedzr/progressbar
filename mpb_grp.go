package progressbar

import (
	"context"
	"strings"
	"sync/atomic"
)

func (s *GroupV2) AddDownloader(dad Repaintable, task string, d *DownloadTask, opts ...TaskBarOpt) (err error) {
	s.muTasks.Lock()
	defer s.muTasks.Unlock()

	if l, ok := dad.(Logger); ok && l != nil {
		d.logger = l.Logger()
	}

	d.wg = &s.wg
	s.wg.Add(1)

	var tsk *TaskBar
	if tsk, err = s.findTask(task); err != nil {
		tsk = &TaskBar{Name: task, downloader: d}
		tsk.dad = s.dad
		for _, opt := range opts {
			opt(tsk)
		}
		s.tasks = append(s.tasks, tsk)
		return nil
	}
	return errTaskExisted
}

func (s *GroupV2) AddTask(dad Repaintable, task string, min, max int64, job Job, opts ...TaskBarOpt) (err error) {
	s.muTasks.Lock()
	defer s.muTasks.Unlock()

	var tsk *TaskBar
	if tsk, err = s.findTask(task); err != nil {
		tsk = &TaskBar{Name: task, min: min, max: max, job: job}
		tsk.dad = s.dad
		for _, opt := range opts {
			opt(tsk)
		}
		s.tasks = append(s.tasks, tsk)
		return nil
	}
	return errTaskExisted
}

func (s *GroupV2) TaskByIndex(index int) *TaskBar {
	s.muTasks.RLock()
	defer s.muTasks.RUnlock()
	if index >= 0 && index < len(s.tasks) {
		return s.tasks[index]
	}
	return nil
}

func (s *GroupV2) TaskByName(task string) *TaskBar {
	s.muTasks.RLock()
	defer s.muTasks.RUnlock()
	for _, tsk := range s.tasks {
		if tsk.Name == task {
			return tsk
		}
	}
	return nil
}

func (s *GroupV2) AllDone() bool {
	s.muTasks.RLock()
	defer s.muTasks.RUnlock()
	return s.allDone()
}

func (s *GroupV2) allDone() bool {
	return int(atomic.LoadInt32(&s.done)) >= len(s.tasks)
}

func (s *GroupV2) runJobs(ctx context.Context, bar *MPBV2) (allDone bool) {
	s.muTasks.RLock()
	defer s.muTasks.RUnlock()
	_ = ctx

	defer func() {
		allDone = s.allDone()
		bar.Repaint()
	}()
	for _, tsk := range s.tasks {
		if tsk.job != nil {
			if progress, _, done := tsk.Done(); !done {
				if delta, err := tsk.job(bar, s, tsk, progress); err == nil {
					if done := tsk.Increase(delta); done {
						atomic.StoreInt64(&tsk.progress, tsk.Max())
						atomic.AddInt32(&s.done, 1)
					}
				}
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}
	return
}

func (s *GroupV2) findTask(task string) (tsk *TaskBar, err error) {
	for _, tsk = range s.tasks {
		if tsk.Name == task {
			return
		}
	}
	return nil, errNotFound
}

func (s *GroupV2) repaint(pc *paintCtx) (ignored bool) {
	if s.muTasks.TryRLock() {
		defer s.muTasks.RUnlock()

		if atomic.CompareAndSwapInt32(&s.titlePainted, 0, 1) {
			// if is.InTracing() {
			// 	println(fmt.Sprintf("%s (%v, %v, %v)", s.Name,
			// 		atomic.LoadInt32(&s.done), pc.lastDoneCount, pc.lastDone))
			// } else {
			// 	println(fmt.Sprintf("%s", s.Name))
			// }
			println(s.Name)
		}
		_ = pc

		var sb strings.Builder
		for i, tsk := range s.tasks {
			// tsk.repaint(&sb, i, pc)
			_, _ = sb.WriteString(tsk.stepper.String(tsk))
			_, _ = sb.WriteRune('\n')
			_ = i
		}

		s.block.Update(sb.String())
	} else {
		ignored = true
	}
	return
}
