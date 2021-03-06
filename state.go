package main

import (
	"github.com/nmeum/mpvfs/mpv"

	"fmt"
	"sync"
)

type playerState struct {
	mpv *mpv.Client

	volume  uint
	volCond *sync.Cond

	playlist []string
	playCond *sync.Cond

	stateCond *sync.Cond
	playing   bool

	posCond *sync.Cond
	pos     int

	errChan chan error
}

func newPlayerState(mpv *mpv.Client) (*playerState, error) {
	state := &playerState{
		pos:       -1,
		mpv:       mpv,
		volCond:   sync.NewCond(new(sync.Mutex)),
		posCond:   sync.NewCond(new(sync.Mutex)),
		playCond:  sync.NewCond(new(sync.Mutex)),
		stateCond: sync.NewCond(new(sync.Mutex)),
		errChan:   make(chan error, 1),
	}

	observers := map[string]func(ch <-chan interface{}){
		"pause":          state.updateState,
		"volume":         state.updateVolume,
		"playlist-pos":   state.updatePosition,
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
		p.stateCond.L.Lock()
		p.playing = !(data.(bool))
		p.stateCond.Broadcast()
		p.stateCond.L.Unlock()
	}
}

func (p *playerState) updateVolume(ch <-chan interface{}) {
	for data := range ch {
		vol := data.(float64)

		p.volCond.L.Lock()
		p.volume = uint(vol)
		p.volCond.Broadcast()
		p.volCond.L.Unlock()
	}
}

func (p *playerState) updatePosition(ch <-chan interface{}) {
	for data := range ch {
		p.posCond.L.Lock()
		pos := data.(float64)
		p.pos = int(pos)

		p.posCond.Broadcast()
		p.posCond.L.Unlock()
	}
}

// XXX: This implementation assumes that the playlist is never cleared.
func (p *playerState) updatePlaylist(ch <-chan interface{}) {
	for data := range ch {
		newCount := int(data.(float64))
		if newCount < 0 {
			panic("unreachable")
		}

		p.playCond.L.Lock()
		oldCount := len(p.playlist)
		diff := newCount - oldCount

		for i := 0; i < diff; i++ {
			entry, err := p.song(oldCount + i)
			if err != nil {
				p.errChan <- err
				continue
			}
			p.playlist = append(p.playlist, entry)
		}

		p.playCond.Broadcast()
		p.playCond.L.Unlock()
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

	// TODO: What if the title contains single qoutes?
	return fmt.Sprintf("%s %s", name, title), nil
}

func (p *playerState) Playing() bool {
	p.stateCond.L.Lock()
	r := p.playing
	p.stateCond.L.Unlock()

	return r
}

func (p *playerState) WaitPlaying() bool {
	p.stateCond.L.Lock()
	oldState := p.playing
	for oldState == p.playing {
		p.stateCond.Wait()
	}

	newState := p.playing
	p.stateCond.L.Unlock()

	return newState
}

// Index returns the current position on the playlist, or
// -1 if there is no current entry (e.g. playlist is empty).
func (p *playerState) Index() int {
	p.posCond.L.Lock()
	r := p.pos
	p.posCond.L.Unlock()

	return r
}

func (p *playerState) WaitIndex() int {
	p.posCond.L.Lock()
	oldPos := p.pos
	for oldPos == p.pos {
		p.posCond.Wait()
	}

	newPos := p.pos
	p.posCond.L.Unlock()

	return newPos
}

func (p *playerState) Volume() uint {
	p.volCond.L.Lock()
	vol := p.volume
	p.volCond.L.Unlock()

	return vol
}

func (p *playerState) WaitVolume() uint {
	p.volCond.L.Lock()
	oldvol := p.volume
	// TODO: What happens when `echo ${oldvol} >> playvol`?
	for oldvol == p.volume {
		p.volCond.Wait()
	}

	vol := p.volume
	p.volCond.L.Unlock()

	return vol
}

func (p *playerState) Playlist() []string {
	p.playCond.L.Lock()
	r := p.playlist
	p.playCond.L.Unlock()

	return r
}

// WaitPlayist blocks until the playlist changes and returns the
// most recent entry added to the playlist.
func (p *playerState) WaitPlayist() string {
	oldlen := len(p.playlist)
	p.playCond.L.Lock()
	for len(p.playlist) <= oldlen {
		p.playCond.Wait()
	}

	newIndex := len(p.playlist) - 1
	r := p.playlist[newIndex]
	p.playCond.L.Unlock()

	return r
}
