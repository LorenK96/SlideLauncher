package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
)

var timeout = time.Duration(5 * time.Second)
var versionURL = "http://lkuich.com/projects/slide/maint/version"
var slideURL = "http://lkuich.com/projects/slide/maint/Slide.jar"

func downloadJar() {
	out, err := os.Create("Slide.jar")
	defer out.Close()

	if err != nil {
		fmt.Printf("Could not write Slide.jar\n")
	}

	client := http.Client{Timeout: timeout}
	jar, err := client.Get(slideURL)
	defer jar.Body.Close()
	if err != nil {
		fmt.Printf("Error downloading Slide.jar\n")
		runSlide()
		return
	}

	io.Copy(out, jar.Body)
}

func downloadVersion() {
	out, err := os.Create("version")
	defer out.Close()

	if err != nil {
		fmt.Printf("Could not write version\n")
	}

	client := http.Client{Timeout: timeout}
	version, err := client.Get(versionURL)
	defer version.Body.Close()
	if err != nil {
		fmt.Printf("Error downloading version\n")
		runSlide()
		return
	}

	io.Copy(out, version.Body)
}

func main() {
	conn, err := net.Dial("tcp", "google.com:80")
	_ = conn
	if err != nil {
		fmt.Printf("No internet")
		runSlide()
		return
	}

	// Remote version
	client := http.Client{Timeout: timeout}
	remoteVer, err := client.Get(versionURL)
	defer remoteVer.Body.Close()

	remoteVerBytes, err := ioutil.ReadAll(remoteVer.Body)
	if err != nil {
		fmt.Printf("Error getting version\n")
		runSlide()
		return
	}
	remoteVerString := string(remoteVerBytes[:])

	// Local version
	localVer, err := ioutil.ReadFile("version")
	localVerString := string(localVer[:])

	// No version present
	if err != nil {
		fmt.Printf("Version not present\n")
		update()
	} else {
		// Compare versions
		if localVerString != remoteVerString {
			fmt.Printf("Version mismatch\n")

			update()
		}
	}

	// Finally run the thing
	runSlide()
}

func runSlide() {
	path := "SOFTWARE\\JavaSoft\\Java Runtime Environment"
	javaVersion := getRegValue(path, "CurrentVersion")
	javaBin := getRegValue(path+"\\"+javaVersion, "JavaHome") + "\\bin\\java.exe"

	cmd := exec.Command(javaBin, "-jar", "Slide.jar")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, _ := cmd.Output()
	_ = out
}

func update() {
	downloadVersion()
	downloadJar()
}

func getRegValue(path string, key string) string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	s, _, err := k.GetStringValue(key)
	if err != nil {
		log.Fatal(err)
	}

	return s
}
