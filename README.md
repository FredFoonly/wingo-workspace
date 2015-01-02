wingo-workspace
===============

Generate a workspace line for gobar in a file someplace.  This can be
sent to gobar with tail -f.  This needs my version of wingo for
$WINGO_SOCKET support, multi-command key definitions, and improved
variable substition.

Options:
  -f          When set, will be the file where the gobar-formatted workspace
              line is written.
  -cur-fg	  When set, will be the foreground color for the current workspace.
  -other-fg   When set, will be the foreground color for non-current workspaces.

It will pick up the wingo UDS IPC socket from either the $WINGO_SOCKET
environment variable (set by wingo), or by replicating wingo's path
creation logic using $XDG_RUNTIME_DIR and $DISPLAY, or as a last
resort by running 'wingo --show-socket'.

I have wingo, gobar, and wingo-workspace set up as follows:

.xsession
--------------
#!/usr/local/bin/rc
wingo_wsp_file=$home/run/wingo/wsp_line
{sleep 2; $home/bin/wingo-workspace -cur-fg 00ef0707 -f $wingo_wsp_file}&
{tail -1 -f $wingo_wsp_file >[2] /dev/null} | $home/bin/gobar&
exec $home/bin/wingo -socket-path $home/run/wingo
--------------

key.wini:
--------------
[Global]
# $wingo_wsp_file comes in from .xsession
$wsp_file := $wingo_wsp_file
$wsp_curfg := 00ef0707
$wsp_nonfg := 000707ef

Mod1-0 := {WorkspaceGreedy "0"; Shell "wingo-workspace -cur-fg $wsp_curfg -f $wsp_file"}
Mod1-1 := {WorkspaceGreedy "1"; Shell "wingo-workspace -cur-fg $wsp_curfg -f $wsp_file"}
Mod1-2 := {WorkspaceGreedy "2"; Shell "wingo-workspace -cur-fg $wsp_curfg -f $wsp_file"}
...
Mod1-2 := {WorkspaceGreedy "9"; Shell "wingo-workspace -cur-fg $wsp_curfg -f $wsp_file"}
--------------


Whenever I switch workspaces with Alt-[0-9], it runs wingo-workspace
to generate a new workspace line in a known file.  The background tail
process that was launched in .xsession picks this up and pipes it into
gobar.


TODO:

Change wingo to support http as an ipc protocol.  I want to eventually
hack wingo to support a plan-9 style of IPC where REST calls allow
various types of state inspection and modification, and html5
websockets can be used for things like state change notifications and
command/response streams.  This will also allow wingo-workspace to
start up in .xsession and just stay resident and pipe workspace lines
to gobar.
