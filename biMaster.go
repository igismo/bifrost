//*********************************************************************************/
// NAME              REV  DATE       REMARKS			@
// Goran Scuric      1.0  10202019  Initial design     goran.scuric@aero.org
//================================================================================
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
)

// OFFICE MANAGER STATES
const MasterInit = "MasterInit"
const MasterConnecting = "MasterConnecting"
const MasterConnected = "MasterConnected"
const MasterUP = "MasterUP"
const MasterDOWN = "MasterDOWN"

var masterState = MasterInit
var masterName = "biMaster"
var masterFullName tbMessages.NameId // struct - needs init
var masterCreationTime = strconv.FormatInt(tbMsgUtils.TBtimestamp(), 10)

// Must initialize following
var masterUdpAddress *net.UDPAddr = nil
var masterIpAddress = ""
var masterIPandPort = ""
var masterConnection *net.UDPConn = nil

var masterSendChannel chan []byte = nil // To Send messages out to other modules
var masterRecvChannel chan []byte = nil // To Receive messages from other modules
// var masterControlChannel     chan []byte = nil // so that all local threads can talk back
var masterSendControlChannel chan []byte = nil // to send control msgs to Send Thread
var masterRecvControlChannel chan []byte = nil // to send control msgs to Recv Thread
var masterTimerChannel chan string = nil       // timer ticks
var masterReceiveCount = 0

var masterLog = tbLogUtils.LogInstance{}

// var masterTicker      *time.Ticker = nil
var sliceOfSatellites []tbMessages.TBmgr

//--------------- -----------------
// ROUTE AND INTERFACE CHANGES
//---------------------------------
// type CommandList         [] LinuxCommand
// type SatRouteTableChange [] CommandList	// each set is per sat, list of commands
// type ConstPosition       [] SatRouteTableChange // one set per sattelite

//var satRoutePosition1 = []tbMessages.CommandList {// each set is per sat, list of commands
//tbMessages.CommandList { // SatA position
//tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2:"", Par3:"", Par4:"",Par5:"", Par6:""},
//tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2:"", Par3:"", Par4:"",Par5:"", Par6:""},
//},
//tbMessages.CommandList { // SatB position 1
//tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2:"", Par3:"", Par4:"",Par5:"", Par6:""},
//tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2:"", Par3:"", Par4:"",Par5:"", Par6:""},
//},
//}
//var satRoutePosition2 = []tbMessages.CommandList {}
//var satRoutePosition3 = []tbMessages.CommandList {}
//var satRoutePosition4 = []tbMessages.CommandList {}
//
//var satPositions = [] tbMessages.SatRouteTableChange{ // one set per sattelite per position
//satRoutePosition1,
//satRoutePosition2,
//satRoutePosition3,
//satRoutePosition4,/
//}

var satRouteInfo = tbMessages.ConstPosition{ // 4 positions for now
	tbMessages.SatRouteTableChange{ // Position 1
		tbMessages.CommandList{ // Sat A
			tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2: "", Par3: "", Par4: "", Par5: "", Par6: ""},
			tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2: "", Par3: "", Par4: "", Par5: "", Par6: ""},
			tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2: "", Par3: "", Par4: "", Par5: "", Par6: ""},
		},
		tbMessages.CommandList{ // Sat B
			tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2: "", Par3: "", Par4: "", Par5: "", Par6: ""},
		},
		tbMessages.CommandList{ // Sat C
			tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2: "", Par3: "", Par4: "", Par5: "", Par6: ""},
		},
		tbMessages.CommandList{ // Sat D
			tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2: "", Par3: "", Par4: "", Par5: "", Par6: ""},
		},
	},
	tbMessages.SatRouteTableChange{ // Position 2
		tbMessages.CommandList{ // Sat A
			tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2: "", Par3: "", Par4: "", Par5: "", Par6: ""},
		},
		tbMessages.CommandList{ // Sat B
			tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2: "", Par3: "", Par4: "", Par5: "", Par6: ""},
		},
		tbMessages.CommandList{ // Sat C
			tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2: "", Par3: "", Par4: "", Par5: "", Par6: ""},
		},
		tbMessages.CommandList{ // Sat D
			tbMessages.LinuxCommand{Cmd: "", Par1: "", Par2: "", Par3: "", Par4: "", Par5: "", Par6: ""},
		},
	},
}

// var len = len(satRouteInfo[0][0])
// var cmd = satRouteInfo[0][1][0]

//=======================================================================
// Enry point for the office Master
// Note that the log is created, but logging is stil outstanding work
//=======================================================================
func main() {
	tbLogUtils.LogLicenseNotice()

	fmt.Println(masterName, "========= START =========================")
	argsWithProg := os.Args
	argsWithoutProg := os.Args[1:]

	for index := range os.Args {
		arg := os.Args[index]
		fmt.Println("Arg", index, "=", arg)
	}

	fmt.Println("Command line=", argsWithProg)
	fmt.Println("Arguments=", argsWithoutProg)

	//////
	fmt.Printf("%v\n", "chdir to /users/scuric  -----------------------------\n")
	var dirToRun = "/users/scuric"
	var err = os.Chdir(dirToRun)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", "START EXEC -----------------------------\n")

	// cmd.Output() → run it, wait, get output
	// cmd.Run() → run it, wait for it to finish.
	// cmd.Start() → run it, don't wait. err = cmd.Wait() to get result.

	var cmd = exec.Command("ifconfig", "-l") // en0
	output, err := cmd.Output()
	if err != nil {
		//panic(err)
		fmt.Printf("ERROR=", err, "\n")
	}
	// ifconfig en0 inet 192.0.2.45/28 add
	// ifconfig en0 inet 192.0.2.45 -alias
	// ifconfig gif1 create
	// sudo ifconfig gif1 10.2.2.2/24 10.2.2.3
	//  ??? addm interface
	// netstat -nr -f inet
	// netstat -nr -f inet -I en0
	// netstat -nr -f inet -I en0 -S   ... show status

	fmt.Printf("OUTPUT=\n%v\n", string(output))
	fmt.Printf("%v\n", "END EXEC -----------------------------\n")
	///////
	masterLog.DebugLog = true
	masterLog.WarningLog = true
	masterLog.ErrorLog = true
	tbLogUtils.CreateLog(&masterLog, masterName)
	masterLog.Warning(&masterLog, "this will be printed anyway")
	if false {
		officeMgrSetState(MasterInit)

		officeMasterInit()

		masterTicker := time.NewTicker(15 * time.Second)

		go func() {
			// fmt.Println(masterName,"MAIN: Starting a new ticker....")
			for t := range masterTicker.C {
				MasterPeriodicFunc(t)
			}
		}()

		consoleInput := make(chan string)
		startConsole(consoleInput)

		fmt.Println(masterName, "MAIN: Initialized, start select loop")

		for {
			select {
			case msg1 := <-masterRecvChannel:
				//fmt.Println(masterName, "MAIN: RecvChannel msg")
				masterHandleMessages(msg1)

			case msg2 := <-masterTimerChannel:
				//fmt.Println(masterName, "MAIN: TimerChannel msg=", msg2)
				handleTimerMessages(msg2)
				// default:
				// fmt.Println("done and exit select")

			case stdin, ok := <-consoleInput:
				if !ok {
					fmt.Println("ERROR Reading input from stdin:", stdin)
					break
				} else {
					fmt.Println("Read input from stdin:", stdin)
				}
			} // EndOfSelect
		}
	} // END OF IF FALSE
}

//=======================================================================
// Some notes for later:
// Just create /etc/resolv.conf and append nameserver 8.8.8.8 then this
// problem will be resolved. According to src/net/dnsclient_unix.go,
// if /etc/resolv.conf is absent, localhost:53 is chosen as a name server.
// Since the Linux in Android is not so "standard". /etc/resolv.conf is not available.
// The app then just keep looking up host in localhost:53.
//=======================================================================
func officeMasterInit() {
	var err error = nil

	_, _ = fmt.Println(masterName, "INIT0: create channels")
	masterTimerChannel = make(chan string)       // Timer ticks
	masterSendControlChannel = make(chan []byte) // so that we can talk to masterSendThread
	masterRecvControlChannel = make(chan []byte) // so that we can talk to masterRecvThread
	masterSendChannel = make(chan []byte)        // so that we can talk to masterSendThread
	masterRecvChannel = make(chan []byte)        // rcv messages from the universe
	//masterControlChannel     = make(chan []byte) // so that all threads can talk to us

	// conn, err := net.ListenUDP("udp", udpAddr)
	//
	// n, addr, err := conn.ReadFromUDP(buf[0:])
	// conn.WriteToUDP([]byte(daytime), addr)
	masterIpAddress = tbNetUtils.GetLocalIp()
	masterIPandPort = masterIpAddress + ":" + tbConfig.BifrostPort
	fmt.Println(masterName, "INIT: masterIpAddress=", masterIpAddress, " masterIPandPort=", masterIPandPort)

	if masterConnection == nil {
		_, _ = fmt.Println(masterName, "INIT1: masterConnection created ", masterIPandPort)

		masterUdpAddress, err = net.ResolveUDPAddr("udp", masterIPandPort)
		if err != nil {
			_, _ = fmt.Println(masterName, "INIT2: ERROR in net.ResolveUDPAddr = ", err)
			_, _ = fmt.Println(masterName, "INIT3: ERROR locating Office Manager, will retry")
			//return false
		}
		fmt.Println(masterName, "INIT4: masterUdpAddress=", masterUdpAddress)

		// conn, err := net.DialUDP("udp", nil, masterUdpAddress)
		masterConnection, err = net.ListenUDP("udp", masterUdpAddress)
		masterCheckError(err)
		fmt.Println(masterName, "INIT: masterConnection=", masterConnection)

		err1 := masterSendThread(masterConnection, masterSendChannel, masterSendControlChannel)
		if err1 != nil {
			fmt.Println(masterName, "INIT: Error creating send thread")
		}

		err2 := masterRecvThread(masterConnection, masterRecvControlChannel)
		if err2 != nil {
			fmt.Println(masterName, "INIT: Error creating Receive thread")
		}

		if err1 != nil || err2 != nil {
			return
		}

		masterFullName = tbMessages.NameId{Name: masterName, OsId: os.Getpid(),
			TimeCreated: masterCreationTime, Address: *masterUdpAddress}
		fmt.Println(masterName, "INIT: masterFullName=", masterFullName)

	}

	officeMgrSetState(MasterConnected)

	fmt.Println(masterName, "INIT: Office Master Start Receive at", masterCreationTime)
}

//=======================================================================
//
//=======================================================================
func MasterPeriodicFunc(tick time.Time) {
	// GS send keepAlive messages at whatever interval
	_, _ = fmt.Println(masterName, "Tick at: ", tick)
	sendKeepAliveMsg()
}

//=================================================
// masterSendThread() - Thread sending our messages out
// The caller supplies the control channel over which
// control messages can be received by this thread
// Parameters:	service - 10.0.0.2:1200
// 				sendControlChannel - channel
//=======================================================================
//
//=======================================================================
func masterSendThread(conn *net.UDPConn, sendChannel, sendControlChannel chan []byte) error {
	var err error = nil
	fmt.Println(masterName, "SendThread: Start SEND THRED")
	go func() {
		connection := conn
		var controlMsg tbMessages.TBmessage
		fmt.Println(masterName, "SendThread: Ready for Sending")
		//masterControlChannel <- tbMsgUtils.TBConnectedMsg(masterCreationTime)

		for {
			select {
			case msgOut := <-sendChannel: // got msg to send out
				fmt.Println(masterName, "SendThread: Sending MSG out")
				_, err = connection.Write(msgOut)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error Sending %s", err.Error())
					// create more descriptive msg
					// send msg up to indicate a problem ?
				}

			case ctrlMsg := <-sendControlChannel: //
				_ = tbJsonUtils.TBunmarshal(ctrlMsg, &controlMsg)
				fmt.Println(masterName, "SendThread got control MSG=", controlMsg)

				if strings.Contains(controlMsg.MsgType, "TERMINATE") {
					fmt.Println(masterName, "SendThread rcvd control MSG=", controlMsg)
					return
				}
			}
		}

	}()

	return err
}

//=======================================================================
// masterRecvThread() - Thread receiving messages from others
//=======================================================================
func masterRecvThread(conn *net.UDPConn, recvControlChannel <-chan []byte) error {
	var err error = nil
	fmt.Println(masterName, "RecvThread: Start RECV THRED")
	go func() {
		connection := conn

		fmt.Println(masterName, "RecvThread: Start Receiving")
		var controlMsg tbMessages.TBmessage
		var oobBuffer [3000]byte

		for {
			recvBuffer := make([]byte, 3000)
			length, oobn, flags, addr, err := connection.ReadMsgUDP(recvBuffer[0:], oobBuffer[0:])

			masterReceiveCount++

			fmt.Println(masterName, "\n=============== Count=", masterReceiveCount,
				"\nRecv from", addr, "len=", length, "oobLen=", oobn, "flags=", flags, "ERR=", err)

			// Just add the received message to the RECV queue.
			// processing is done inside the main()
			masterRecvChannel <- recvBuffer[0:length]

			if len(recvControlChannel) != 0 {
				ctrlMsg := <-recvControlChannel
				_ = tbJsonUtils.TBunmarshal(ctrlMsg, &controlMsg)
				fmt.Println(masterName, "RecvThread got CONTROL MSG=", controlMsg)
				if strings.Contains(controlMsg.MsgType, "TERMINATE") {
					fmt.Println(masterName, "RecvThread rcvd control MSG=", controlMsg)
					return
				}
			}
		}
	}()

	return err
}

//=======================================================================
// A message was found in the receive queue. Handle it depending on the
// master state (Init, Connecting, Connected, UP, DOWN)
//=======================================================================
func masterHandleMessages(message []byte) {
	//fmt.Println(masterName, "HandleMessages: Recv Message in State", masterState, "Message=",string(message))
	msg := new(tbMessages.TBmessage)
	_ = tbJsonUtils.TBunmarshal(message, &msg)
	//fmt.Println(masterName, "HandleMessages: Recv Message in State", masterState, "Type:",msg.MsgType,"From=",msg.MsgSender)

	switch masterState {
	case MasterInit:
		break
	case MasterConnecting:
		// stateConnectingMessages(msg)
		break
	case MasterConnected:
		// fmt.Println(masterName,"State=",masterState," Send MSG to=", receiver, "Type:", msg.MsgType)
		stateUpMessages(message)
		break
	case MasterUP:
		break
	case MasterDOWN:
		// Handle the down state later properly
		stateUpMessages(message)
		break
	default:
	}
}

//=======================================================================
//
//=======================================================================
func handleTimerMessages(message string) {
	fmt.Println(masterName, "MAIN: Timer tick in state", masterState, "MSG=", message)
	// fmt.Println(masterName,"HandleTimerMessages MSG(",unsafe.Sizeof(message),")=", string(message))
	// switch on masterState
	switch masterState {
	case MasterInit:
	case MasterConnecting:
		officeMasterInit()
		break
	case MasterConnected:
	case MasterUP:
	case MasterDOWN:
		fmt.Println("Timer Tick")
		break
	default:
	}
}

//insert masterUdpAddress into message insted of id Uint ..........
//    to be used for replies
//=======================================================================
//
//=======================================================================
func stateUpMessages(message []byte) {

	// Unmarshal
	msg := new(tbMessages.TBmessage)
	_ = tbJsonUtils.TBunmarshal(message, &msg)

	switch msg.MsgType {
	case tbMessages.MSG_TYPE_REGISTER:
		receiveRegisterMsg(msg)
		break
	case tbMessages.MSG_TYPE_HELLO:
		masterSendHelloReplyMsg(msg)
		break
	case tbMessages.MSG_TYPE_CMD:
		// masterCommandMsg(msg)
		break
	case tbMessages.MSG_TYPE_CMD_REPLY:
		// masterCmdReplyMsg(msg)
		break
	case tbMessages.MSG_TYPE_SAT_STATUS:
		// masterSatStatusMsg(msg)
		break
	default:
		break
	}
	//currentTime := strconv.FormatInt(tbMsgUtils.TBtimestamp(), 10)
	//replyMessage := "stateUpMessages:Replying to You at " + currentTime + "ms"
	//fmt.Println("REPLY to ", )
	//fmt.Println(masterName, "stateUpMessages: REPLY=", replyMessage)
	//sendBuffer, _ := tbJsonUtils.TBmarshal(replyMessage)
	//masterConnection.WriteToUDP(sendBuffer, masterUdpAddress)

}

//=======================================================================
//
//=======================================================================
func officeMgrSetState(newState string) {
	fmt.Println(masterName, "OldState=", masterState, " NewState=", newState)
	masterState = newState
}

//=======================================================================
//
//=======================================================================
func masterCheckError(err error) {
	if err != nil {
		_, _ = fmt.Println(masterName, "Fatal error ", err)
		os.Exit(1)
	}
}

//=======================================================================
//
//=======================================================================
func masterSendHelloReplyMsg(msg *tbMessages.TBmessage) {
	remoteUdpAddress := net.UDPAddr{IP: msg.MsgSender.Address.IP,
		Port: msg.MsgSender.Address.Port}

	replyBuffer := tbMsgUtils.TBhelloReplyMsg(masterFullName, msg.MsgSender, string(msg.MsgBody))

	// fmt.Println(masterName,"WriteToUdp Reply remoteUdpAddress=", remoteUdpAddress)
	_, _ = masterConnection.WriteToUDP(replyBuffer, &remoteUdpAddress)
}

//=======================================================================
//
//=======================================================================
func receiveRegisterMsg(msg *tbMessages.TBmessage) {
	fmt.Println("REGISTER MSG FROM: ", msg.MsgSender)

	// Unmarshal the message body
	var newSatellite tbMessages.TBmgr
	_ = tbJsonUtils.TBunmarshal(msg.MsgBody, &newSatellite)

	fmt.Println(masterName, "SAT:", newSatellite.Name.Name, "STATUS:", newSatellite.Up, "ADDRESS:", newSatellite.Name.Address,
		"CREATED:", newSatellite.Name.TimeCreated, "MSGSRCVD:", newSatellite.MsgsRcvd)

	// in sliceOfSatellites, check if new one already there if yes update, otherwise append
	knownSatellite := MasterLocateSatellite(sliceOfSatellites, newSatellite.Name.Name)
	if knownSatellite != nil { // Update existing mgr/master record
		fmt.Println(masterName, "UPDATE in sliceOfSatellites MGR=", knownSatellite.Name)
		*knownSatellite = newSatellite
	} else { // Add a new satellite
		fmt.Println(masterName, "STORE in sliceOfSatellites MGR=", newSatellite.Name)
		sliceOfSatellites = append(sliceOfSatellites, newSatellite)
	}
	fmt.Println(masterName, "New sliceOfSatellites=", sliceOfSatellites)
	fmt.Println(masterName, "LENGTH of sliceOfSatellites=", len(sliceOfSatellites))
}

//============================================================================
// Locate specific row in the slice of all satellites, containing rows for
// all known satellites. Return nil if row not found
//============================================================================
func MasterLocateSatellite(sliceTable []tbMessages.TBmgr, satellite string) *tbMessages.TBmgr {
	for index := range sliceTable {
		if sliceTable[index].Name.Name == satellite {
			return &sliceTable[index]
		}
	}

	return nil
}

//====================================================================================
// send keep alive messages to everybody registered
//====================================================================================
func sendKeepAliveMsg() {
	// GS also send only to modules that are up ....
	//{TB-EXPMASTER 1 1522878314281123 {172.18.0.3 1200 }} true 1522878314281123 0 1522880258356310 0 0}}
	// TO DO a lots of cleanup and better logic btw this, registration and peiodic timer

	var names = ""
	if len(sliceOfSatellites) > 0 {
		fmt.Println(masterName, "sendKeepAlive: LENGTH of sliceOfSatellites=", len(sliceOfSatellites))
		msgBody, _ := tbJsonUtils.TBmarshal(sliceOfSatellites)
		fmt.Println(masterName, "sendKeepAlive: LENGTH of msgBody=", len(msgBody))
		for mgrIndex := range sliceOfSatellites {
			receiver := sliceOfSatellites[mgrIndex].Name
			if receiver.Name != masterName { // Do not send to self
				udpAddress := sliceOfSatellites[mgrIndex].Name.Address
				newMsg := tbMsgUtils.TBkeepAliveMsg(masterFullName, receiver, string(msgBody))
				tbMsgUtils.TBsendMsgOut(newMsg, udpAddress, masterConnection)
				names += " " + receiver.Name
			}
		}
		fmt.Println(masterName, ": KNOWN Satellites=", names)
	}
}

func sendCommandListMsg() {
	//msgBody, _ := tbJsonUtils.TBmarshal(commandList)
}

//====================================================================================
// send routing update commands to all sattelites
//====================================================================================
func sendRoutingUpdate() {
	var names = ""

	for mgrIndex := range sliceOfSatellites {
		receiver := sliceOfSatellites[mgrIndex].Name
		if receiver.Name != masterName { // Do not send to self
			udpAddress := sliceOfSatellites[mgrIndex].Name.Address
			// for a position
			// for each sattelite
			msgBody, _ := tbJsonUtils.TBmarshal(commandList)
			newMsg := tbMsgUtils.TBkeepAliveMsg(masterFullName, receiver, string(msgBody))
			tbMsgUtils.TBsendMsgOut(newMsg, udpAddress, masterConnection)
			names += " " + receiver.Name
		}
	}
	fmt.Println(masterName, ": Routing sent to Satellites=", names)
}
