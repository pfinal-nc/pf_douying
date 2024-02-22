package main

import (
	"changeme/lib"
	"context"
	"fmt"
	"golang.org/x/exp/rand"
	"time"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) JoinRoom(roomId string) int {
	rand.Seed(uint64(time.Now().UnixNano()))

	liveID := roomId
	fetcher := lib.NewDouyinLiveWebFetcher(liveID)
	go fetcher.Start()
	// defer fetcher.Stop()
	// Keep main alive
	return 1
}

func (a *App) LiveRoom() int {
	// 通道中写入状态
	lib.StateChan <- "colse"
	return 0
}

func (a *App) GetRoomMsg() string {
	// MessageChan 从通道中读取消息
	select {
	case message := <-lib.MessageChan:
		fmt.Println(message)
		return message
	default:
		return ""
	}
}
