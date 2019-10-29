//*********************************************************************************/
// NAME              REV  DATE       REMARKS			@
// Goran Scuric      1.0  01012018  Initial design     goran@usa.net
//================================================================================
/* Various net utilities */

package tbNetUtils

import ( 
    "fmt"
    "net"
	"strconv"
	"strings"
)


//================================================
// GetLocalIp() - return non-loopback IP address
//================================================
func GetLocalIp() (string) {

    netInterfaceAddresses, err := net.InterfaceAddrs()

    if err == nil { // no error 
	    for _, netInterfaceAddress := range netInterfaceAddresses {

			networkIp, ok := netInterfaceAddress.(*net.IPNet)
			//fmt.Println("GetIP: ", networkIp.IP.String)
        	if ok && !networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil {

				ip := networkIp.IP.String()
					fmt.Println("NetUtil: ", "Resolved Host IP: " + ip)
						return ip
            }
        }
    }

    return ""
}

//====================================================================================
//
//====================================================================================
func GetMastersIP(master string) string {
	a := [] string{""}
	for i := 1; i < 10; i++ {
		currIp := "172.18.0." + strconv.Itoa(i)
		// fmt.Println("i=", i, "  LOOKUP ", currIp)

		a, _ = net.LookupAddr(currIp) //names []string, err error)
		if a != nil && len(a) > 0 {
			// fmt.Println("a[0]=", a[0])
			if strings.Contains(a[0], master) {
				// FOUND
				return currIp
				break;
			}
		}
	}
	return ""
}
//====================================================================================
//
//====================================================================================
func myfindUdpAddress(service string) {
	udpAddr, _ := net.ResolveUDPAddr("udp", service)

	fmt.Println("Bifrost found at ", udpAddr)
}