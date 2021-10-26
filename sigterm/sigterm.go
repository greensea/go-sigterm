package sigterm

import (
    "sync"
    "log"
    "os"
    "syscall"
    "os/signal"
)

/// 用于控制退出的 waitgroup，当一个不能中断的任务开始的时候，应该 WG.Add(1)，当结束的时候，应该 WG.Done()
/// 收到 TERM 信号之后，整个进程会调用 WG.Wait()，最后才退出
var WG sync.WaitGroup

var isExit = false

func init() {
    go handleSignal()
}

/// app.WG.Add() 的快捷方式
func Add(n int) {
    WG.Add(n)
}

/// app.WG.Done() 的快捷方式
func Done() {
    WG.Done()
}

func handleSignal() {    
    var ExitOnce sync.Once
    
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    force_quit := 10
    
    for s := range c {
        force_quit--
        
        isExit = true
        
        if force_quit <= 0 {
            log.Printf("已经收到了多次停止信号，直接结束进程")
            os.Exit(0)
        } else {
            log.Printf("收到信号 %v，正在要求所有任务结束，若要直接强行退出，请再发送 %d 次信号", s, force_quit)
            
            go func() {
                ExitOnce.Do(func() {
                    WG.Wait()
                    os.Exit(0)
                })
            }()
        }
    }
}


/// 进程是否应该退出
/// 其他任务应该间断性地检查这个标志，当这个标志设为 true 的时候，应当结束任务，并视情况进行 WG.Done()
/// 当 sigterm 发现所有 WG 都已经 Done 后，自动执行 os.Exit(0)
func Is() bool {
    return isExit
}
