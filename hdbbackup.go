// +build hdbbackup

package main

import (
	"os"

	. "github.com/greenplum-db/gpbackup/backup"
	"github.com/greenplum-db/gpbackup/options"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "hdbbackup",
		Short:   "hdbbackup is the parallel backup utility for HDB",
		Args:    cobra.NoArgs,
		Version: GetVersion(),
		Run: func(cmd *cobra.Command, args []string) {
			defer DoTeardown()
			DoFlagValidation(cmd)
			DoSetup()
			DoBackup()
		}}
	rootCmd.SetArgs(options.HandleSingleDashes(os.Args[1:]))
	DoInit(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}
