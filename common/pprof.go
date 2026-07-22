package common

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

// Monitor 定时监控cpu使用率，超过阈值输出pprof文件。
// P1 修复：
//  1. 移除 panic(err)，避免后台任务异常导致进程退出，改为记录错误后继续运行。
//  2. 接受 ctx，服务退出时可通过 ctx.Done() 停止循环。
func Monitor(ctx context.Context) {
	for {
		percent, err := cpu.Percent(time.Second, false)
		if err != nil {
			// P1 修复：后台任务禁止 panic 导致进程退出
			SysError("Monitor cpu.Percent error: " + err.Error())
			if !SleepWithContext(ctx, 30*time.Second) {
				SysLog("Monitor stopping: context cancelled")
				return
			}
			continue
		}
		if len(percent) > 0 && percent[0] > 80 {
			SysLog("cpu usage too high")
			// write pprof file
			if _, err := os.Stat("./pprof"); os.IsNotExist(err) {
				err := os.Mkdir("./pprof", os.ModePerm)
				if err != nil {
					SysLog("创建pprof文件夹失败 " + err.Error())
					if !SleepWithContext(ctx, 30*time.Second) {
						SysLog("Monitor stopping: context cancelled")
						return
					}
					continue
				}
			}
			f, err := os.Create("./pprof/" + fmt.Sprintf("cpu-%s.pprof", time.Now().Format("20060102150405")))
			if err != nil {
				SysLog("创建pprof文件失败 " + err.Error())
				if !SleepWithContext(ctx, 30*time.Second) {
					SysLog("Monitor stopping: context cancelled")
					return
				}
				continue
			}
			err = pprof.StartCPUProfile(f)
			if err != nil {
				SysLog("启动pprof失败 " + err.Error())
				f.Close()
				if !SleepWithContext(ctx, 30*time.Second) {
					SysLog("Monitor stopping: context cancelled")
					return
				}
				continue
			}
			// 采样 10 秒，期间可被 ctx 取消
			if !SleepWithContext(ctx, 10*time.Second) {
				pprof.StopCPUProfile()
				f.Close()
				SysLog("Monitor stopping: context cancelled")
				return
			}
			pprof.StopCPUProfile()
			f.Close()
		}
		if !SleepWithContext(ctx, 30*time.Second) {
			SysLog("Monitor stopping: context cancelled")
			return
		}
	}
}
