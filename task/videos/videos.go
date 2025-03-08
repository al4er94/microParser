package videos

import (
	"awesomeProject/common"
	"awesomeProject/config"
	"awesomeProject/repo"
	"awesomeProject/task"
	"log"
	"time"
)

type videoTask struct {
	closed chan struct{}
	ticker *time.Ticker
}

func (t *videoTask) Run() {
	if err := t.handler(); err != nil {
		log.Fatal("Err fatal task: ", err)
	}

	go func() {
		for {
			select {
			case <-t.closed:
				return
			case <-t.ticker.C:
				if err := t.handler(); err != nil {
					log.Println("Err task: ", err)
				}
			}
		}
	}()
}

func (t *videoTask) handler() error {
	log.Println(len(repo.RepoUrl))

	common.UpdateRepo(config.DB)

	return nil
}

func (t *videoTask) Stop() {
	close(t.closed)
}

func NewTask() task.Task {
	return &videoTask{
		closed: make(chan struct{}),
		ticker: time.NewTicker(1 * time.Minute),
	}
}
