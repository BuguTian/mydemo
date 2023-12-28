package main

import (
	"chatroom/net"
	"chatroom/util"
	"log"
	"gopkg.in/ini.v1"
	"reflect"
	"os"
	"os/exec"
	"fmt"
	"flag"
)

type ConfigItem struct {
	MyName        	string  `default:"chatroom"`
	LogPath       	string  `default:"./log"`
	LogLevel      	string  `default:"ERROR"`
	ListenPort    	int 	`default:8688`
	YnbsLogAddr   	string 	`data`
}

var g_Config ConfigItem

func SetConfig(key string, value string) {
	if (value == ""){
		return
	}
	pp := reflect.ValueOf(&g_Config)
	field := pp.Elem().FieldByName(key)
	field.SetString(value)
}

func InitLog() {
	err1 := os.MkdirAll(g_Config.LogPath, os.ModePerm)
	if err1 != nil {
		panic(err1)
	}
	file := g_Config.LogPath + "/run.log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // 将文件设置为log输出的文件
	log.SetPrefix( "[" + g_Config.MyName + "] " )
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

func InitConfig() {
	g_Config.MyName     = "chatroom"
	g_Config.LogPath    = "./log" 
	g_Config.LogLevel   = "ERROR"
	g_Config.ListenPort = 8688
  
	cfg, err := ini.Load("config.ini")
	if err != nil {
	  log.Fatal("Fail to read file: ", err)
	}

	tmpstr := cfg.Section("websocket").Key("listen_port").String()
	if (len(tmpstr) > 0){
	  g_Config.ListenPort = util.Str2int(tmpstr)
	}
  
	SetConfig("YnbsLogAddr",cfg.Section("websocket").Key("ynbs_log_addr").String())
	SetConfig("MyName",cfg.Section("").Key("my_name").String())
	SetConfig("LogLevel",cfg.Section("").Key("log_level").String())
	SetConfig("LogPath",cfg.Section("").Key("log_path").String())
}

const (
	DAEMON = "daemon"
	FOREVER = "forever"
	VERARG = "version"
)
  
func RunServer() {
	InitConfig()
	InitLog()
	util.EnableLog = true
	s := net.Server{}
	s.Listen(g_Config.ListenPort)
}

func StripSlice(slice []string, element string) []string {
	for i := 0; i < len(slice); {
		if slice[i] == element && i != len(slice)-1 {
			slice = append(slice[:i], slice[i+1:]...)
		} else if slice[i] == element && i == len(slice)-1 {
			slice = slice[:i]
		} else {
			i++
		}
	}
	return slice
}

func SubProcess(args []string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Println(os.Stderr, "[-] Error: %s\n", err)
	}
	return cmd
}

func main(){
	daemon := flag.Bool(DAEMON, false, "run in daemon")
	forever := flag.Bool(FOREVER, false, "run forever")
	version := flag.Bool(VERARG, false, "get version")
	flag.Parse()

	if *version {
		fmt.Println("Project: chatroom")
		fmt.Println("Version:", VERSION)
		fmt.Println("GitRev:", GITREV)
		fmt.Println("BuildTime:", BUILDTIME)
		return 
	}

	fmt.Printf("[*] PID: %d PPID: %d ARG: %s\n", os.Getpid(), os.Getppid(), os.Args)
	if *daemon {
		SubProcess(StripSlice(os.Args, "-"+DAEMON))
		log.Printf("[*] Daemon running in PID: %d PPID: %d\n", os.Getpid(), os.Getppid())
		os.Exit(0)
	} else if *forever {
		for {
			cmd := SubProcess(StripSlice(os.Args, "-"+FOREVER))
			log.Printf("[*] Forever running in PID: %d PPID: %d\n", os.Getpid(), os.Getppid())
			cmd.Wait()
		}
		os.Exit(0)
	} else {
		log.Printf("[*] Service running in PID: %d PPID: %d\n", os.Getpid(), os.Getppid())
	}
	RunServer()
}
  