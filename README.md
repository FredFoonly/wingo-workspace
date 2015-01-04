wingo-workspace
===============

Generate a workspace line for gobar.  This needs my version of wingo for
figuring out the UDS IPC and notification sockets.

Options    | Description
-----------|------------
-cur-fg	   | When set, will be the foreground color for the current workspace.
-other-fg  | When set, will be the foreground color for non-current workspaces.

It will pick up the wingo UDS IPC and notify sockets from either the
$WINGO_SOCKET and $WINGO_NOTIFY_SOCKET environment variables (set by
wingo), or by replicating wingo's path creation logic using
$XDG_RUNTIME_DIR and $DISPLAY, or as a last resort by running 'wingo
--show-socket' and 'wingo --show-notify-socket'.

I have wingo, gobar, and wingo-workspace set up as follows:

.xsession
```
#!/usr/local/bin/rc
{$home/bin/wingo-workspace '-cur-fg=00ef0707' | $home/bin/gobar '--geometry=Mx20+0+0' '--fonts=/usr/local/lib/X11/fonts/terminus:16' --bottom}&
exec $home/bin/wingo -socket-path $home/run/wingo
```


