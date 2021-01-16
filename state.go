package main

import (
	"github.com/nmeum/mpvfs/mpv"

	"fmt"
	"sync"
	"sync/atomic"
)

type playerState struct {
	mpv *mpv.Client
	mtx *sync.Mutex

	volume   uint32
	playing  bool
	playlist []string

	errChan chan error
}

func newPlayerState(mpv *mpv.Client) (*playerState, error) {
	state := &playerState{
		mpv:     mpv,
		mtx:     new(sync.Mutex),
		errChan: make(chan error, 1),
	}

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

func (p *playerState) ErrChan() <-chan error {
	return p.errChan
}

func (p *playerState) updateState(ch <-chan interface{}) {
	for data := range ch {
		p.mtx.Lock()
		p.playing = !(data.(bool))
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
		if newCount < 0 {
			panic("unreachable")
		}

		p.mtx.Lock()
		entry, err := p.song(newCount - 1)
		if err != nil {
			p.mtx.Unlock()
			p.errChan <- err
			continue
		}
		p.playlist = append(p.playlist, entry)
		p.mtx.Unlock()
	}
}

func (p *playerState) song(idx int) (string, error) {
	nameProp := fmt.Sprintf("playlist/%d/filename", idx)
	name, err := p.mpv.GetProperty(nameProp)
	if err != nil {
		return "", err
	}

	titleProp := fmt.Sprintf("playlist/%d/title", idx)
	title, err := p.mpv.GetProperty(titleProp)
	if err != nil {
		// Property may not be available
		return name.(string), nil
	}

	return fmt.Sprintf("%s \"%s\"", name, title), nil
}

func (p *playerState) IsPlaying() bool {
	p.mtx.Lock()
	r := p.playing
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
