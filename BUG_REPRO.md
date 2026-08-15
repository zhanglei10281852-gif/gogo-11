# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

闰日 cron 在跨年推演时丢失，请帮我修复。

表达式 0 0 29 2 * 在 2024-02-29 之后查询下一次执行时返回不存在；把区间扩到 2032 年也只能枚举到第一个闰日。每五分钟等高频表达式目前正常。

期望所有合法 cron 都能找到公历周期内真实存在的后续触发，多年 AllBetween 能完整有序枚举 2024、2028、2032 的闰日，同时不能把普通表达式变慢或改变 DOM/DOW 语义。修复后请保证 go test ./... 全绿。

## 含 Bug 版本

- 仓库：zhanglei10281852-gif/gogo-11
- 仓库地址：https://github.com/zhanglei10281852-gif/gogo-11.git
- parent SHA：18d7a57bf1b1d8a24331e6a1179965d15ff3ccfa

## 复现步骤

```bash
git clone -- https://github.com/zhanglei10281852-gif/gogo-11.git bug-repro
cd bug-repro
git checkout --detach 18d7a57bf1b1d8a24331e6a1179965d15ff3ccfa
go test ./pkg/cron -run "^(TestNextLeapDayAcrossFourYears|TestAllBetweenEnumeratesLeapDaysAcrossYears|TestAllBetweenFrequentSchedule)$" -count=1 -v
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./pkg/cron -run "^(TestNextLeapDayAcrossFourYears|TestAllBetweenEnumeratesLeapDaysAcrossYears|TestAllBetweenFrequentSchedule)$" -count=1 -v
=== RUN   TestNextLeapDayAcrossFourYears
    schedule_sparse_test.go:26: Next reported no future execution
--- FAIL: TestNextLeapDayAcrossFourYears (0.02s)
=== RUN   TestAllBetweenEnumeratesLeapDaysAcrossYears
    schedule_sparse_test.go:45: AllBetween returned 1 executions, want 3: [2024-02-29 00:00:00 +0000 UTC]
--- FAIL: TestAllBetweenEnumeratesLeapDaysAcrossYears (0.02s)
=== RUN   TestAllBetweenFrequentSchedule
--- PASS: TestAllBetweenFrequentSchedule (0.00s)
FAIL
FAIL	github.com/ops/cronchecker/pkg/cron	0.035s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./pkg/cron -run "^(TestNextLeapDayAcrossFourYears|TestAllBetweenEnumeratesLeapDaysAcrossYears|TestAllBetweenFrequentSchedule)$" -count=1 -v
=== RUN   TestNextLeapDayAcrossFourYears
    schedule_sparse_test.go:26: Next reported no future execution
--- FAIL: TestNextLeapDayAcrossFourYears (0.12s)
=== RUN   TestAllBetweenEnumeratesLeapDaysAcrossYears
    schedule_sparse_test.go:45: AllBetween returned 1 executions, want 3: [2024-02-29 00:00:00 +0000 UTC]
--- FAIL: TestAllBetweenEnumeratesLeapDaysAcrossYears (0.12s)
=== RUN   TestAllBetweenFrequentSchedule
--- PASS: TestAllBetweenFrequentSchedule (0.00s)
FAIL
FAIL	github.com/ops/cronchecker/pkg/cron	0.370s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

定向测试、go test ./... -count=1、go build ./...、go vet ./... 全部通过；linux/amd64 和 linux/arm64 均通过；不得只扩大固定分钟循环上限。
