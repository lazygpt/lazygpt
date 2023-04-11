//

package main

import (
	"github.com/lazygpt/lazygpt/cmd/lazygpt/app"
)

func main() {
	lazyGPT := app.NewLazyGPTApp()
	lazyGPT.InitConfig()
	lazyGPT.Execute()
}
