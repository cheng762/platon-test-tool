package testcases

import (
	"fmt"

	"github.com/PlatONnetwork/platon-test-tool/common"

	"github.com/PlatONnetwork/PlatON-Go/core/types"

	"log"

	"github.com/robfig/cron"
)

const PrefixCase = "Case"

func NewCommonCases(jobNode *common.Node) *commonCases {
	c := new(commonCases)
	c.corn = cron.New()
	if _, err := c.corn.AddFunc("@every 1s", c.schedule); err != nil {
		panic(err)
	}
	c.jobNode = jobNode
	c.donech = make(chan struct{}, 0)
	c.errch = make(chan error, 0)
	c.errors = make([]error, 0)
	return c
}

type commonCases struct {
	corn    *cron.Cron
	jobs    []*job
	jobNode *common.Node
	errors  []error
	errch   chan error
	donech  chan struct{}
}

type job struct {
	handle func(block *types.Block, params ...interface{}) (bool, error)
	params []interface{}
	done   bool
	runing bool
	desc   string
}

func (j *job) run(block *types.Block) (bool, error) {
	return j.handle(block, j.params...)
}

func (c *commonCases) End() error {
	c.corn.Start()
	for {
		select {
		case err := <-c.errch:
			c.errors = append(c.errors, err)
		case <-c.donech:
			c.corn.Stop()
			close(c.errch)
			return nil
		}
	}
}

//添加需等待的测试，每隔1s将当前块高传入f中并执行
// f为需要执行的函数，当执行成功返回true，nil，当未达到执行条件返回false，nil，当执行失败返回false，error
// params为f所需要的参数
func (c *commonCases) addJobs(desc string, f func(block *types.Block, params ...interface{}) (bool, error), params ...interface{}) {
	var j *job
	j = new(job)
	j.params = params
	j.handle = f
	j.desc = desc
	c.jobs = append(c.jobs, j)
}

func (c *commonCases) schedule() {
	block, err := c.jobNode.BlockByNumber(nil)
	if err != nil {
		c.errch <- fmt.Errorf("job error: %v", err)
		c.donech <- struct{}{}
		close(c.donech)
		return
	}
	if block.Number().Uint64()%100 == 0 {
		log.Printf("schedule working,current block num is %v", block.Number())
	}
	for i := 0; i < len(c.jobs); {
		if c.jobs[i].done {
			log.Printf("[job] %v done", c.jobs[i].desc)
			c.jobs = append(c.jobs[:i], c.jobs[i+1:]...)
		} else {
			if !c.jobs[i].runing {
				go func(j *job) {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("Recovered in :%v", r)
							c.errch <- fmt.Errorf("job %v painc", j.desc)
							j.done = true
						}
					}()
					j.runing = true
					ok, err := j.run(block)
					if err != nil {
						c.errch <- fmt.Errorf("job %v fail:%v", c.jobs[i].desc, err)
						j.done = true
						return
					}
					if ok {
						j.done = true
					}
					j.runing = false
				}(c.jobs[i])
			}
			i++
		}
	}
	if len(c.jobs) == 0 {
		c.donech <- struct{}{}
		close(c.donech)
	}
}
