package main

import (
	"github.com/nmeum/mpvfs/mpv"

	"fmt"
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

	volume   uint32
	status   playback
	playlist []string
}

func newPlayerState(mpv *mpv.Client) (*playerState, error) {
	state := &playerState{mpv: mpv, mtx: new(sync.Mutex)}

	observers := map[string]func(ch <-chan interface{}){
		"pause":          state.updateState,
		"volume":         state.updateVolume,
		"playlist-count": state.updatePlaylist,
	}

	for property, observer := range observers {
		ch, err := mpv.ObserveProperty(property)
		if err != nil {
			return nil, err
		}

		go observer(ch)
	}

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

func (p *playerState) updatePlaylist(ch <-chan interface{}) {
	for data := range ch {
		newCount := int(data.(float64))

		nameProp := fmt.Sprintf("playlist/%d/filename", newCount-1)
		name, err := p.mpv.GetProperty(nameProp)
		if err != nil {
			panic(err)
		}

		p.mtx.Lock()
		p.playlist = append(p.playlist, name.(string))
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

func (p *playerState) Volume() uint {
	vol := atomic.LoadUint32(&p.volume)
	return uint(vol)
}

func (p *playerState) Playlist() []string {
	p.mtx.Lock()
	r := p.playlist
	p.mtx.Unlock()

	return r
}

func (p *playerState) Index() uint {
	return 0
}
