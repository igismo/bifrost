
//*********************************************************************************/
// NAME              REV  DATE       REMARKS			@
// Goran Scuric      1.0  01012018  Initial design     goran@usa.net
//================================================================================
/* SaveJSON
 */
package tbJsonUtils

import ( 
    "encoding/json"
    "fmt"
	"os"
)


func TBmarshal(key interface{}) ([]byte, error) {
    msg, err := json.Marshal(key)
    return msg,err
}

func TBunmarshal(input []byte, key interface{}) (error) {
	err := json.Unmarshal(input, &key)
	return err
}

func TBmarshalAndSave(filename string,key interface{}) ([]byte, error) {
	// Marshall into file
	if filename != "" {
		outFile, err := os.Create(filename)
		checkError(err)
		encoder := json.NewEncoder(outFile)
		err = encoder.Encode(key)
		outFile.Close()
	}
	// Marshall into msg
	msg, err := json.Marshal(key)
	return msg,err
}

func TBloadAndUnmarshal(fileName string, key interface{}) {
	inFile, err := os.Open(fileName)
	checkError(err)
	decoder := json.NewDecoder(inFile)
	err = decoder.Decode(key)
	checkError(err)
	inFile.Close()
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
