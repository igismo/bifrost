
//*********************************************************************************/
// NAME              REV  DATE       REMARKS			@
// Goran Scuric      1.0  10222019  Initial design
//================================================================================
package tbUtils

import (
	"bifrost/common/tbMessages"
	"net"
	"bifrost/common/tbConfiguration"
	"fmt"
	"bifrost/common/tbJsonUtils"
	"strconv"
	"bifrost/common/tbMsgUtils"
)

//============================================================================
// Locate specific master module in the slice of all masters (containing rows for
// all known masters including ourselves and the office master). The slice is
// populated from the KeepAlive msg received from officeMaster
// Return nil if row not found
//============================================================================
func LocateMaster(slice [] tbMessages.TBmgr, NameOfMaster string) (*tbMessages.TBmgr, int){
	for index := range slice {
		if  slice[index].Name.Name == NameOfMaster {
			return &slice[index], index
		}
	}
	// requested master not found in the slice of master
	return nil, -1
}
//====================================================================================
//
//====================================================================================
func LocateOfficeMaster(myName string, sliceOfMgrs [] tbMessages.TBmgr) (*net.UDPAddr,
						tbMessages.NameId, bool) {
	// NOTE that this will fail if officeMasterUdpAddress has not been initialized due
	// to Office Manager not being up . Add state and try again to Resolve before
	// doing this
	var err error
	var UdpAddress	= new(net.UDPAddr)
	var FullName      tbMessages.NameId

	// fmt.Println(myName, ": Locate Office Manager")
	UdpAddress, err = net.ResolveUDPAddr("udp", tbConfig.BifrostMaster)
	if err != nil {
		fmt.Println(myName,": ERROR locating Office Manager, will retry", err)
		return nil, FullName, false
	}

	FullName = tbMessages.NameId{Name: tbConfig.BifrostMasterURL, OsId: 0,
		TimeCreated: "0", Address: *UdpAddress}

	// Now that we know that officeMaster is alive add entry to our slice of masters
	officeMaster := tbMessages.TBmgr{Name: FullName, Up:true, LastChangeTime:"0",
		MsgsSent: 0, LastSentAt: "0", MsgsRcvd:0, LastRcvdAt:"0"}

	sliceOfMgrs = append(sliceOfMgrs, officeMaster)

	theMgr, _ :=  LocateMaster(sliceOfMgrs, tbConfig.BifrostMasterURL)
	if theMgr != nil {
		fmt.Println(myName, ": MGR=",theMgr.Name.Name, "ADDRESS:",theMgr.Name.Address,
			"CREATED:",theMgr.Name.TimeCreated, "MSGSRCVD:",theMgr.MsgsRcvd)
	}
	return UdpAddress, FullName, true
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
// Send registration msg to officeMaster containing our full row from our slice of
// masters.
//====================================================================================
func SendRegisterMsg(senderFullName tbMessages.NameId,
			sliceOfMasters [] tbMessages.TBmgr, myConnection *net.UDPConn ) {
	// Locate our own record in the slice of masters
	me, _ := LocateMaster(sliceOfMasters, senderFullName.Name)

	if me != nil {
		officeUdpAddress, rcvrFullName, result :=
			LocateOfficeMaster(senderFullName.Name, sliceOfMasters )
		if result == false {
			fmt.Println(senderFullName.Name, ": Failed to Send Register Msg")
			return
		}
		// Marshall our own row from the slice
		msgBody, _ := tbJsonUtils.TBmarshal(me)
		me.LastSentAt = strconv.FormatInt(tbMsgUtils.TBtimestamp(),10)
		newMsg := tbMsgUtils.TBregisterMsg(senderFullName, rcvrFullName, string(msgBody))
		// fmt.Println(myName, "stateConnected REGISTER with officeMaster ")
		tbMsgUtils.TBsendMsgOut(newMsg, *officeUdpAddress, myConnection)
	} else {
		fmt.Println(senderFullName.Name, ": FAILED to locate mngr record in the slice")
	}
}