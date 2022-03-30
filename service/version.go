package service

import (
	"fmt"
	"github.com/jibenliu/utMsgDaemon/utils/color"
)

// see https://patorjk.com/software/taag/#p=testall&f=Graffiti&t=utMsgDaemon
const _UI = `
██╗   ██╗████████╗███╗   ███╗███████╗ ██████╗ ██████╗  █████╗ ███████╗███╗   ███╗ ██████╗ ███╗   ██╗
██║   ██║╚══██╔══╝████╗ ████║██╔════╝██╔════╝ ██╔══██╗██╔══██╗██╔════╝████╗ ████║██╔═══██╗████╗  ██║
██║   ██║   ██║   ██╔████╔██║███████╗██║  ███╗██║  ██║███████║█████╗  ██╔████╔██║██║   ██║██╔██╗ ██║
██║   ██║   ██║   ██║╚██╔╝██║╚════██║██║   ██║██║  ██║██╔══██║██╔══╝  ██║╚██╔╝██║██║   ██║██║╚██╗██║
╚██████╔╝   ██║   ██║ ╚═╝ ██║███████║╚██████╔╝██████╔╝██║  ██║███████╗██║ ╚═╝ ██║╚██████╔╝██║ ╚████║
 ╚═════╝    ╚═╝   ╚═╝     ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
                                                                                                    
`

var (
	Version   string
	GoVersion string
	BuildTime string
	GitBranch string
	GitTag    string
	GitCommit string
)

func init() {
	fmt.Println(color.Blue(_UI))
	fmt.Println(color.Green(fmt.Sprintf("Version: %s", Version)))
	fmt.Println(color.Green(fmt.Sprintf("GoVersion: %s", GoVersion)))
	fmt.Println(color.Green(fmt.Sprintf("GitBranch: %s", GitBranch)))
	fmt.Println(color.Green(fmt.Sprintf("GitTag: %s", GitTag)))
	fmt.Println(color.Green(fmt.Sprintf("GitCommit: %s", GitCommit)))
	fmt.Println(color.Green(fmt.Sprintf("Built: %s", BuildTime)))
}
