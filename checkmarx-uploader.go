// checkmarx-uploader.go
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	tf "github.com/mtesauro/tfclient"
)

var version = "1.3"

// Stuff for config file
var configFile string = "checkmarx-uploader.config"
var watchLocation = ""
var logLocation = ""

// Stuff for logger
var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func Init(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {

	Trace = log.New(traceHandle, "TRACE:   ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "INFO:    ", log.Ldate|log.Ltime)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime)
	Error = log.New(errorHandle, "ERROR:   ", log.Ldate|log.Ltime|log.Lshortfile)

	// Allows for these to be use globally
	// Trace.Println("Trace message")
	// Info.Println("Info message")
	// Warning.Println("Warning message")
	// Error.Println("Error message")
}

func readConfig() (bool, error) {
	found := false

	_, err := os.Stat(configFile)
	if err != nil {
		err := createDefaultConfig()
		if err != nil {
			// Config file has been found
			msg := fmt.Sprintf("  Unable to create config file\n\t at %s\n", configFile)
			return found, errors.New(msg)
		}
	}

	// Read configuration file to pull out any configured items
	file, err := os.Open(configFile)
	if err != nil {
		msg := fmt.Sprintf("  Unable to open config file at %s\n  Error message was: %s\n", configFile, err.Error())
		return found, errors.New(msg)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Handle lines that are not comments
		if strings.Index(scanner.Text(), "#") != 0 {
			found = true
			line := strings.Trim(scanner.Text(), " ")

			// Pull out the config values
			if strings.Contains(line, "watchLocation=") {
				v := strings.SplitAfterN(line, "=", 2)
				watchLocation = strings.Replace(strings.TrimSpace(v[1]), "\"", "", -1)
			}

			if strings.Contains(line, "logLocation=") {
				v := strings.SplitAfterN(line, "=", 2)
				logLocation = strings.Replace(strings.TrimSpace(v[1]), "\"", "", -1)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		msg := fmt.Sprintf("Error parsing %s\n  Error message was: %s\n", configFile, err.Error())
		return found, errors.New(msg)
	}

	return found, nil
}

func createDefaultConfig() error {
	// Create a default config file
	defaultConfig := "# Default checkmarx-uploader.config file\n"
	defaultConfig += "# Please put the values for your Checkmarx installation below\n"
	defaultConfig += "# as these are simple place holders\n"
	defaultConfig += "# NOTE: Full or relative paths can be used\n"
	defaultConfig += "#   so the example below looks for a directory called 'checkmarx'\n"
	defaultConfig += "#   in the same directory checkmarx-uploader is run in.\n"
	defaultConfig += "watchLocation=\"checkmarx\"\n"
	defaultConfig += "logLocation=\"log\"\n"

	// Write a default config
	fileBytes := []byte(defaultConfig)
	err := ioutil.WriteFile(configFile, fileBytes, 0600)
	if err != nil {
		msg := fmt.Sprintf("\nUnable to write config file.\nPlease check permissions of %s\n", configFile)
		return errors.New(msg)
	}

	// Exit here because you probably want to edit the default config file
	fmt.Println("=====[ Default config file created ]=====")
	fmt.Println("")
	fmt.Printf("A default configuration file for checkmarx-uploader has been created \n")
	fmt.Printf("in the current working directory named '%s'.  Please edit the default\n", configFile)
	fmt.Printf("values before running this program again.\n")
	fmt.Printf("Cheers!\n")
	fmt.Println("=====[ Default config file created ]=====")
	fmt.Println("")
	os.Exit(1)

	return nil
}

func getAppId(f os.FileInfo) (int, error) {
	appId := 0

	// Strip off AppID from the filename
	i := strings.Index(f.Name(), "[") - 1

	// Check i is not less then 1 smallest index for single digit AppID's or no "[" found
	if i < 1 {
		msg := "Bad file name -  No \"[\" found - file names begin with {AppID}_[App Name] - e.g. 12_[Foo].Foo-28.4.2015-16.20.41.xml"
		Error.Printf("%s", msg)
		return 0, errors.New(msg)
	}

	appId, err := strconv.Atoi(f.Name()[0:(i)])
	if err != nil {
		Error.Printf("Error converting App ID to a int: The error was: \"%v\"", err)
		fmt.Printf("\nError converting App ID to a int\n - The error was: \"%v\"\n\n", err)
		os.Exit(1)
	}

	return appId, nil
}

func scanUpload(d string, n string, app int) error {
	// Create a client to talk to the API
	tfc, err := tf.CreateClient()
	if err != nil {
		msg := fmt.Sprintf("Error creating TF client. The error was: %+v", err)
		return errors.New(msg)
	}

	filePath := path.Join(d, n)

	// Upload the scan file to the requested App
	upResp, err := tf.ScanUpload(tfc, app, filePath)
	if err != nil {
		msg := fmt.Sprintf("An error occurred during scan upload. The error was: %+v", err)
		return errors.New(msg)
	}

	var u tf.UpldResp
	err = tf.MakeUploadStruct(&u, upResp)
	if err != nil {
		msg := fmt.Sprintf("An error occurred during JSON parsing. The error was: %+v", err)
		return errors.New(msg)
	}

	// Alternative, you could count the map in the struct - len(map)==0 to see if its empty
	if u.Success != true {
		msg := fmt.Sprintf("An error occured during upload. Threadfix reported: %+v", u.Msg)
		return errors.New(msg)
	}

	return nil
}

func moveBadFile(f os.FileInfo) {
	// If a parse-errors directory doesn't exist, create it
	d := path.Join(watchLocation, "parse-errors")
	_, err := os.Stat(d)
	if err != nil {
		// Need to create config directory
		Info.Printf("Creating %s for 'bad' files", d)
		err = os.Mkdir(d, 0700)
		if err != nil {
			Error.Printf("Unable to create config directory at %s", d)
			return
		}
	}

	// Move the file with parse errors
	mv := path.Join(d, f.Name())
	err = os.Rename(path.Join(watchLocation, f.Name()), mv)
	if err != nil {
		Error.Printf("Unable to move %s to %s - %+v", f.Name(), mv, err)
		return
	}
	Info.Printf("Problem file moved to %s", mv)

	return
}

func copyFile(dst string, src string) (int64, error) {
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}

func main() {
	// Read location from a config file in CWD
	_, err := readConfig()
	if err != nil {
		fmt.Printf("Unable to read config file %s.  Error was:\n  %+v\n", configFile, err)
		os.Exit(1)
	}

	// Setup logging
	logPath := path.Join(logLocation, "checkmarx-uploader.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("\nPlease create any directories needed to write logs to %v\n\n", logPath)
		log.Fatalf("Failed to open log file %s.  Error was:\n  %+v\n", logPath, err)
	}
	// Log everthing to the specificied log file location
	Init(logFile, logFile, logFile, logFile)

	// Log the version number
	Info.Printf("Starting up checkmarx-uploader version %v", version)

	// Read the directory which we're watching
	dir, err := os.Open(watchLocation)
	if err != nil {
		Error.Printf("Unable to read output directory.  Error was: %+v", err)
		fmt.Printf("Unable to read output directory.  Error was:\n  %+v\n", err)
		os.Exit(1)
	}
	defer dir.Close()
	Info.Printf("Opening uploads directory of %s", watchLocation)

	// Get a list of files in that directory
	files, err := dir.Readdir(-1)
	if err != nil {
		Error.Printf("Unable to read files in the output directory.  Error was: %+v", err)
		fmt.Printf("Unable to read files in the output directory.  Error was:\n  %+v\n", err)
		os.Exit(1)
	}
	Info.Printf("Reading upload files from %s", watchLocation)

	if len(files) == 0 {
		Warning.Printf("No files to process in %s", watchLocation)
		Info.Printf("Exiting early since there are no files in %s", watchLocation)
		os.Exit(0)
	}

	name := make(map[string]int)
	for _, f := range files {
		// Only look at regular files
		if f.Mode().IsRegular() {
			appId, err := getAppId(f)
			if err != nil {
				// Cannot get an AppID from this file, move to bad files directory
				Error.Printf("Unable to get AppID from %+v.", f.Name())
				moveBadFile(f)
			} else {
				Info.Printf("Uploading %+v", f.Name())
				err := scanUpload(watchLocation, f.Name(), appId)
				if err != nil {
					// Log an error and leave the file off the deletion list
					Error.Printf("Error uploading %+v.  Error was: %+v", f.Name(), err)
					// Move parse errors to parse-errors directory - create if needed
					if strings.Contains(err.Error(), "JSON parsing") {
						Error.Printf("Parse error for %+v", f.Name())
						moveBadFile(f)
					}

				} else {
					Info.Printf("Successful upload of %+v to ThreadFix AppID %+v", f.Name(), appId)
					// Set the name for later deletion
					name[f.Name()] = appId
				}
			}
		}
	}

	for k, v := range name {
		Info.Printf("Deleting file %v after uploading to ThreadFix AppID %v", k, v)
		err := os.Remove(path.Join(watchLocation, k))
		if err != nil {
			Error.Printf("Problem deleting %s.  Error was %+v", path.Join(watchLocation, k), err)
		}
	}

}
