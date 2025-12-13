package plw

import (
	"dev_tool/internal/app/dtool/struct"
	"github.com/playwright-community/playwright-go"
	"sync"
	"time"
)

var pageActives map[string]_struct.PageActiveTime
var pageActivesLock sync.RWMutex

type PageActiveTime struct {
}

func InitPageActiveTime() {
	pageActivesLock.Lock()
	pageActives = make(map[string]_struct.PageActiveTime)
	pageActivesLock.Unlock()
	go func() {
		for {
			time.Sleep(time.Second)
			pageActivesLock.Lock()
			newMap := make(map[string]_struct.PageActiveTime)
			for pageUrl, activeTime := range pageActives {
				if activeTime.ActiveTime.Add(time.Second * time.Duration(activeTime.AutoCloseSecond)).Before(time.Now()) {
					go func() {
						_ = (*activeTime.Page).Close()
					}()
				} else {
					newMap[pageUrl] = activeTime
				}
			}
			pageActives = newMap
			pageActivesLock.Unlock()
		}
	}()

}

func NewPageActiveTime() *PageActiveTime {
	return &PageActiveTime{}
}

func (h *PageActiveTime) Add(page *playwright.Page, autoCloseSecond int) {
	go func() {
		pageActivesLock.Lock()
		defer pageActivesLock.Unlock()
		pageActives[(*page).URL()] = _struct.PageActiveTime{
			ActiveTime:      time.Now(),
			AutoCloseSecond: autoCloseSecond,
			Page:            page,
		}
	}()

}
