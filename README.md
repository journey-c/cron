# cron

线程安全的golang 定时任务库，兼容crontab命令格式

# 使用方法
## 代码示例

```
func test1() {
	fmt.Println(time.Now().UnixNano(), "test1")
}

func test2() {
	fmt.Println(time.Now().UnixNano(), "test2")
}

func TestCron(t *testing.T) {
	cron := NewCron()

	if err := cron.JobAdd("1 * * * * *", test1); err != nil {
		t.Log(err.Error())
		return
	}

	if err := cron.JobAdd("2 * * * * *", test2); err != nil {
		t.Log(err.Error())
		return
	}

	cron.Start()

	time.Sleep(300 * time.Second)

	cron.Stop()

	time.Sleep(5 * time.Second)
}
```

## 命令格式
- 普通命令
crontab支持五位
```
* * * * * cmd
│ │ │ │ │
│ │ │ │ └─周0~7
│ │ │ └───月1~12
│ │ └─────日1~31
│ └───────时0~23
└─────────分0~59
```
本系统支持六位
```
* * * * * * cmd
│ │ │ │ │ │
│ │ │ │ │ └─周0~7
│ │ │ │ └───月1~12
│ │ │ └─────日1~31
│ │ └───────时0~23
│ └─────────分0~59
└───────────秒0~59
```
如果命令为5位时，默认认为省略了秒位
```
* * * * * 相当于 0 * * * * *
```

- 当然也支持特殊命令
```
string         meaning
------         -------
@reboot        Not support
@yearly        Run once a year, "0 0 0 1 1 *".
@annually      (same as @yearly)
@monthly       Run once a month, "0 0 0 1 * *".
@weekly        Run once a week, "0 0 0 * * 0".
@daily         Run once a day, "0 0 0 * * *".
@midnight      (same as @daily)
@hourly        Run once an hour, "0 0 * * * *".
```
