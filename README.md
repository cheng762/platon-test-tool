# PlatON-Integration-Tests
An automated PlatON's Integration testing tool.


how to test
```
./platon-tool-test prepare   会根据配置的账户随机生成10个账户并分别转账0x200000000000000000000000000000000000000000000000000000000
./platon-tool-test start 启动所有测试用例
```

command
```
COMMANDS:
   exec     exec a test cases
   list     list all cases
   prepare  prepare some accounts are used for  test 
   start    start all test cases
   help, h  Shows a list of commands or help for one command

```

配置文件
```
{
  "url":"http://127.0.0.1:6980",//所请求的节点的端口号
  "account":"0xf66CB3C7f28D058AE3C6eD9493C6A9e2a7d7786d", //默认账户
  "prikey": "bfa6c75e2240a4735fdc99a73b48ae42d625f34b859327fc2f0e553f7e97888e",账户私钥
  "dir": "config/",//配置目录
  "restricting_config_file":"case_restricting.json",//测试用例的配置文件
  "private_key_file": "privateKeys.txt",//生成的私钥
  "default_account_addr_file": "addr.json" //生成的账户
}
```


开发说明  
编写用例实现caseTest
```
func init() {
	allCases = make(map[string]caseTest)
	allCases["restricting"] = new(restrictCases)
}

type caseTest interface {
    //执行你的所有测试用例
	Start() error
    //执行你指定的单个测试用例
	Exec(string)error
    //准备你的测试用例
	Prepare() error
    //结束你的测试用例后需要做的
	End() error
    //列出你的所有测试用例
	List()[]string
}

type restrictCases struct {
	base commonCases
}

func (r *restrictCases) Prepare() error {
    ...
	return nil
}

func (r *restrictCases) Start() error {
    ...
	return nil
}

func (r *restrictCases)Exec(caseName string)error {
	...
	return nil
}

func (r *restrictCases) End() error {
	...
	return nil
}

func (r *restrictCases)List()[]string  {
	...
	return names
}
```