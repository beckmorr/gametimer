package notify

import "os/exec"

func Send(title, body string) {
	if _, err := exec.LookPath("notify-send"); err != nil {
		return
	}
	_ = exec.Command("notify-send", title, body).Run()
}

func Sound() {
	if _, err := exec.LookPath("paplay"); err != nil {
		return
	}
	_ = exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga").Run()
}
