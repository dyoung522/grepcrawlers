/*
Copyright © 2025 Donovan C. Young <dyoung522@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nerdwerx/dccseeder/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	outputFile string
	debug      bool
	force      bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dccseeder [flags] <epub-files>",
	Short: "Reads DCC epubs and builds a list of known crawler numbers",
	Long: `dccseeder is a tool to read Dungeon Crawler Carl epubs and
and build a list of all the known crawler numbers.`,
	Run: gatherCrawlers,
}

func gatherCrawlers(cmd *cobra.Command, args []string) {
	var (
		crawlers   = make(lib.Crawlers)
		err        error
		sortedKeys []string
		ofile      io.WriteCloser
	)

	// Seed Carl
	crawlers.Add(lib.Crawler{ID: "4,122", Name: "Carl"})

	if len(args) < 1 {
		fmt.Println("Usage: dccseeder [--output <filename>] <epub-files>")
		os.Exit(1)
	}

	if debug {
		log.Println("Debug mode enabled.")
	}

	ofileName := cmd.Flag("output").Value.String()
	if ofileName != "" {
		ofile, err = os.Create(ofileName)
		if err != nil {
			log.Fatal("Error creating output file:", err)
		}
		defer func() {
			if err := ofile.Close(); err != nil {
				log.Fatal("Error closing output file:", err)
			}
		}()
		log.Printf("Writing to %q", ofileName)
	} else {
		log.Println("No output flag set, printing to STDOUT.")
		ofile = os.Stdout
	}

	for _, rawfile := range args {
		if rawfile == "" {
			log.Fatal("No EPUB file provided.")
		}

		if debug {
			log.Println("Reading EPUB file:", rawfile)
		}

		for id, crawler := range lib.ScanBook(rawfile, debug) {
			verb := "will NOT overwrite"

			if existing, ok := crawlers[id]; ok {
				if crawler != existing {
					if force {
						verb = "force overwrites"
					}
					log.Printf("Duplicate crawler # found: %s %s %q\n", crawler, verb, existing.Name)
					if !force {
						log.Println("Use --force if you wish to overwrite existing crawlers")
						continue
					}
				} else {
					if debug {
						log.Printf("Duplicate crawler mention found: %s, skipping\n", crawler)
					}
					continue
				}
			}
			crawlers.Add(crawler)
		}
	}

	if sortedKeys, err = crawlers.SortIDsNumerically(); err == nil {
		for _, k := range sortedKeys {
			if debug {
				log.Printf("Found crawler: ID=%s, Name=%s\n", crawlers[k].ID, crawlers[k].Name)
			}
			if _, err := ofile.Write(crawlers[k].MarshalCSV()); err != nil {
				log.Fatal("Error writing to output file:", err)
			}
		}
	} else {
		log.Fatal("Error sorting crawlers: %w", err)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite of duplicates")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for results (default is stdout)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
