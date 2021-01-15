package main

import (
	"github.com/nmeum/mpvfs/mpv"
	"sync"
	"sync/atomic"
)

type playback int

const (
	paused = iota
	playing
)

type playerState struct {
	mpv *mpv.Client
	mtx *sync.Mutex

	volume uint32
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
		vol := data.(float64)
		atomic.StoreUint32(&p.volume, uint32(vol))
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

func (p *playerState) Volume() uint {
	vol := atomic.LoadUint32(&p.volume)
	return uint(vol)
}

func (p *playerState) Index() uint {
	return 0
}
