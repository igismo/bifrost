//*********************************************************************************/
// NAME              REV  DATE       REMARKS			@
// Goran Scuric      1.0  01012018  Initial design     goran@usa.net
//================================================================================
/* UDPDaytimeClient
make function allocates and initializes an object of type slice, map, or chan only.
Like new, the first argument is a type. But, it can also take a second argument, the size.
Unlike new, make’s return type is the same as the type of its argument, not a pointer to it.
And the allocated value is initialized (not set to zero value like in new).
The reason is that slice, map and chan are data structures.
They need to be initialized, otherwise they won't be usable.
This is the reason new() and make() need to be different.
p := new(chan int)   // p has type: *chan int
c := make(chan int)  // c has type: chan int
p *[]int = new([]int) // *p = nil, which makes p useless
v []int = make([]int, 100) // creates v structure that has pointer to an array,
            length field, and capacity field. So, v is immediately usable
*/
package main

import (
	"bifrost/common/tbConfiguration"
	"bifrost/common/tbJsonUtils"
	"bifrost/common/tbLogUtils"
	"bifrost/common/tbMessages"
	"bifrost/common/tbMsgUtils"
	"bifrost/common/tbNetUtils"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	// "os/exec"
	// "log"
	//"bifrost/common/tbDbaseUtils"
	//"database/sql"
	//"testbedGS/common/tbExpUtils"
	//"database/sql"
)

// EXPERIMENT MANAGER STATES
const StateInit = "INIT"
const StateConnecting = "CONNECTING"
const StateConnected = "CONNECTED"
const StateUP = "UP"
const StateDOWN = "DOWN"

// Set the following 4 using specified attributes
var myName = "SatA" //tbConfig.TBexpMgrName
var myFullName tbMessages.NameId
var myUdpAddress = new(net.UDPAddr)
var myIpAddress = ""

var myIPandPort      = tbConfig.BifrostMasterIP + ":" + tbConfig.BifrostSatPort
var mastersIPandPort = tbConfig.BifrostMasterIPandPort

var myConnection *net.UDPConn = nil
var myState = StateInit

var myCreationTime = strconv.FormatInt(tbMsgUtils.TBtimestamp(), 10)
var myReceiveCount = 0
var myConnectionTimer = 0
var myLastKeepAliveReceived = time.Now()

var myRecvChannel chan []byte = nil               // To Receive messages from other modules
var myControlChannel chan []byte = nil            // so that all local threads can talk back
var satelliteSendChannel chan []byte = nil        // To Send messages out to other modules
var satelliteSendControlChannel chan []byte = nil // to send control msgs to Send Thread
var myRecvControlChannel chan []byte = nil        // to send control msgs to Recv Thread

var mastersUdpAddress = new(net.UDPAddr) // the address of the master control
var offMgrFullName tbMessages.NameId

var Log = tbLogUtils.LogInstance{}

// Arrays of satellites and soldiers learned
var sliceOfOtherSatellites []tbMessages.TBmgr
var sliceOfSoldiers []tbMessages.TBmgr

//====================================================================================
//
//====================================================================================
func main() {
	// TODO:  initialize IP, IPandPORT, mastersIpAddress, maybe bifrostPort
	Log.DebugLog = true
	Log.WarningLog = true
	Log.ErrorLog = true
	tbLogUtils.CreateLog(&Log, myName)
	Log.Warning(&Log, "this will be printed anyway")

	myInit()
	fmt.Println(myName, "MAIN: Starting a new ticker....")

	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for t := range ticker.C {
			//Call the periodic function here.
			periodicFunc(t)
		}
	}()
	//var msg tbMessages.TBmessage{}
	for {
		select {
		case msg1 := <-myRecvChannel:
			//fmt.Println(myName, "MAIN: DATA MSG in state", myState, "MSG=",string(msg1))
			handleMessages(msg1)

		case msg3 := <-myControlChannel: // ???
			fmt.Println(myName, "MAIN: Control msg in state", myState, "MSG=", string(msg3))
			handleControlMessages(msg3)
			// default:
			// fmt.Println("done and exit select")
		} // EndOfSelect
	}

	// os.Exit(0)
}

//====================================================================================
//
//====================================================================================
func periodicFunc(tick time.Time) {
	//fmt.Println("---->> EXP MASTER Tick", tick)
	//fmt.Println("TICK: myConnectionTimer=",myConnectionTimer)
	if myConnectionTimer > 0 {
		myConnectionTimer--
		//fmt.Println("TEST: myConnectionTimer=",myConnectionTimer)
		if myConnectionTimer == 0 {
			if locateTheMaster() == true {
				fmt.Println(myName, "CONNECTED TO MASTER")
				expSetState(StateConnected)
				sendRegisterMsg()
				myLastKeepAliveReceived = time.Now()
			} else {
				fmt.Println(myName, "NO CONNECTION TO MASTER")
				//fmt.Println("SET: myConnectionTimer=",myConnectionTimer)
				myConnectionTimer = 5 // 3*5=15 sec, check periodic timer above
			}
		}
	} else {
		currTime := time.Now()
		elapsedTime := currTime.Sub(myLastKeepAliveReceived)
		//fmt.Println("Elapsed time=", elapsedTime)
		if elapsedTime > (time.Second * 30) {
			if locateTheMaster() == true {
				expSetState(StateConnected)
				sendRegisterMsg()
				myLastKeepAliveReceived = time.Now()
			} else {
				expSetState(StateConnecting)
				myConnectionTimer = 5 // 3*5=15 sec, check periodic timer above
			}
		}
	}
}

//====================================================================================
// sendThread() - Thread sending our messages out
// The caller supplies the control channel over which
// control messages can be received by this thread
// Parameters:	service - 10.0.0.2:1200
// 				sendControlChannel - channel
//
//====================================================================================
func sendThread(conn *net.UDPConn, sendChannel, sendControlChannel chan []byte) error {
	var err error = nil
	fmt.Println(myName, "SendThread: Start SEND THRED")
	go func() {
		connection := conn
		var controlMsg tbMessages.TBmessage
		fmt.Println(myName, "SendThread: connected")

		myControlChannel <- tbMsgUtils.TBConnectedMsg(myFullName, myFullName, "")

		for {
			select {
			case msgOut := <-sendChannel: // got msg to send out
				// fmt.Println(myName, "SendThread: Sending MSG=",msgOut)
				fmt.Println(myName, "SendThread: Sending MSG out to", mastersUdpAddress)
				// _, err = connection.Write([]byte(msgOut))
				_, err = connection.WriteToUDP(msgOut, mastersUdpAddress)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error Sending %s", err.Error())
					// create more descriptive msg
					// send msg up to indicate a problem ?
				}

			case ctrlMsg := <-sendControlChannel: //
				_ = tbJsonUtils.TBunmarshal(ctrlMsg, &controlMsg)
				fmt.Println(myName, "SendThread got control MSG=", controlMsg)

				if strings.Contains(controlMsg.MsgType, "TERMINATE") {
					fmt.Println(myName, "SendThread rcvd control MSG=", controlMsg)
					return
				}
			}
		}

	}()

	return err
}

//====================================================================================
// recvThread - Thread receiving messages from others
//====================================================================================
func RecvThread(conn *net.UDPConn, recvControlChannel <-chan []byte) error {
	var err error = nil

	//fmt.Println(myName,"RecvThread: Start RECV THRED")
	go func() {
		connection := conn

		fmt.Println(myName, "RecvThread: Start Receiving")
		var controlMsg tbMessages.TBmessage
		var oobBuffer [3000]byte
		// Tell main we are coonected all is good
		// myControlChannel <- tbMsgUtils.TBConnectedMsg(myFullName, myFullName, "")

		for {
			recvBuffer := make([]byte, 3000)
			// length, oobn, flags, addr, err := connection.ReadMsgUDP(recvBuffer[0:], oobBuffer[0:])
			length, _, _, _, _ := connection.ReadMsgUDP(recvBuffer[0:], oobBuffer[0:])
			myReceiveCount++
			//fmt.Println(myName, "\n============== Receive Count=", myReceiveCount,
			//	"\nRecvThread UDP MSG from", addr, "len=", length, "oobLen=", oobn, "flags=", flags, "ERR=", err)
			// fmt.Println(myName,"RecvThread MSG=", string(recvBuffer[0:length]))

			myRecvChannel <- recvBuffer[0:length]

			if len(recvControlChannel) != 0 {
				ctrlMsg := <-recvControlChannel
				_ = tbJsonUtils.TBunmarshal(ctrlMsg, &controlMsg)
				fmt.Println("RecvThread got CONTROL MSG=", controlMsg)
				if strings.Contains(controlMsg.MsgType, "TERMINATE") {
					fmt.Println("RecvThread rcvd control MSG=", controlMsg)
					return
				}
			}
		}
	}()

	return err
}

//====================================================================================
//
//====================================================================================
func handleMessages(message []byte) {
	// Unmarshal
	//msg := new(tbMessages.TBmessage)

	//tbJsonUtils.TBunmarshal(message, &msg)
	// fmt.Println(myName,"HandleMessages MSG=", string(message),"Sizeof(msg)=",unsafe.Sizeof(msg))
	//fmt.Println(myName,":HandleMessages MSG Type:",msg.MsgType, " From:",msg.MsgSender.Name," To Me:",msg.MsgReceiver.Name)
	//fmt.Println(myName, ":HandleMessages MAIN: BODY=", msg.MsgBody)
	//fmt.Println(myName, ":HandleMessages MAIN: TIMESENT=", msg.TimeSent)

	switch myState {
	case StateInit:
		stateInitMessages(message)
		break
	case StateConnecting:
		stateConnectingMessages(message)
		break
	case StateConnected:
		stateConnectedMessages(message)
		break
	case StateUP:
	case StateDOWN:
		stateConnectedMessages(message)
		break
	default:
	}
}

//====================================================================================
// STATE = INIT, nothing should be really happening here
//====================================================================================
func stateInitMessages(message []byte) {
	msg := new(tbMessages.TBmessage)

	_ = tbJsonUtils.TBunmarshal(message, &msg)
	messageType := msg.MsgType
	switch messageType {
	case tbMessages.MSG_TYPE_CONNECTED:
	default:

	}
}

//====================================================================================
// STATE=CONNECTING to OFFICE MANAGER
//====================================================================================
func stateConnectingMessages(message []byte) {
	msg := new(tbMessages.TBmessage)

	_ = tbJsonUtils.TBunmarshal(message, &msg)
	messageType := msg.MsgType
	switch messageType {
	case tbMessages.MSG_TYPE_HELLO:
		expSetState(StateConnected)
		receiver := msg.MsgSender
		// GS
		newMsg := tbMsgUtils.TBhelloReplyMsg(myFullName, receiver, "")

		// satelliteSendChannel <- mymsg
		fmt.Println(myName, "stateConnectingMessages: sendMsgOut ")
		tbMsgUtils.TBsendMsgOut(newMsg, receiver.Address, myConnection)
		fmt.Println(myName, "State=", myState, " Send MSG to=", receiver)

	default:
	}

}

//====================================================================================
//
//====================================================================================
func stateConnectedMessages(message []byte) {
	msg := new(tbMessages.TBmessage)
	_ = tbJsonUtils.TBunmarshal(message, &msg)

	messageType := msg.MsgType
	//fmt.Println("RCVD MSG=", msg)
	switch messageType {
	case tbMessages.MSG_TYPE_KEEPALIVE:
		// GS set last hello received msg time to check within periodic timer
		// that Office Master is alive
		fmt.Println("..... KEEP ALIVE MESSAGE FROM MASTER")
		receivedKeepAliveMsg(msg)
		myLastKeepAliveReceived = time.Now()
		sendRegisterMsg()
		break
	case tbMessages.MSG_TYPE_CMD:
		// fmt.Println("..... CMD MESSAGE FROM MASTER")
		commandMessages(msg)
		break
	case tbMessages.MSG_TYPE_TERMINATE:
		fmt.Println("..... TERMINATE MESSAGE FROM MASTER")
		os.Exit(0)
		break
	default:
	}
}
func commandMessages(msg *tbMessages.TBmessage) {
	var cmds []tbMessages.LinuxCommand
	_ = tbJsonUtils.TBunmarshal(msg.MsgBody, &cmds)
	//fmt.Println("RCVD COMMANDS=", cmds)
	//
	//
	for cmdIndex := range cmds {
		var cmd tbMessages.LinuxCommand
		cmd = cmds[cmdIndex]
		//fmt.Println("RCVD CMD ", cmdIndex, " =",cmd)
		fmt.Println("CMD=", cmd.Cmd, " ", cmd.Par1, " ", cmd.Par2, " ", cmd.Par3, " ", cmd.Par4, " ", cmd.Par5, " ", cmd.Par6)
		//cmd.Output() → run it, wait, get output
		//cmd.Run() → run it, wait for it to finish.
		//cmd.Start() → run it, don't wait. err = cmd.Wait() to get result.

		var thisCmd = exec.Command(cmd.Cmd, cmd.Par1, cmd.Par2, cmd.Par3, cmd.Par4, cmd.Par5, cmd.Par6)
		//output, err := thisCmd.Output()

		output, _ := thisCmd.Output()
		//if err != nil && err.Error() != "exit status 1" {
		//	fmt.Println("CMDx=", cmd.Cmd, " ", cmd.Par1, " ", cmd.Par2, " ", cmd.Par3, " ", cmd.Par4,
		//		" ", cmd.Par5, " ", cmd.Par6, " :  cmd.Run() failed with ", err)
		//} else {
			fmt.Println("CMDy=", cmd.Cmd, " ", cmd.Par1, " ", cmd.Par2, " ", cmd.Par3, " ", cmd.Par4,
				" ", cmd.Par5, " ", cmd.Par6, " :  OUTPUT:", string(output))
		//}
		//if err != nil && err.Error() != "exit status 1" {
		//	//panic(err)
		//	//fmt.Printf("ERROR=", err, "\n")
		//	fmt.Printf("%T\n", err)
		//} else {
		//	fmt.Printf("CMD OUTPUT=",string(output))
		//	// SEND REEPLY, OR MAYBE COMBINED ALL FIRST
		//}
	}
	// var cmd = satInfo[0]
}

//====================================================================================
//
//====================================================================================
func handleControlMessages(message []byte) {
	// Unmarshal
	var msg tbMessages.TBmessage
	_ = tbJsonUtils.TBunmarshal(message, &msg)
	//fmt.Println(myName, "HandleControlMessages MSG(",unsafe.Sizeof(msg),")=", msg)
	fmt.Println(myName, "HandleControlMessages MSG Type=", msg.MsgType)
	switch myState {
	case StateInit:
		stateInitControlMessages(msg)
		break
	case StateConnecting:
		break
	case StateConnected:
		stateConnectedControlMessages(msg)
		break
	case StateUP:
		break
	case StateDOWN:
		break
	default:
	}
}

//====================================================================================
//
//====================================================================================
func stateInitControlMessages(msg tbMessages.TBmessage) {
	messageType := msg.MsgType
	switch messageType {
	case tbMessages.MSG_TYPE_CONNECTED:
		fmt.Println(myName, "HMMM ... wrong msg in state", myState)
		// expSetState(StateConnected)
		// send helloMsg to remote server
		// ERROR
	default:
	}
}

//====================================================================================
//
//====================================================================================
func stateConnectedControlMessages(msg tbMessages.TBmessage) {
	messageType := msg.MsgType
	switch messageType {
	case tbMessages.MSG_TYPE_CONNECTED:
		expSetState(StateConnected)
		sendRegisterMsg()
		break

	default:
	}
}

//=======================================================================
//
//=======================================================================
func receivedKeepAliveMsg(msg *tbMessages.TBmessage) {
	var rcvdSliceOfMgrs []tbMessages.TBmgr
	_ = tbJsonUtils.TBunmarshal(msg.MsgBody, &rcvdSliceOfMgrs)
	var names = ""
	for mgrIndex := range rcvdSliceOfMgrs {
		receiver := rcvdSliceOfMgrs[mgrIndex].Name

		ip := rcvdSliceOfMgrs[mgrIndex].Name.Address.IP.String()
		port := rcvdSliceOfMgrs[mgrIndex].Name.Address.Port

		names += " " + receiver.Name + " at " + ip + ":" + strconv.Itoa(port)

	}
	fmt.Println(myName, ": LEARNED MODULES=", names)
}

//=======================================================================
//
//=======================================================================
func stateConnectingControlMessages(msg tbMessages.TBmessage) {
	messageType := msg.MsgType
	switch messageType {
	case tbMessages.MSG_TYPE_CONNECTED:
		expSetState(StateConnected)
		// satelliteSendChannel <- tbMsgUtils.TBhelloMsg(myFullName, offMgrFullName, "ABCDEFG")
		newMsg := tbMsgUtils.TBhelloMsg(myFullName, offMgrFullName, "ABCDEFG")
		fmt.Println(myName, "stateConnectingControlMessages: sendMsgOut ")
		tbMsgUtils.TBsendMsgOut(newMsg, *mastersUdpAddress, myConnection)
		break

	default:
	}

}

//====================================================================================
//
//====================================================================================
func expSetState(newState string) {
	fmt.Println(myName, "OldState=", myState, " NewState=", newState)
	myState = newState
}

//====================================================================================
//
//====================================================================================
func myInit() {
	var err error
	fmt.Println(myName, "INIT: expMgr Init at ", myCreationTime)

	satelliteSendControlChannel = make(chan []byte) // so that we can talk to sendThread
	myRecvControlChannel = make(chan []byte)        // so that we can talk to recvThread
	satelliteSendChannel = make(chan []byte)        // so that we can talk to sendThread
	myRecvChannel = make(chan []byte)               //
	myControlChannel = make(chan []byte)            // so that all threads can talk to us

	myIpAddress = tbNetUtils.GetLocalIp()
	myIPandPort = myIpAddress + ":" + tbConfig.BifrostSatPort

	myUdpAddress, _ = net.ResolveUDPAddr("udp", myIPandPort)

	fmt.Println(myName, "INIT: My Local IP=", myIpAddress, " My UDP address=", myUdpAddress)

	myFullName = tbMessages.NameId{Name: myName, OsId: os.Getpid(),
		TimeCreated: myCreationTime, Address: *myUdpAddress}
	fmt.Println(myName, "INIT: myFullName=", myFullName)

	myConnection, err = net.ListenUDP("udp", myUdpAddress) // from officeMgr
	checkError(err)

	myEntry := tbMessages.TBmgr{Name: myFullName, Up: true, LastChangeTime: myCreationTime,
		MsgsSent: 0, LastSentAt: "0", MsgsRcvd: 0, LastRcvdAt: "0"}
	sliceOfOtherSatellites = append(sliceOfOtherSatellites, myEntry)

	fmt.Println(myName, "SAT:", myEntry.Name.Name, "ADDRESS:", myEntry.Name.Address,
		"CREATED:", myEntry.Name.TimeCreated, "MSGSRCVD:", myEntry.MsgsRcvd)

	//err1 := sendThread(myConnection, satelliteSendChannel, satelliteSendControlChannel)
	//if err1 != nil {
	//	fmt.Println(myName,"INIT: Error creating send thread")
	//}
	err2 := RecvThread(myConnection, myRecvControlChannel)
	if err2 != nil {
		fmt.Println(myName, "INIT: Error creating Receive thread")
	}

	expSetState(StateConnecting)

	if locateTheMaster() == true {
		fmt.Println(myName, "INIT: MASTER LOCATED")
		expSetState(StateConnected)
		sendRegisterMsg()
		myLastKeepAliveReceived = time.Now()
	} else {
		myConnectionTimer = 5 // 3*5=15 sec, check periodic timer above
	}
	//fmt.Println(myName, "++++++++++++++++++++++++++++++++++")
	//fmt.Println(myName, "++++++++++++++++++++++++++++++++++")
	//fmt.Println(myName, "START HANDLE CREATE")
}

//====================================================================================
//
//====================================================================================
func locateTheMaster() bool {
	// NOTE that this will fail if mastersUdpAddress has not been initialized due
	// to master not being up . Add state and try again to Resolve before
	// doing this
	var err error
	//fmt.Println("Locate master control, not ground")
	mastersUdpAddress, err = net.ResolveUDPAddr("udp", mastersIPandPort)
	if err != nil {
		fmt.Println("ERROR in net.ResolveUDPAddr = ", err)
		fmt.Println("ERROR locating master, will retry")
		return false
	}

	offMgrFullName = tbMessages.NameId{Name: mastersIPandPort, OsId: 0,
		TimeCreated: "0", Address: *mastersUdpAddress}
	fmt.Println(myName, "INIT: masterFullName=", offMgrFullName)
	entry2 := tbMessages.TBmgr{Name: offMgrFullName, Up: true, LastChangeTime: "0",
		MsgsSent: 0, LastSentAt: "0", MsgsRcvd: 0, LastRcvdAt: "0"}
	sliceOfOtherSatellites = append(sliceOfOtherSatellites, entry2)
	// fmt.Println("New SLICE of Managers=", sliceOfOtherSatellites)

	theMgr := locateOtherSatellite(sliceOfOtherSatellites, mastersIPandPort)
	if theMgr != nil {
		fmt.Println("SAT:", theMgr.Name.Name, "ADDRESS:", theMgr.Name.Address,
			"CREATED:", theMgr.Name.TimeCreated, "MSGSRCVD:", theMgr.MsgsRcvd)
	}
	return true
}

//====================================================================================
//
//====================================================================================
func formatReceiver(name string, osId int, udpAddress net.UDPAddr) tbMessages.NameId {
	receiver := tbMessages.NameId{Name: name, OsId: osId,
		TimeCreated: "", Address: udpAddress}
	return receiver
}

//====================================================================================
// Save the pointer to my own row for faster handling
//====================================================================================
func sendRegisterMsg() {
	theMgr := locateOtherSatellite(sliceOfOtherSatellites, myName)
	msgBody, _ := tbJsonUtils.TBmarshal(theMgr)
	if theMgr != nil {
		theMgr.LastSentAt = strconv.FormatInt(tbMsgUtils.TBtimestamp(), 10)
		newMsg := tbMsgUtils.TBregisterMsg(myFullName, offMgrFullName, string(msgBody))
		// fmt.Println(myName, "stateConnected REGISTER with offMgr ")
		tbMsgUtils.TBsendMsgOut(newMsg, *mastersUdpAddress, myConnection)
	} else {
		fmt.Println("FAILED to locate a record for master in the slice")
	}
}

//====================================================================================
//
//====================================================================================
func sendHelloReplyMsg(msg *tbMessages.TBmessage) {
	receiver := msg.MsgSender
	newMsg := tbMsgUtils.TBhelloMsg(myFullName, receiver, "ABCDEFG")
	tbMsgUtils.TBsendMsgOut(newMsg, receiver.Address, myConnection)
}

//====================================================================================
//
//====================================================================================
func checkError(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Fatal error %s", err.Error())
		os.Exit(1)
	}
	//time.Sleep(time.Millisecond * 3000)
}

//============================================================================
// Locate specific satellite in the slice of all learned satellites
// Return nil if row not found
//============================================================================
func locateOtherSatellite(slice []tbMessages.TBmgr, satellite string) *tbMessages.TBmgr {
	for index := range slice {
		if slice[index].Name.Name == satellite {
			return &slice[index]
		}
	}
	return nil
}
