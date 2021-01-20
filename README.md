# mpvfs

[9P][9P wikipedia] file server for controlling [mpv][mpv web] playback.

## Status

Toy project for experimenting with [neinp][neinp source]. Currently
provides a very buggy, incomplete, and unfinished implementation of
[playlistfs][9front playlistfs].

## Installation

Install using `go get` as follows:

	$ go get github.com/nmeum/mpvfs

## Usage

This software relies on mpv's IPC mechanism. Currently, mpv must be
started separately. For instance, as:

	$ mpv --keep-open=yes --idle --pause --input-ipc-server=/tmp/mpv-socket song.flac

Afterwards, mpvfs itself must be started as:

	$ ./mpvfs -a localhost:9999 /tmp/mpv-socket

The created 9P server can than be mounted using any 9P implementation.
Examples are given below. On Plan 9 derivatives, it is possible to
partially control playback using the [games/jukebox][9front jukebox]
client for playlistfs.

### Linux

Linux provides [v9fs][v9fs documentation], an in-tree implementation of
the 9P protocol. If this implementation is enabled in your utilized
Linux Kernel version, you can mount the file server as follows:

	# mount -t9p -o port=9999,noextend 127.0.0.1 /media/9p/

Afterwards, you can interact with the `playctl`, `playvol`, and
`playlist` files provided at the given mount point. For instance,
`echo play > /media/9p/playctl` will start playback. Refer to the
[playlistfs manual][9front playlistfs] for more information on the
provided files.

### Plan 9

If you are running Plan 9 in QEMU you need to create a corresponding
`guestfwd` rule, e.g. `guestfwd=tcp:10.0.2.4:1234-cmd:"nc 127.0.0.1 9999"`.
If you want to use jukebox, you might also need to create
`/sys/lib/music/map`, refer to the [juke man page][9front juke] for more
information.

Mount mpvfs on `/n/mpvfs` using `srv(4)`:

	% srv -m net!10.0.2.4!1234 mpvfs /n/mpvfs

If jukebox is available, add the resources to the `/mnt` union directory:

	% bind -b /n/mpvfs/ /mnt/

Finally, start jukefs and jukebox as follows:

	% games/jukefs
	% games/jukebox

Controlling playback through the buttons in the top left corner should
work, changing volume should also work fine. Everything else is a bit
wonky at the moment.

## License

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later
version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
this program. If not, see <https://www.gnu.org/licenses/>.

[9P wikipedia]: https://en.wikipedia.org/wiki/9P_(protocol)
[mpv web]: https://mpv.io/
[9front playlistfs]: http://man.9front.org/7/playlistfs
[neinp source]: https://git.sr.ht/~rbn/neinp
[9front jukebox]: http://man.9front.org/7/juke
[9front juke]: http://man.9front.org/7/juke
[v9fs documentation]: https://www.kernel.org/doc/html/latest/filesystems/9p.html
