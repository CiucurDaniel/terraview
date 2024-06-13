/*
Copyright © 2024 Daniel Ciucur ciucur.daniel14@gmail.com
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/CiucurDaniel/terraview/internal/config"
	"github.com/CiucurDaniel/terraview/internal/graph"
	"github.com/CiucurDaniel/terraview/internal/render"
	"github.com/CiucurDaniel/terraview/internal/tfstatereader"
	"github.com/spf13/cobra"
)

// Define the format, url, and config-file flags
var format string
var url string
var configFile string

// printCmd represents the print command
var printCmd = &cobra.Command{
	Use:   "print [path]",
	Short: "Print diagram from terraform code",
	Long: `Print diagram from terraform code. This command receives exactly one arg 
representing the path to the main.tf file. For example:

terraview print .
or
terraview print /users/Mike/terraform/`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		// Load the configuration if a config-file path is provided
		if configFile != "" {
			err := config.LoadConfig(configFile)
			if err != nil {
				fmt.Println(fmt.Errorf("ERROR: Could not load config: %v", err))
				return
			}
		}
		cfg := config.GetConfig()

		// Determine the state file path from the url flag or the path argument
		stateFilePath := url
		if stateFilePath == "" {
			stateFilePath = filepath.Join(path, "terraform.tfstate")
		}

		// Create a TFStateHandler
		handler, err := tfstatereader.NewTFStateHandler(stateFilePath)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to create TFStateHandler: %v", err))
			return
		}

		futureDiagram, err := graph.PrepareGraphForPrinting(path, cfg, handler)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to prepare graph for printing: %v", err))
			return
		}

		// Determine the output format from the flag
		if format == "" {
			format = "png" // Default format
		}

		// Save the graph in the specified format
		err = render.SaveGraphAs(futureDiagram, "./diagram", format)
		if err != nil {
			fmt.Println(fmt.Errorf("error occurred generating image: %v", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(printCmd)

	// Define the format flag
	printCmd.Flags().StringVarP(&format, "format", "f", "png", "Output format (png, jpg, svg, pdf, dot)")

	// Define the url flag
	printCmd.Flags().StringVarP(&url, "url", "u", "", "URL to the terraform state file (local file, http/https, s3, remote, gs, azurerm). Defaults to local if flag omitted")

	// Define the config-file flag
	printCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "Path to the configuration file. Defaults to built-in config if flag omitted")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// printCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// printCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
