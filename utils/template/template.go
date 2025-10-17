package template

import (
	"fmt"
	"runtime"
	"time"
)

// http://patorjk.com/software/taag/#p=testall&f=Alpha&t=scream
// Isometric3 | Star Wars

const TemplateScream = `      ___           ___           ___           ___           ___           ___     
     /  /\         /  /\         /  /\         /  /\         /  /\         /__/\    
    /  /:/_       /  /:/        /  /::\       /  /:/_       /  /::\       |  |::\   
   /  /:/ /\     /  /:/        /  /:/\:\     /  /:/ /\     /  /:/\:\      |  |:|:\  
  /  /:/ /::\   /  /:/  ___   /  /:/~/:/    /  /:/ /:/_   /  /:/~/::\   __|__|:|\:\ 
 /__/:/ /:/\:\ /__/:/  /  /\ /__/:/ /:/___ /__/:/ /:/ /\ /__/:/ /:/\:\ /__/::::| \:\
 \  \:\/:/~/:/ \  \:\ /  /:/ \  \:\/:::::/ \  \:\/:/ /:/ \  \:\/:/__\/ \  \:\~~\__\/
  \  \::/ /:/   \  \:\  /:/   \  \::/~~~~   \  \::/ /:/   \  \::/       \  \:\      
   \__\/ /:/     \  \:\/:/     \  \:\        \  \:\/:/     \  \:\        \  \:\     
     /__/:/       \  \::/       \  \:\        \  \::/       \  \:\        \  \:\    
     \__\/         \__\/         \__\/         \__\/         \__\/         \__\/    `

const TemplateDividingLine = `====================================================================================`

func TemplateInit(AppName, ConfigPath string) {
	printTemplate(AppName, ConfigPath)
}

func printTemplate(AppName, ConfigPath string) {
	fmt.Println(TemplateScream)
	fmt.Println()
	fmt.Printf("App Name:   %s\n", AppName)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CPU Cores:  %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Config:     %s\n", ConfigPath)
	fmt.Printf("Start Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("Hello Scream!")
	fmt.Println(TemplateDividingLine)
	fmt.Println()
}
