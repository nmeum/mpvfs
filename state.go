package main

import (
	"github.com/nmeum/mpvfs/mpv"
	"sync"
)

type playback int

const (
	paused = iota
	playing
)

type playerState struct {
	mpv *mpv.Client
	mtx *sync.Mutex

	status playback
}

func newPlayerState(mpv *mpv.Client) (*playerState, error) {
	state := &playerState{mpv: mpv, mtx: new(sync.Mutex)}

	pause, err := mpv.ObserveProperty("pause")
	if err != nil {
		return nil, err
	}

	go func(ch <-chan interface{}) {
		isPaused := (<-ch).(bool)

		state.mtx.Lock()
		if isPaused {
			state.status = paused
		} else {
			state.status = playing
		}
		state.mtx.Unlock()
	}(pause)

	return state, nil
}

func (p *playerState) IsPaused() bool {
	p.mtx.Lock()
	r := p.status == paused
	p.mtx.Unlock()

	return r
}

func (p *playerState) IsPlaying() bool {
	p.mtx.Lock()
	r := p.status == playing
	p.mtx.Unlock()

	return r
}

func (p *playerState) Index() uint {
	return 0
}
