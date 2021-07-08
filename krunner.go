package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
)

const introspectString = `
<node>
  <interface name="org.kde.krunner1">
    <method name="Actions">
      <annotation name="org.qtproject.QtDBus.QtTypeName.Out0" value="RemoteActions" />
      <arg name="matches" type="a(sss)" direction="out" />
    </method>
    <method name="Run">
      <arg name="matchId" type="s" direction="in"/>
      <arg name="actionId" type="s" direction="in"/>
    </method>
    <method name="Match">
      <arg name="query" type="s" direction="in"/>
      <annotation name="org.qtproject.QtDBus.QtTypeName.Out0" value="RemoteMatches"/>
      <arg name="matches" type="a(sssuda{sv})" direction="out"/>
    </method>
  </interface>` + introspect.IntrospectDataString + `</node>`

type RemoteAction struct {
	ID, Text, IconName string
}

type RemoteMatch struct {
	ID, Text, IconName string
	Type               int32
	Relevance          float64
	Properties         map[string]interface{}
}

type Runner struct{}

func (r Runner) Match(query string) ([]RemoteMatch, *dbus.Error) {
	if strings.HasPrefix(strings.ToLower(query), "caffeinate") || strings.HasPrefix(strings.ToLower(query), "caff") {
		querySplit := strings.Split(strings.ToLower(query), " ")
		// if no duration is specified (there's no space) then we assume it's until disabled
		if len(querySplit) == 1 {
			return []RemoteMatch{
				{
					ID:        "-1",
					Text:      "Caffeinate until disabled",
					IconName:  "accept_time_event",
					Type:      200,
					Relevance: 1,
				},
			}, nil
		}
		duration, err := time.ParseDuration(querySplit[1])
		// if the duration cannot be parsed then we dont accept
		if err != nil {
			return make([]RemoteMatch, 0), nil
		}
		return []RemoteMatch{
			{
				ID:        strconv.FormatInt(int64(duration), 10),
				Text:      "Caffeinate for " + querySplit[1], //if we can parse then it's valid
				IconName:  "accept_time_event",
				Type:      200,
				Relevance: 1,
			},
		}, nil
	}
	// got no matches
	return make([]RemoteMatch, 0), nil
}

func (r Runner) Actions() ([]string, *dbus.Error) {
	return make([]string, 0), nil
}

func (r Runner) Run(matchId string, actionId string) *dbus.Error {
	if matchId == "-1" {
		//TODO infinte caffeination
	}
	duration, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		return &dbus.ErrMsgInvalidArg
	}
	conn, err := dbus.SessionBus()
	if err != nil {
		// is this the correct error to return?
		return &dbus.Error{}
	}

	var pmCookie, ssCookie uint32
	err = conn.Object("org.freedesktop.PowerManagement", "/org/freedesktop/PowerManagement/Inhibit").Call("Inhibit", 0, "Caffeinate", "user triggered").Store(&pmCookie)
	if err != nil {
		fmt.Printf("Failed to call org.freedesktop.PowerManagement.Inhibit with error: %s\n", err.Error())
		// is this the correct error to return?
		return &dbus.Error{}
	}
	err = conn.Object("org.freedesktop.ScreenSaver", "/org/freedesktop/ScreenSaver").Call("org.freedesktop.ScreenSaver.Inhibit", 0, "Caffeinate", "user triggered").Store(&ssCookie)
	if err != nil {
		fmt.Printf("Failed to call org.freedesktop.ScreenSaver.Inhibit with error: %s\n", err.Error())
		// is this the correct error to return?
		return &dbus.Error{}
	}
	fmt.Printf("powermanagement inhibited with cookie %d and screensaver with cookie %d for %dns\n", pmCookie, ssCookie, duration)
	go func() {
		time.Sleep(time.Duration(duration))
		pmErr := conn.Object("org.freedesktop.PowerManagement", "/org/freedesktop/PowerManagement/Inhibit").Call("UnInhibit", 0, pmCookie).Err
		ssErr := conn.Object("org.freedesktop.ScreenSaver", "/org/freedesktop/ScreenSaver").Call("org.freedesktop.ScreenSaver.UnInhibit", 0, ssCookie).Err
		if pmErr == nil && ssErr == nil {
			fmt.Printf("uninhibited powermanagement %d and screensaver %d\n", pmCookie, ssCookie)
		} else {
			fmt.Printf("failure to uninhibit, error: %s\n", err.Error())
			// handle error in goroutine
		}
	}()
	return nil
}

func main() {
	// connect to session dbus
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	f := Runner{}
	conn.Export(f, "/krunner", "org.kde.krunner1")
	conn.Export(introspect.Introspectable(introspectString), "/krunner", "org.freedesktop.DBus.Introspectable")

	reply, err := conn.RequestName("uk.matthew-jones.krunner-caffeinate", dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		panic("Name already taken")
	}
	select {}
}
