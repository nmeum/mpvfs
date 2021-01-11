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

	volume float64
	status playback
}

func newPlayerState(mpv *mpv.Client) (*playerState, error) {
	state := &playerState{mpv: mpv, mtx: new(sync.Mutex)}

	pause, err := mpv.ObserveProperty("pause")
	if err != nil {
		return nil, err
	}
	go state.updateState(pause)

	volume, err := mpv.ObserveProperty("volume")
	if err != nil {
		return nil, err
	}
	go state.updateVolume(volume)

	return state, nil
}

func (p *playerState) updateState(ch <-chan interface{}) {
	for data := range ch {
		isPaused := data.(bool)
		p.mtx.Lock()
		if isPaused {
			p.status = paused
		} else {
			p.status = playing
		}
		p.mtx.Unlock()
	}

}

func (p *playerState) updateVolume(ch <-chan interface{}) {
	for data := range ch {
		p.mtx.Lock()
		p.volume = data.(float64)
		p.mtx.Unlock()
	}
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
