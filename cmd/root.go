package cmd

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	logFile string
	gethURI string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gethalyzer",
	Short: "Custom Geth log analyzer",
	Long: `

         _          _            _       _    _                   _           _        _       _                 _            _      
        /\ \       /\ \         / /\    / /\ / /\                /\ \     _  /\ \     /\_\   /\ \               /\ \         /\ \    
       /  \ \      \_\ \       / / /   / / // /  \              /  \ \   /\_\\ \ \   / / /  /  \ \             /  \ \       /  \ \   
      / /\ \ \     /\__ \     / /_/   / / // / /\ \            / /\ \ \_/ / / \ \ \_/ / /__/ /\ \ \           / /\ \ \     / /\ \ \  
     / / /\ \_\   / /_ \ \   / /\ \__/ / // / /\ \ \          / / /\ \___/ /   \ \___/ //___/ /\ \ \         / / /\ \_\   / / /\ \_\ 
    / /_/_ \/_/  / / /\ \ \ / /\ \___\/ // / /  \ \ \        / / /  \/____/     \ \ \_/ \___\/ / / /        / /_/_ \/_/  / / /_/ / / 
   / /____/\    / / /  \/_// / /\/___/ // / /___/ /\ \      / / /    / / /       \ \ \        / / /        / /____/\    / / /__\/ /  
  / /\____\/   / / /      / / /   / / // / /_____/ /\ \    / / /    / / /         \ \ \      / / /    _   / /\____\/   / / /_____/   
 / / /______  / / /      / / /   / / // /_________/\ \ \  / / /    / / /           \ \ \     \ \ \__/\_\ / / /______  / / /\ \ \     
/ / /_______\/_/ /      / / /   / / // / /_       __\ \_\/ / /    / / /             \ \_\     \ \___\/ // / /_______\/ / /  \ \ \    
\/__________/\_\/       \/_/    \/_/ \_\___\     /____/_/\/_/     \/_/               \/_/      \/___/_/ \/__________/\/_/    \_\/    
                                                                                                                                     
This is a log analyzer that is intended to be used along with a hacked version 
of Geth which logs output to show the miner transaction logic. This tool is 
intended for demonstration purposes only and will be used in an interview
with Kaliedo. Please refer to the ../README.md for more information.
`,
	Run: runEthalyzer,
}

func runEthalyzer(cmd *cobra.Command, args []string) {
	prompt := promptui.Select{
		Label: "Select Day",
		Items: []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday",
			"Saturday", "Sunday"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Printf("You choose %q\n", result)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initEthalyzer)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&logFile, "logfile", "l", "$HOME/node1-minder.log", "geth miner log output file (default is $HOME/node1-minder.log)")
	rootCmd.PersistentFlags().StringVarP(&gethURI, "gethuri", "g", "http://localhost:8501", "URI to the geth node RPC Apis (default is 'http://localhost:8501')")
}

// initConfig reads in config file and ENV variables if set.
func initEthalyzer() {
	// TODO setup geth connection here. Also start reading from the log file here.
}
