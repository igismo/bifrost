//=============================================================================
// FILE NAME: tbMessages.go
// DESCRIPTION:
// Contains description of all possible messages
//
// NAME              REV  DATE       REMARKS			@
// Goran Scuric      1.0  01012019  Initial design
//================================================================================

package tbMessages

import (
	"net"
)

//      "os/user"
type TBmessage struct {
	MsgReceiver NameId
	MsgSender   NameId
	// MsgClass  string
	MsgType  string
	TimeSent string // msg send time
	MsgBody  []byte // Message type specific body
}

type TBmessageExpCreate struct {
	MsgReceiver NameId
	MsgSender   NameId
	// MsgClass  string
	MsgType  string
	TimeSent string    // msg send time
	MsgBody  ExpCreate // Message type specific body
}

type MsgData interface {
	area() float64
	perim() float64
}

type TBmgr struct {
	Name           NameId
	Up             bool
	LastChangeTime string
	MsgsSent       int64
	LastSentAt     string
	MsgsRcvd       int64
	LastRcvdAt     string
}

type NameId struct {
	Name        string // Name String
	OsId        int    // Task id, if known
	TimeCreated string // my incarnation time
	Address     net.UDPAddr
	Terminate   bool
}

type LinuxCommand struct {
	Cmd  string
	Par1 string
	Par2 string
	Par3 string
	Par4 string
	Par5 string
	Par6 string
}
type CommandList []LinuxCommand
type SatRouteTableChange []CommandList   // each set is per sat, list of commands
type ConstPosition []SatRouteTableChange // one set per sattelite

// Message source or destination
//----------------------------------------------------------------------------
const EXP_MGR = "ExpMgr"
const RSRC_MGR = "RsrcMgr"
const OFFICE_MGR = "OfficeMgr"

// Message Classes and types within classes:
const MSG_CLASS_CONTROL = "MSG_CLASS_CONTROL"

// Format of message body part, if any specific data ....
//----------------------------------------------------------------------------
const MSG_TYPE_INIT = "MSG_INIT"

type MsgInit struct {
	//
}

const MSG_TYPE_REGISTER = "REGISTER"

type MsgRegister struct {
	Mgr TBmgr
}

const MSG_TYPE_SAT_STATUS = "SAT STATUS"

type MsgSatStatus struct {
	Mgr       TBmgr
	SatStatus string
}

const MSG_TYPE_CMD = "COMMANDS"

type MsgCmd struct {
	Mgr  TBmgr
	cmds []LinuxCommand
}

const MSG_TYPE_CMD_REPLY = "CMD_REPLY"

type MsgCmdReply struct {
	Mgr      TBmgr
	CmdReply string
}

const MSG_TYPE_KEEPALIVE = "KEEPALIVE"

type MsgKeepAlive struct {
	tableOfMgrs []TBmgr
}

const MSG_TYPE_HELLO = "HELLO"

type MsgHello struct {
	tableOfMgrs []TBmgr
}

const MSG_TYPE_HELLO_REPPLY = "HELLO_REPLY"

type MsgHelloReply struct {
}

const MSG_TYPE_CONNECT = "CONNECT"

type MsgConnect struct {
}

const MSG_TYPE_CONNECTING = "CONNECTING"

type MsgConnecting struct {
}

const MSG_TYPE_CONNECTED = "CONNECTED"

type MsgConnected struct {
}

const MSG_TYPE_DISCONNECT = "DISCONNECT"

type MsgDisconnect struct {
}

const MSG_TYPE_DISCONNECTING = "DISCONNECTING"

type MsgDisconnecting struct {
}

const MSG_TYPE_DISCONNECTED = "MSG_DISCONNECTED"

type MsgDisconnected struct {
}

const MSG_TYPE_TERMINATE = "MSG_TERMINATE"

type MsgTerminate struct {
}

const MSG_TYPE_TERMINATING = "MSG_TERMINATING"

type MsgTerminating struct {
}

const MSG_TYPE_TERMINATED = "MSG_TERMINATED"

type MsgTerminated struct {
}

const MSG_CLASS_RSRCMGR = "MSG_CLASS_RSRCMGR"

//----------------------------------------------------------------------------

const MSG_CLASS_OFFICEMGR = "MSG_CLASS_OFFICEMGR"

//----------------------------------------------------------------------------
const MSG_TYPE_EXPCREATE = "EXPCREATE"

type ExpCreate struct {
	Project      string // $exp_pid
	Experiment   string // $exp_id
	GroupId      string // $exp_gid
	LinktestArgs string // $linktestarg
	ExtraGroups  string // $extragroups
	ExpSwappable string // $exp_swappable
	ExpDesc      string // $exp_desc
	BatchArgs    string // $batcharg
	UserName     string // $uid
	FileName     string // $thensfile
}
