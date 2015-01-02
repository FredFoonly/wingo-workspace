//
// Connects to wingo ipc using the WINGO_SOCKET or XDG_RUNTIME_DIR
// variables and writes out a gobar-formatted line containing the
// current workspace, list of workspaces, and whether each workspace
// has clients.
//

package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/BurntSushi/cmd"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

var wsfile = flag.String("f", "", "When set, will be the file where the gobar-formatted workspace line is written.")
var hilight = flag.String("cur-fg", "", "When set, will be the foreground color for the current workspace.")
var lolight = flag.String("other-fg", "", "When set, will be the foreground color for non-current workspaces.")

func main() {
	flag.Parse()
	// Connect to the Wingo command server.
	sockpath := socketFilePath()
	conn, err := net.Dial("unix", sockpath)
	if err != nil {
		log.Fatalf("Could not connect to Wingo IPC: '%s'\n", err)
	}
	defer conn.Close()

	// Build and format the gobar line
	ws_lst := strings.Split(sendCmd(conn, "GetWorkspaceList"), "\n")
	curws := strings.TrimSpace(sendCmd(conn, "GetWorkspace"))
	disp := buildDisplay(conn, curws, ws_lst)

	// And write it out to either stdout or the specified file
	if len(*wsfile) > 0 {
		ioutil.WriteFile(*wsfile, []byte(disp), 0666)
	} else {
		print(disp)
	}
}

func buildDisplay(conn net.Conn, curws string, ws_lst []string) string {
	ws_str := make([]byte, 0)
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
	ws_str = append(ws_str, []byte("\n")...)
	return string(ws_str)
}

func sendCmd(conn net.Conn, cmds string) string {
	// Build the wingo command buffer
	cmd := []byte(fmt.Sprintf("%s%c", cmds, 0))

	// Send it.
	if _, err := conn.(io.Writer).Write(cmd); err != nil {
		log.Fatalf("Error writing command: %s", err)
	}

	// Read the response.
	reader := bufio.NewReader(conn.(io.Reader))
	msg, err := reader.ReadString(0)
	if err != nil {
		log.Fatalf("Could not read response: %s", err)
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

	// Try to build it from XDG
	xdg_run := os.Getenv("XDG_RUNTIME_DIR")
	disp := os.Getenv("DISPLAY")
	if len(xdg_run) > 0 && len(disp) > 0 {
		return fmt.Sprintf("%s/wingo/%s.0", xdg_run, disp)
	}

	// Eff it, just ask wingo where it is
	c := cmd.New("wingo", "--show-socket")
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(c.BufStdout.String())
}
