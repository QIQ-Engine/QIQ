package main

import (
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/config"
	"QIQ/cmd/qiq/ini"
	"QIQ/cmd/qiq/interpreter"
	"QIQ/cmd/qiq/request"
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/stats"
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var serverAddr string
var documentRoot string
var iniFilePath string

func main() {
	iniFilePathFlag := flag.String("ini", "", "Path to ini file to use.")
	createIniFile := flag.String("create-ini", "", "Create given ini file.")
	file := flag.String("f", "", "Parse and execute <file>.")
	isDev := flag.Bool("dev", false, "Run in developer mode.")
	// Developer tools
	showStats := flag.Bool("stats", false, "Show statistics.")
	debugMode := flag.Bool("debug", false, "Enable debug mode.")
	// Web server
	addr := flag.String("S", "", "Run with built-in web server. <addr>:<port>")
	docRoot := flag.String("t", "", "Specify document root <docroot> for built-in web server.")

	flag.Parse()

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("QIQ %s (PHP %s)\n", config.QIQVersion, config.Version)
		os.Exit(0)
	}

	if *createIniFile != "" {
		iniWriter := ini.NewWriter(*createIniFile)
		if err := iniWriter.Write(); err != nil {
			fmt.Printf(`Creation of ini file "%s" failed: %s`, *createIniFile, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	iniFilePath = *iniFilePathFlag
	config.IsDevMode = *isDev
	config.ShowStats = *showStats
	if *debugMode {
		config.ShowParserCallStack = true
		config.ShowInterpreterCallStack = true
	}

	// Serve with built-in web server
	if *addr != "" {
		serverAddr = *addr
		documentRoot = *docRoot
		webServer()
		os.Exit(0)
	}

	// Parse given file
	if *file != "" {
		absFilePath, err := common.GetAbsPath(*file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		processFile(absFilePath)
	}

	// Read stdin or wait for it
	processStdin()
}

// -------------------------------------- stdin -------------------------------------- MARK: stdin

func processStdin() {
	content := ""
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if content != "" || scanner.Text() == "" {
			content += "\n"
		}
		content += scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error:", err)
	}

	output, exitCode := processContent(nil, nil, string(content), "")
	fmt.Print(output)
	os.Exit(exitCode)
}

// -------------------------------------- given file -------------------------------------- MARK: given file

func processFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	output, exitCode := processContent(nil, nil, string(content), filename)
	fmt.Print(output)
	if exitCode == 500 {
		exitCode = 1
	}
	os.Exit(exitCode)
}

// -------------------------------------- web server -------------------------------------- MARK: web server

func webServer() {
	if documentRoot == "" {
		documentRoot, _ = os.Getwd()
	} else {
		absPath, err := common.GetAbsPath(documentRoot)
		if err != nil {
			fmt.Println("Error: Could not find directory " + documentRoot)
			os.Exit(1)
		}
		documentRoot = absPath
	}

	var mode string
	if config.IsDevMode {
		mode = "Development"
	} else {
		mode = "Production"
	}

	fmt.Printf("[%s] QIQ %s (PHP %s) %s Server (%s) started\n",
		time.Now().Format("Mon Jan 02 15:04:05 2006"), config.QIQVersion, config.Version, mode, serverAddr,
	)
	fmt.Println("Document root is " + documentRoot)
	fmt.Println("Press Ctrl-C to quit")
	fmt.Println("")

	http.HandleFunc("/", requestHandler)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		fmt.Println("Failed to start web server")
		fmt.Println(err)
	}
}

func getNotFoundText(resource string) string {
	return `<!doctype html><html><head><title>404 Not Found</title><style>
			body { background-color: #fcfcfc; color: #333333; margin: 0; padding:0; }
			h1 { font-size: 1.5em; font-weight: normal; background-color: #42b5fc; min-height:2em; line-height:2em; border-bottom: 1px inset black; margin: 0; }
			h1, p { padding-left: 10px; }
			code.url { background-color: #eeeeee; font-family:monospace; padding:0 2px;}
			</style>
			</head><body><h1>Not Found</h1><p>The requested resource <code class="url">` + resource + `</code> was not found on this server.</p></body></html>
		`
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	absFilePath := filepath.Join(documentRoot, r.URL.Path)
	_, err := os.Stat(absFilePath)
	if err != nil {
		fmt.Println("404", absFilePath)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, getNotFoundText(r.URL.Path))
		return
	}

	if common.IsDir(absFilePath) {
		index := filepath.Join(absFilePath, "index.html")
		if common.PathExists(index) && !common.IsDir(index) {
			absFilePath = index
		} else if index = filepath.Join(absFilePath, "index.php"); common.PathExists(index) && !common.IsDir(index) {
			absFilePath = index
		} else {
			fmt.Println("404", absFilePath)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, getNotFoundText(r.URL.Path))
			return
		}
	}

	if !strings.HasSuffix(absFilePath, ".php") {
		http.ServeFile(w, r, absFilePath)
		return
	}

	content, err := os.ReadFile(absFilePath)
	if err != nil {
		fmt.Println("Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	output, exitCode := processContent(w, r, string(content), absFilePath)
	if exitCode == 500 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("%d %s\n", 500, absFilePath)
	} else {
		fmt.Printf("%d %s\n", 200, absFilePath)
	}
	fmt.Fprint(w, output)
}

// -------------------------------------- core logic -------------------------------------- MARK: core logic

func processContent(w http.ResponseWriter, r *http.Request, content string, filename string) (output string, exitCode int) {
	stat := stats.Start()
	defer stats.StopAndPrint(stat, "Total")

	var initIni *ini.Ini
	if config.IsDevMode {
		initIni = ini.NewDevIni()
	} else {
		initIni = ini.NewDefaultIni()
	}

	iniReader := ini.NewReader()

	// Read ini in folder of this binary
	iniFile := common.Join(common.GetExecutablePath(), "qiq.ini")
	if common.PathExists(iniFile) {
		if err := iniReader.Read(iniFile); err != nil {
			return fmt.Sprintf("Cannot read ini file: %s", iniFile), 1
		}
		initIni = iniReader.GetIni()
	}

	// Read ini in document root
	iniFile = common.Join(documentRoot, "qiq.ini")
	if common.PathExists(iniFile) {
		if err := iniReader.Read(iniFile); err != nil {
			return fmt.Sprintf("Cannot read ini file: %s", iniFile), 1
		}
		initIni = iniReader.GetIni()
	}

	// Read given ini file
	iniFile = common.GetAbsPathForWorkingDir(documentRoot, iniFilePath)
	if common.PathExists(iniFile) {
		if err := iniReader.Read(iniFile); err != nil {
			return fmt.Sprintf("Cannot read ini file: %s", iniFile), 1
		}
		initIni = iniReader.GetIni()
	}

	if config.IsDevMode {
		initIni.SetToDev()
	}

	request := request.NewRequestFromGoRequest(r, documentRoot, serverAddr, filename)
	interpreter, err := interpreter.NewInterpreter(runtime.NewExecutionContext(), initIni, request, filename)

	if w != nil {
		// TODO content-type returned from interpreter?
		w.Header().Set("Content-Type", "text/html")
		if initIni.GetBool("expose_php") {
			w.Header().Add("X-Powered-By", config.SoftwareVersion)
		}
	}

	if err != nil {
		return interpreter.ErrorToString(err), 500
	}
	result, err := interpreter.Process(content)
	if err != nil {
		result += "\n" + interpreter.ErrorToString(err)
		return result, 500
	}
	if err := common.DeleteFiles(request.UploadedFiles); err != nil {
		fmt.Printf("Cleanup failed: %s\n", err)
	}
	return result, interpreter.GetResponse().ExitCode
}
