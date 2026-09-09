// play-sound — cross-platform notification sound for Claude Code Stop / StopFailure / Notification hooks.
//
// Plays the bundled notify.wav on Windows (.NET SoundPlayer), macOS (afplay), or
// Linux (first available of pw-play / paplay / aplay / play / ffplay / canberra-gtk-play).
// The launcher passes the wav path as the first argument.
//
// Best-effort: a missing player or absent audio device must never break the
// session, so every failure is swallowed and the process exits 0.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// play runs a player command, discarding its output; true if it exited 0.
func play(name string, args ...string) bool {
	return exec.Command(name, args...).Run() == nil
}

func main() {
	defer func() {
		_ = recover()
		os.Exit(0)
	}()

	if len(os.Args) < 2 {
		return
	}
	sound := os.Args[1]
	if _, err := os.Stat(sound); err != nil {
		return
	}

	switch runtime.GOOS {
	case "windows":
		// Built-in .NET sound player, no extra tools required.
		play("powershell", "-NoProfile", "-Command",
			fmt.Sprintf("(New-Object Media.SoundPlayer '%s').PlaySync()", sound))
	case "darwin":
		play("afplay", sound)
	default:
		// Linux / other: try common players in order, stop at the first that works.
		candidates := [][]string{
			{"pw-play", sound},     // PipeWire
			{"paplay", sound},      // PulseAudio / PipeWire
			{"aplay", "-q", sound}, // ALSA
			{"play", "-q", sound},  // SoX
			{"ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", sound}, // FFmpeg
			{"canberra-gtk-play", "-f", sound},                              // libcanberra
		}
		for _, c := range candidates {
			if play(c[0], c[1:]...) {
				break
			}
		}
	}
}
