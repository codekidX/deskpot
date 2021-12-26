package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

type NewDeskpotProject struct {
	Name      string
	NameLower string
	RunID     int64
}

type CopyrightConfig struct {
	Year string `json:"year,omitempty"`
	Name string `json:"name,omitempty"`
}

type PublishConfig struct {
	Icon      string          `json:"icon,omitempty"`
	Copyright CopyrightConfig `json:"copyright,omitempty"`
}

type DeveloperConfig struct {
	Timeout int `json:"timeout,omitempty"`
}

type DeskpotConfig struct {
	PackageIdentifier string          `json:"identifier,omitempty"`
	AppName           string          `json:"name,omitempty"`
	AppDescription    string          `json:"description,omitempty"`
	RunID             string          `json:"run_id,omitempty"`
	AppVersion        string          `json:"version,omitempty"`
	OSXCategory       string          `json:"osx_category,omitempty"`
	Publish           PublishConfig   `json:"publish,omitempty"`
	Dev               DeveloperConfig `json:"dev,omitempty"`
}

// ----------- SCAFFOLD ----------------

//go:embed templates/scaffold/mainGo.plate
var mainGoFile string

//go:embed templates/scaffold/deskpotJson.plate
var deskpotJson string

//go:embed templates/scaffold/indexJsx.plate
var indexJsx []byte

//go:embed templates/scaffold/appJsx.plate
var appJsx []byte

//go:embed templates/scaffold/webpackDevConfigJs.plate
var webpackDevConfigJs []byte

//go:embed templates/scaffold/webpackProdConfigJs.plate
var webpackProdConfigJs []byte

//go:embed templates/scaffold/indexHtml.plate
var indexHtml []byte

//go:embed templates/scaffold/babelrc.plate
var babelRc []byte

var webpackServeProc *exec.Cmd

// ----------- PACKAGING ----------------

//go:embed templates/package/infoPlist.plate
var infoPlist string

//go:embed templates/package/icon.icns
var appIcon []byte

func main() {
	var root = &cobra.Command{
		Use:   "dpot",
		Short: "Create, manage and ship webview based desktop applications in a breeze",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("i guess you need --help")
		},
	}

	var newCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new webview project",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Name of the project not specified")
				return
			}

			name := args[0]
			projectFolder := path.Join(".", name)
			if f, _ := os.Stat(projectFolder); f != nil {
				// some folder with name exists
				fmt.Printf("Folder with name: %s already exists", name)
				return
			}

			// -- start with creating a new deskpot project --
			// create the project folder
			if err := os.MkdirAll(projectFolder, 0755); err != nil {
				panic(err)
			}

			var buf bytes.Buffer
			config := NewDeskpotProject{
				Name:      strings.Title((name)),
				NameLower: strings.ToLower(name),
				RunID:     time.Now().UnixMilli(),
			}

			// compile main.go template
			t, _ := template.New("main.go").Parse(mainGoFile)

			if err := t.Execute(&buf, config); err != nil {
				panic(err)
			}
			// write the main.go file from the template
			if err := ioutil.WriteFile(path.Join(projectFolder, "main.go"), buf.Bytes(), 0755); err != nil {
				panic(err)
			}
			buf.Reset()

			// compile deskpot.json template
			t, _ = template.New("deskpot.json").Parse(deskpotJson)

			if err := t.Execute(&buf, config); err != nil {
				panic(err)
			}
			// write the main.go file from the template
			if err := ioutil.WriteFile(path.Join(projectFolder, "deskpot.json"), buf.Bytes(), 0755); err != nil {
				panic(err)
			}
			buf.Reset()

			// check if npm exists
			if _, err := exec.LookPath("npm"); err != nil {
				fmt.Println("Node Package Manager does not exists, are you sure it is in path?")
				os.Exit(1)
			}

			// Initialize a node project and install react and webpack
			os.Chdir(projectFolder)
			run(fmt.Sprintf("go mod init %s", name))
			run("go mod tidy")
			run("npm init -y")
			run("npm install react react-dom")
			run("npm install webpack webpack-cli webpack-dev-server html-webpack-plugin" +
				" inline-source-webpack-plugin css-loader style-loader --save-dev")
			run("npm i --save-dev babel-loader @babel/preset-env @babel/core @babel/plugin-syntax-dynamic-import" +
				" @babel/plugin-transform-runtime @babel/preset-react babel-eslint @babel/runtime")

			// create the react project
			uiFolder := path.Join(".", "ui")
			publicFolder := path.Join(".", "public")
			prodOutFolder := path.Join(".", "out", "prod")
			// ui/index.jsx
			if err := os.MkdirAll(uiFolder, 0755); err != nil {
				panic(err)
			}
			if err := os.MkdirAll(publicFolder, 0755); err != nil {
				panic(err)
			}
			if err := os.MkdirAll(prodOutFolder, 0755); err != nil {
				panic(err)
			}
			if err := ioutil.WriteFile(path.Join(uiFolder, "index.jsx"), indexJsx, 0755); err != nil {
				panic(err)
			}
			if err := ioutil.WriteFile(path.Join(uiFolder, "App.jsx"), appJsx, 0755); err != nil {
				panic(err)
			}
			// .babelrc
			if err := ioutil.WriteFile(path.Join(".", ".babelrc"), babelRc, 0755); err != nil {
				panic(err)
			}
			// .gitignore
			if err := ioutil.WriteFile(path.Join(".", ".gitignore"), []byte("node_modules/\nout/"), 0755); err != nil {
				panic(err)
			}
			// webpack.config.js = development
			if err := ioutil.WriteFile(path.Join(".", "webpack.config.js"), webpackDevConfigJs, 0755); err != nil {
				panic(err)
			}

			// webpack.prod.config.js
			// TODO: optimize for production export
			if err := ioutil.WriteFile(path.Join(".", "webpack.prod.config.js"), webpackProdConfigJs, 0755); err != nil {
				panic(err)
			}
			// public/index.html
			if err := ioutil.WriteFile(path.Join(publicFolder, "index.html"), indexHtml, 0755); err != nil {
				panic(err)
			}

			// dummy out/prod/index.html
			if err := ioutil.WriteFile(path.Join(prodOutFolder, "index.html"), indexHtml, 0755); err != nil {
				panic(err)
			}

			fmt.Printf("\n\nEnjoy developing with Deskpot!!\n\n Next steps:\n	cd %s\n	deskpot run\n", name)
		},
	}

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Runs your react app in webpack-dev-server",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: we can refactor unmarshalling of the config into
			// a function as it is also used inside the publish function
			if f, _ := os.Stat("./deskpot.json"); f == nil {
				fmt.Println("This is not a deskpot project. Visit github.com/codekidx/deskpot for more info")
				return
			}

			b, err := ioutil.ReadFile("./deskpot.json")
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			var config DeskpotConfig
			if err := json.Unmarshal(b, &config); err != nil {
				fmt.Println(err.Error())
				return
			}

			if config.Dev.Timeout != 0 {
				config.Dev.Timeout = 5
			}
			timeout := time.Duration(config.Dev.Timeout)

			go runWebpackServe()
			os.Setenv("DPOT_DEV", "deskpot")
			time.Sleep(timeout * time.Second)
			run("go run main.go")
			os.Unsetenv("DPOT_DEV")
			killWebpackServe()
		},
	}

	var packageCmd = &cobra.Command{
		Use:   "pack",
		Short: "Package your Deskpot application",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Supported platforms: mac")
				return
			}

			if f, _ := os.Stat("./deskpot.json"); f == nil {
				fmt.Println("This is not a deskpot project. Visit github.com/codekidx/deskpot for more info")
				return
			}

			b, err := ioutil.ReadFile("./deskpot.json")
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			var config DeskpotConfig
			if err := json.Unmarshal(b, &config); err != nil {
				fmt.Println(err.Error())
				return
			}

			targetFolder := path.Join(".", "out", config.AppVersion)
			if f, _ := os.Stat(targetFolder); f != nil {
				os.RemoveAll(targetFolder)
				os.MkdirAll(targetFolder, 0755)
			}

			appFolder := path.Join(targetFolder, fmt.Sprintf("%s.app", config.AppName))
			contentsFolder := path.Join(appFolder, "Contents")
			execFolder := path.Join(contentsFolder, "MacOS")
			resFolder := path.Join(contentsFolder, "Resources")
			os.MkdirAll(appFolder, 0755)
			os.MkdirAll(execFolder, 0755)
			os.MkdirAll(resFolder, 0755)

			// compile Info.plist template
			var buf bytes.Buffer
			t, _ := template.New("infoplist").Parse(infoPlist)

			if err := t.Execute(&buf, config); err != nil {
				panic(err)
			}
			// write the Info.plist into the appFolder
			if err := ioutil.WriteFile(path.Join(contentsFolder, "Info.plist"), buf.Bytes(), 0755); err != nil {
				panic(err)
			}
			buf.Reset()

			// if developer has mentioned a custom icon path in config
			if config.Publish.Icon != "DEFAULT" {
				// check if icon is present at source
				if f, _ := os.Stat(config.Publish.Icon); f == nil {
					fmt.Printf("Icon not present at: %s\n", config.Publish.Icon)
					return
				}

				if err := copy(config.Publish.Icon, path.Join(resFolder, "AppIcon.icns")); err != nil {
					panic(err)
				}
			} else {
				ioutil.WriteFile(path.Join(resFolder, "AppIcon.icns"), appIcon, 0755)
			}

			run(fmt.Sprintf("go build -o %s main.go", path.Join(execFolder, config.AppName)))
		},
	}

	root.AddCommand(newCmd)
	root.AddCommand(runCmd)
	root.AddCommand(packageCmd)

	if err := root.Execute(); err != nil {
		panic(err)
	}

}

func run(cmd string) {
	splitted := strings.Split(cmd, " ")
	fmt.Println(":]	", splitted)
	c := exec.Command(splitted[0], splitted[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Run()
}

func runWebpackServe() {
	webpackServeProc = exec.Command("webpack", "serve")
	webpackServeProc.Stdout = os.Stdout
	webpackServeProc.Stderr = os.Stderr
	webpackServeProc.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	webpackServeProc.Run()
}

func killWebpackServe() {
	pgid, err := syscall.Getpgid(webpackServeProc.Process.Pid)
	if err != nil {
		fmt.Println("Cannot kill process", err.Error())
	}
	syscall.Kill(-pgid, syscall.SIGINT)
}

// copy file from `src` path to the `dst` path
func copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
