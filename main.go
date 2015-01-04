//
// Connects to wingo notify and writes out the notices as they arrive
//
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	//"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/cmd"
)

const (
	battUnknownStat = "unknown"
	datefmt         = "2006-01-02 03:04PM"
)

var (
	hilight = flag.String("cur-fg", "", "When set, will be the foreground color for the current workspace.")
	lolight = flag.String("other-fg", "", "When set, will be the foreground color for non-current workspaces.")
)

func main() {
	flag.Parse()

	// Connect to the Wingo command server.
	cmdSockPath := socketFilePath()
	cmdConn, err := net.Dial("unix", cmdSockPath)
	if err != nil {
		panic(fmt.Sprintf("Could not connect to Wingo IPC: '%s'\n", err))
	}
	defer cmdConn.Close()

	notifChan := make(chan map[string]interface{})
	defer close(notifChan)
	go notifierListener(notifChan)

	tickerChan := time.NewTicker(time.Minute)
	defer tickerChan.Stop()

	batt := getBatt()
	showLine(cmdConn, time.Now(), batt)
	for {
		select {
		case time := <-tickerChan.C:
			batt = getBatt()
			showLine(cmdConn, time, batt)
			break
		case notif := <-notifChan:
			evt := notif["EventName"]
			//fmt.Fprintln(os.Stderr, "Event ", evt, " = ", notif)
			switch evt {
			case "ChangedVisibleWorkspace", "ManagedClient", "UnmanagedClient":
				showLine(cmdConn, time.Now(), batt)
			}
		}
	}
}

func getBatt() string {
	c := cmd.New("apm", "-m")
	if err := c.Run(); err != nil {
		return battUnknownStat
	}
	return strings.TrimSpace(c.BufStdout.String())
}

func showLine(conn net.Conn, now time.Time, batt string) {
	// Build and format the gobar line
	ws_lst := strings.Split(sendCmd(conn, "GetWorkspaceList"), "\n")
	curws := strings.TrimSpace(sendCmd(conn, "GetWorkspace"))
	disp := buildDisplayLine(conn, curws, ws_lst, now.Format(datefmt), batt)
	fmt.Fprint(os.Stdout, disp)
}

func buildDisplayLine(conn net.Conn, curws string, ws_lst []string, stime string, sbatt string) string {
	ws_str := make([]byte, 0)

	// workspaces
	for _, ws := range ws_lst {
		var piece []byte
		clients := sendCmd(conn, fmt.Sprintf("GetClientList \"%s\"", ws))
		clients = strings.TrimSpace(clients)
		clmark := ""
		if len(clients) > 0 {
			clmark = "*"
		}
		if curws == ws {
			hilight_ctrl := ""
			if len(*hilight) > 0 {
				hilight_ctrl = "CF0x" + *hilight
			}
			piece = []byte(fmt.Sprintf("{%s[%s]}  ", hilight_ctrl, ws))
		} else {
			lolight_ctrl := ""
			if len(*lolight) > 0 {
				lolight_ctrl = "CF0x" + *lolight
			}
			piece = []byte(fmt.Sprintf("%s{%s%s}  ", clmark, lolight_ctrl, ws))
		}
		ws_str = append(ws_str, []byte(piece)...)
	}

	// time
	if sbatt == battUnknownStat {
		ws_str = append(ws_str, []byte(fmt.Sprintf("{ARpwr | %s}", stime))...)
	} else {
		ws_str = append(ws_str, []byte(fmt.Sprintf("{ARbatt %sm | %s}", sbatt, stime))...)
	}

	ws_str = append(ws_str, []byte("\n")...)
	return string(ws_str)
}

func notifierListener(notifChan chan map[string]interface{}) {
	// Connect to the Wingo notification server.
	notifSockPath := notifySocketFilePath()
	notifConn, err := net.Dial("unix", notifSockPath)
	if err != nil {
		panic(fmt.Sprintf("Could not connect to Wingo notifications: '%s'\n", err))
	}
	defer notifConn.Close()

	rdr := bufio.NewReader(notifConn)
	for {
		notice, err := rdr.ReadString(0)
		if err != nil {
			break
		}
		notice = notice[:len(notice)-1] // get rid of trailing null
		data := []byte(notice)

		jsonMap := make(map[string]interface{})
		if err = json.Unmarshal(data, &jsonMap); err != nil {
			fmt.Fprint(os.Stderr, "Error marshalling JSON: ", err)
			continue
		}
		evt, ok := jsonMap["EventName"]
		if !ok {
			continue
		}
		if evt == "Noop" {
			continue
		}
		notifChan <- jsonMap
	}
}

func sendCmd(conn net.Conn, cmds string) string {
	// Build the wingo command buffer
	cmd := []byte(fmt.Sprintf("%s%c", cmds, 0))

	// Send it.
	if _, err := conn.(io.Writer).Write(cmd); err != nil {
		panic(fmt.Sprintf("Error writing command: %s", err))
	}

	// Read the response.
	reader := bufio.NewReader(conn.(io.Reader))
	msg, err := reader.ReadString(0)
	if err != nil {
		panic(fmt.Sprintf("Could not read response: %s", err))
	}
	msg = msg[:len(msg)-1] // get rid of null terminator

	return strings.TrimSpace(msg)
}

func socketFilePath() string {
	// Try to read it from env
	sockpath := os.Getenv("WINGO_SOCKET")
	if len(sockpath) > 0 {
		return strings.TrimSpace(sockpath)
	}

	// Ask wingo where it is
	bGotIt := false
	var backoff int64 = 1
	for !bGotIt {
		c := cmd.New("wingo", "--show-socket")
		if err := c.Run(); err != nil {
			fmt.Fprint(os.Stderr, err)
			time.Sleep(time.Duration(int64(time.Millisecond) * backoff))
			backoff += 100
			if backoff > 2000 {
				break
			}
		} else {
			return strings.TrimSpace(c.BufStdout.String())
		}
	}

	// Try to build it from XDG
	xdg_run := os.Getenv("XDG_RUNTIME_DIR")
	disp := os.Getenv("DISPLAY")
	if len(xdg_run) > 0 && len(disp) > 0 {
		return fmt.Sprintf("%s/wingo/%s.0", xdg_run, disp)
	}

	return ""
}

func notifySocketFilePath() string {
	// Try to read it from env
	sockpath := os.Getenv("WINGO_NOTIFY_SOCKET")
	if len(sockpath) > 0 {
		return strings.TrimSpace(sockpath)
	}

	// Ask wingo where it is
	bGotIt := false
	var backoff int64 = 1
	for !bGotIt {
		c := cmd.New("wingo", "--show-notify-socket")
		if err := c.Run(); err != nil {
			fmt.Fprint(os.Stderr, err)
			time.Sleep(time.Duration(int64(time.Millisecond) * backoff))
			backoff += 100
			if backoff > 2000 {
				break
			}
		} else {
			return strings.TrimSpace(c.BufStdout.String())
		}
	}

	// Try to build it from XDG
	xdg_run := os.Getenv("XDG_RUNTIME_DIR")
	disp := os.Getenv("DISPLAY")
	if len(xdg_run) > 0 && len(disp) > 0 {
		return fmt.Sprintf("%s/wingo/%s.0", xdg_run, disp)
	}

	return ""
}
