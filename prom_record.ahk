#NoEnv
#Persistent
#SingleInstance, force

SetMouseDelay, 10
SendMode, event

f12::
	clicktoggle := !clicktoggle

	if (!clicktoggle)
	{
	  SetTimer, startclick, off
	  return
	}

startclick:
	click
	sleep 500
	click
	sleep 500
	send, {p}
	sleep 500
	SetTimer, startclick, -500
	return
