
//*********************************************************************************/
// NAME              REV  DATE       REMARKS			@
// Goran Scuric      1.0  10222019  Initial design
//================================================================================

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)
//=======================================================================
// Early work to test some options for officeMaster console
//=======================================================================
func startConsole(consoleInput <-chan string) {

	go func(ch <-chan string) {
		reader := bufio.NewReader(os.Stdin)
		for {
			swapinCommand  := flag.NewFlagSet("swapin", flag.ContinueOnError)
			projectFlag    := swapinCommand.String("p", "",      "project")
			experimentFlag := swapinCommand.String("e", "",      "experiment")
			userFlag       := swapinCommand.String("u", "goran", "user")

			swapoutCommand    := flag.NewFlagSet("swapout", flag.ContinueOnError)
			projectFlagOut    := swapoutCommand.String("p", "",      "-p <project>")
			experimentFlagOut := swapoutCommand.String("e", "",      "-e <experiment>")
			userFlagOut       := swapoutCommand.String("u", "goran", "user")

			commandOK := ""
			s, err := reader.ReadString('\n')
			if err != nil { // Maybe log non io.EOF errors, if you want
				fmt.Println("Console:Error during ReadString")
			}

			//ch <- s

			s1 := strings.Split(s, "\n")
			sa := strings.Split(s1[0], " ")

			//for i := 0; i < length; i++ {println("START SA[", i, "]=", sa[i])}

			switch sa[0] {
			case "quit":
				fmt.Printf("Exiting\n")
				os.Exit(0)
			case "exit":
				fmt.Printf("Exiting\n")
				os.Exit(0)
			case "help":
				fmt.Printf("No HELP available yet\n")
				break
			case "swapin":
				var _ = swapinCommand.Parse(sa[1:])
				break
			case "stop":
				RotationEnabled = false
				break
			case "start":
				if sa[1] != "" {
					RotationPeriod, _ = strconv.ParseFloat(sa[1], 64)
				}
				RotationEnabled = true
				break
			case "terminate":
				if sa[1] != "" {
					sat := MasterLocateSatellite(sliceOfSatellites, sa[1])
					if sat != nil {
						sat.Name.Terminate = true
					}
				}
				break
			default:
				flag.PrintDefaults()
				fmt.Printf("%q is not valid command.\n", sa[0])
				commandOK = "Error Bad Command"
				break
			}

			if commandOK == "" {
				if swapinCommand.Parsed() {
					if *projectFlag == "" {
						fmt.Println("SwapIn: Please supply the project name")
					} else {fmt.Printf("SwapIn p=%q", *projectFlag)}

					if *experimentFlag == "" {
						fmt.Println("SwapIn: Please supply experiment name")
					} else {fmt.Printf("SwapIn e=%q", *experimentFlag)}
					if *userFlag == "" {
						fmt.Println("SwapIn: Please supply the user name.")
					} else {fmt.Printf("Swapin u=%q", *userFlag)}
					sendSwapInMsg(*projectFlag, *experimentFlag,*userFlag, "")
				}
				if swapoutCommand.Parsed() {
					if *projectFlagOut == "" {
						fmt.Println("SwapOut: Please supply the project name")
					} else {fmt.Printf("\nSwapOut p=%q", *projectFlagOut)}
					if *experimentFlagOut == "" {
						fmt.Println("SwapOut: Please supply the experiment name")
					} else {fmt.Printf("\nSwapOut e=%q", *experimentFlagOut)}

					fmt.Printf("\nSwapOut u=%q", *userFlagOut)
					sendSwapOutMsg(*projectFlagOut, *experimentFlagOut,*userFlagOut, "")
				}
			}
		} // end of for ever
		//close(ch)
	}(consoleInput)
}
//=======================================================================
//
//=======================================================================
func sendSwapInMsg(project, experiment,userName, fileName string) {
//	expMgrName := tbConfig.BifrostMasterURL
//	expMgrFullName := TBmasterLocateMngr(masterSliceOfMgrs, expMgrName)
//	if expMgrFullName != nil && expMgrFullName.Up == true {
//		swapIn := tbMessages.SwapIn{Project:project, Experiment:experiment,
//						UserName:userName, FileName: fileName}
//		messageBody, _ := tbJsonUtils.TBmarshal(swapIn)
//		newMsg := tbMsgUtils.TBswapinMsg(masterFullName, expMgrFullName.Name, string(messageBody))
//
//		tbMsgUtils.TBsendMsgOut(newMsg, expMgrFullName.Name.Address, masterConnection)
//	} else {
//		fmt.Println("Console: Exp Master not available - try later")
//	}
}
//=======================================================================
//
//=======================================================================
func sendSwapOutMsg(project, experiment,userName, fileName string) {
//	expMgrName := tbConfig.BifrostMasterURL
//	expMgrFullName := TBmasterLocateMngr(masterSliceOfMgrs, expMgrName)
//	if expMgrFullName != nil && expMgrFullName.Up == true {
//		swapOut := tbMessages.SwapOut{Project:project, Experiment:experiment,
//			UserName:userName, FileName: fileName}
//		messageBody, _ := tbJsonUtils.TBmarshal(swapOut)
//		newMsg := tbMsgUtils.TBswapoutMsg(masterFullName, expMgrFullName.Name, string(messageBody))
//
//		tbMsgUtils.TBsendMsgOut(newMsg, expMgrFullName.Name.Address, masterConnection)
//	} else {
//		fmt.Println("Console: Exp Master not available - try later")
//	}
}