//=============================================================================
// FILE NAME: tbMsgUtils.go
// DESCRIPTION:
// Utilities for creating and handling messages used by the
// Experiment Master and Experiment Controllers.
// Contains description of all possible messages related to Experiments
//
// NAME              REV  DATE       REMARKS			@
// Goran Scuric      1.0  01012019  Initial design
//================================================================================

package tbMsgUtils

import (
	"bifrost/common/tbJsonUtils"
	"bifrost/common/tbMessages"
	"fmt"
	"net"
	"strconv"
	"time"
)

//============================================================================
// Create a Hello message (to send to office mgr)
//============================================================================
func TBtimestamp() int64 {
	return time.Now().UnixNano() / 1000 // / (int64(time.Millisecond)/int64(time.Nanosecond))
}

//====================================================================================
//
//====================================================================================
func TBsendMsgOut(msgOut []byte, udpAddress net.UDPAddr, udpConnection *net.UDPConn) {

	if udpConnection != nil { // returns numBytes, err
		_, _ = udpConnection.WriteToUDP(msgOut, &udpAddress)
	} else {
		fmt.Println("ERROR Sending Out to", udpAddress, ": udpConnection = nil")
	}
}

//=============================================================================
// Function:    expMessage
// Description:
//         Create marshaled message ready to be sent out.
// Input:  sender,receiver    = name of the sender and receiver
//         mclass,mtype,mbody = message class and type, message body string
// Output: msg = marshalled message
// Error Conditions:
//      None [or state condition for each possible error]
//=============================================================================
func TBmarschalMessage(sender, receiver tbMessages.NameId, mtype, mbody string) []byte {
	// length := len(mbody)
	currentTime := strconv.FormatInt(TBtimestamp(), 10)
	msgBytes := []byte(mbody)
	myMsg := tbMessages.TBmessage{
		MsgSender:   sender,
		MsgReceiver: receiver,
		MsgType:     mtype,
		TimeSent:    currentTime,
		MsgBody:     msgBytes,
	}

	msg, _ := tbJsonUtils.TBmarshal(myMsg)
	return msg
}

//============================================================================
// Create a Hello message (to send to office mgr)
//============================================================================
func TBConnectedMsg(sender, receiver tbMessages.NameId, mBody string) []byte {
	//fmt.Println("MsgUtils: create CONNECTED msg")
	msg := TBmarschalMessage(sender, receiver, tbMessages.MSG_TYPE_CONNECTED, mBody)
	return msg
}

//============================================================================
// Create a Hello message (to send to office mgr)
//============================================================================
func TBkeepAliveMsg(sender, receiver tbMessages.NameId, mBody string) []byte {
	//fmt.Println("MsgUtils: create KEEP ALIVE msg, mBody=", mBody)
	msg := TBmarschalMessage(sender, receiver, tbMessages.MSG_TYPE_KEEPALIVE, mBody)
	return msg
}

//============================================================================
// Create a Hello message (to send to office mgr)
//============================================================================
func BiRouteUpdateMsg(sender, receiver tbMessages.NameId, mBody string) []byte {
	// fmt.Println("MsgUtils: create Route Update for ", receiver, " mBody=", mBody)
	msg := TBmarschalMessage(sender, receiver, tbMessages.MSG_TYPE_CMD, mBody)
	return msg
}

//============================================================================
// Create a special command message
//============================================================================
func BiControlMsg(sender, receiver tbMessages.NameId, mBody string) []byte {
	fmt.Println("MsgUtils: create COMMANDS msg for ", receiver.Name, "  mBody=", mBody)
	replyBuffer := TBmarschalMessage(sender, receiver, tbMessages.MSG_TYPE_TERMINATE, mBody)
	return replyBuffer
}

//============================================================================
// Create a Hello message (to send to office mgr)
//============================================================================

// INSTEAD OF  receiver provide *msg.MsgReceiver
// AND extract all of receivers fields

func TBhelloMsg(sender, receiver tbMessages.NameId, mBody string) []byte {
	//fmt.Println("MsgUtils: create HELLO msg, mBody=", mBody)
	msg := TBmarschalMessage(sender, receiver, tbMessages.MSG_TYPE_HELLO, mBody)
	return msg
}

//============================================================================
// Create a REGISTER message (to send to office mgr)
//============================================================================
func TBregisterMsg(sender, receiver tbMessages.NameId, mBody string) []byte {
	//fmt.Println("MsgUtils: create REGISTER msg, mBody=",mBody)
	msg := TBmarschalMessage(sender, receiver, tbMessages.MSG_TYPE_REGISTER, mBody)
	return msg
}

//============================================================================
// Create a Hello message (to send to office mgr)
//============================================================================
func TBhelloReplyMsg(sender, receiver tbMessages.NameId, mBody string) []byte {
	//fmt.Println("MsgUtils: create HELLO REPPLY msg, mBody=", mBody)
	msg := TBmarschalMessage(sender, receiver, tbMessages.MSG_TYPE_HELLO_REPPLY, mBody)
	return msg

}

//====================================================================================
//
//====================================================================================
func sendHelloReplyMsg(senderName tbMessages.NameId, msg *tbMessages.TBmessage) {
	//	receiver := msg.MsgSender
	//	newMsg := TBhelloMsg(senderName, receiver, senderName.Name + " is alive")
	//	TBsendMsgOut(newMsg, receiver.Address, myConnection)
}

//============================================================================
// Create a Hello message (to send to office mgr)
//============================================================================
func TBmsgSample(myName, receiver tbMessages.NameId, mBody string) []byte {

	msgBytes := []byte(mBody)

	myMsg := tbMessages.TBmessage{
		MsgSender:   myName,
		MsgReceiver: receiver,
		// MsgClass:    tbMessages.MSG_CLASS_CONTROL,
		MsgType: tbMessages.MSG_TYPE_INIT,
		MsgBody: msgBytes,
	}

	msg, _ := tbJsonUtils.TBmarshal(myMsg)
	// fmt.Println("JSON=", string(msg))
	return msg
}
