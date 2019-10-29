
//=============================================================================
// FILE NAME: tbConfig.go
// DESCRIPTION: configuration, IP addresses, ...
// Notice that Docker provides resolution within created networks
// By defaulkt we create "TB-NETWORK" network where all testbed modules live.
// Each docker testbed module will have a preassigned name like "biMaster",
// which will be translated by the dockers DNS into its IP address
//
// NAME              REV  DATE       REMARKS			@
// Goran Scuric      1.0  01012018  Initial design     goran@usa.net
//================================================================================

package tbConfig

import (
	"time"
)

//# Configure variables
var TBDIR     		   	= "/Users/scuric/go/src/bifrost"
var TBlogPath           = TBDIR + "/log/"
var TBversion           = time.Now()

//strconv.FormatInt(tbMsgUtils.TBtimestamp(),10)

var BifrostPort       = "1200"




