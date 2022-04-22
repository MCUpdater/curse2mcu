package cmd

import (
	"fmt"
	"github.com/mcupdater/curse2mcu/pkg/mcu"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var globalOpts struct {
	outFile string
}

var rootCmd = &cobra.Command{
	Use:   "curse2mcu",
	Short: "A command line tool for importing modpacks from Curse into MCUpdater.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing required argument: <cursemod.zip>")
		}
		if mcu.FileExists(args[0]) {
			log.Printf("Found pack at %q, analyzing...\n", args[0])
		} else {
			return fmt.Errorf("file not found %q", args[0])
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return mcu.ImportPackage(args[0], globalOpts.outFile)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&globalOpts.outFile,
		"out",
		"serverpack.xml",
		"xml file to output resulting pack to",
	)
	//_ = rootCmd.MarkPersistentFlagRequired("out")
}
