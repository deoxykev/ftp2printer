package main

import (
	"fmt"
	"io"
	"os"

	"github.com/deoxykev/ftp2printer/m/v2/ftp"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "ftp2printer",
	Short: "FTP to Printer bridge",
	Run:   run,
}

func init() {
	// Cobra setup
	rootCmd.PersistentFlags().String("username", "admin", "FTP username")
	rootCmd.PersistentFlags().String("password", "admin", "FTP password")
	rootCmd.PersistentFlags().Bool("keepfiles", false, "Keep files after processing")

	// Bind Viper to the flags
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("keepfiles", rootCmd.PersistentFlags().Lookup("keepfiles"))

	// Viper configuration
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func run(cmd *cobra.Command, args []string) {
	fmt.Println(len(args))
	fs := afero.NewMemMapFs()

	opts := &ftp.Options{
		Username:  viper.GetString("username"),
		Password:  viper.GetString("password"),
		Keepfiles: viper.GetBool("keepfiles"),
	}

	ftp.StartServer(opts, fs, func(file io.Reader) (int64, error) {
		return io.Copy(os.Stdout, file)
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
