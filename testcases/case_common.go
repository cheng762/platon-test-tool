package testcases

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	"github.com/PlatONnetwork/PlatON-Go/crypto"
	"github.com/PlatONnetwork/PlatON-Go/ethclient"

	common2 "github.com/PlatONnetwork/platon-test-tool/common"

	"github.com/robfig/cron"
	"log"
	"reflect"
	"strings"
)

const PrefixCase = "Case"

type commonCases struct {
	TxManager common2.TxManager
	corn      *cron.Cron
	jobs      []*job
	errors    []error
	errch     chan error
	donech    chan struct{}
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

func (c *commonCases) list(m caseTest) []string {
	var names []string
	object := reflect.TypeOf(m)
	for i := 0; i < object.NumMethod(); i++ {
		method := object.Method(i)
		if strings.HasPrefix(method.Name, PrefixCase) {
			names = append(names, strings.TrimPrefix(method.Name, "Case"))
		}
	}
	return names
}

func (c *commonCases) exec(caseName string, m caseTest) error {
	methodNames := PrefixCase + caseName
	val := reflect.ValueOf(m).MethodByName(methodNames).Call([]reflect.Value{})
	if val[0].IsNil() {
		return nil
	}
	err := val[0].Interface().(error)
	return err
}

//初始化case
func (c *commonCases) Prepare() error {
	client, err := ethclient.Dial(config.Url)
	if err != nil {
		return err
	}
	c.client = client
	//	c.adrs = GetAllAddress()
	c.corn = cron.New()
	if _, err := c.corn.AddFunc("@every 1s", c.schedule); err != nil {
		return err
	}
	c.donech = make(chan struct{}, 0)
	c.errch = make(chan error, 0)
	c.errors = make([]error, 0)
	return nil
}

func (c *commonCases) Start() error {
	return nil
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

func (c *commonCases) SendError(caseName string, err error) error {
	log.Printf("[fail]test case %v fail: %v ", caseName, err)
	return fmt.Errorf("test case %v fail: %v ", caseName, err)
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

//生成一个新的空账户，私钥和地址
func (c *commonCases) generateEmptyAccount() (*ecdsa.PrivateKey, common.Address) {
	privateKey, _ := crypto.GenerateKey()
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	var pri PriAccount
	pri.Priv = privateKey
	pri.Address = address
	allAccounts[address] = &pri
	return privateKey, address
}

func (c *commonCases) schedule() {
	block, err := c.client.BlockByNumber(context.Background(), nil)
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
