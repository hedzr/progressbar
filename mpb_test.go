package progressbar

import (
	"context"
	"math/rand"
	"net/url"
	"path"
	"strconv"
	"testing"
	"time"
)

func TestNewV2(t *testing.T) {
	defaultMPB.Close()

	mpb := NewV2()
	defer mpb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	job := func(bar *MPBV2, grp *GroupV2, tsk *TaskBar, progress int64, args ...any) (delta int64, err error) {
		time.Sleep(time.Duration(rng.Intn(60)+30) * time.Millisecond)
		delta += int64(rng.Intn(5) + 1)
		return
	}

	versions := []string{"1.16.1", "1.17.1", "1.18.1", "1.23.1"}
	verIdx := 0

	total, num, numTasks := int64(100), 2, 3
	for i := range num {
		for j := 0; j < numTasks; j++ {
			if (j == numTasks-1 || i == 0) && verIdx < len(versions) {
				ver := versions[verIdx]
				url1 := TitledUrl("https://dl.google.com/go/go" + ver + ".src.tar.gz") // url := fmt.Sprintf("https://dl.google.com/go/go%v.src.tar.gz", ver)
				mpb.AddDownloadingBar(
					"Group "+strconv.Itoa(i), "Task #"+strconv.Itoa(j),
					&DownloadTask{
						Url:      url1.String(),
						Filename: url1.Title(),
						Title:    url1.Title(),
					},
				)
				verIdx++
				continue
			}
			mpb.AddBar("Group "+strconv.Itoa(i), "Task #"+strconv.Itoa(j), 0, total, job)
		}
	}

	// var wg sync.WaitGroup
	// wg.Add(num * numTasks)

	mpb.Run(ctx)
}

type TitledUrl string

func (t TitledUrl) String() string {
	return string(t)
}

func (t TitledUrl) Title() string {
	parse, err := url.Parse(string(t))
	if err != nil {
		return string(t)
	}
	return path.Base(parse.Path)
}
