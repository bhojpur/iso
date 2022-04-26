package util

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	gc "github.com/bhojpur/iso/pkg/manager/api/core/garbagecollector"
	"github.com/bhojpur/iso/pkg/manager/api/core/logger"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	extensions "github.com/bhojpur/iso/pkg/manager/extensions"
	"github.com/bhojpur/iso/pkg/manager/solver"
	"github.com/ipfs/go-log/v2"
	"github.com/pterm/pterm"
	"go.uber.org/zap/zapcore"

	helpers "github.com/bhojpur/iso/pkg/manager/helpers"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	BhojpurEnvPrefix = "BHOJPUR"
)

var cfgFile string

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	setDefaults(viper.GetViper())
	// Bhojpur ISO support these priorities on read configuration file:
	// - command line option (if available)
	// - $PWD/.iso.yaml
	// - $HOME/.iso.yaml
	// - /etc/bhojpur/iso.yaml
	//
	// Note: currently a single viper instance support only one config name.

	viper.SetEnvPrefix(BhojpurEnvPrefix)
	viper.SetConfigType("yaml")

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Retrieve pwd directory
		pwdDir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		homeDir := helpers.GetHomeDir()

		if fileHelper.Exists(filepath.Join(pwdDir, ".iso.yaml")) || (homeDir != "" && fileHelper.Exists(filepath.Join(homeDir, ".iso.yaml"))) {
			viper.AddConfigPath(".")
			if homeDir != "" {
				viper.AddConfigPath(homeDir)
			}
			viper.SetConfigName(".iso")
		} else {
			viper.SetConfigName("iso")
			viper.AddConfigPath("/etc/bhojpur")
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Create EnvKey Replacer for handle complex structure
	replacer := strings.NewReplacer(".", "__")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetTypeByDefaultValue(true)
	// If a config file is found, read it in.
	viper.ReadInConfig()

}

var DefaultContext *context.Context

// InitContext inits the context by parsing the configurations from viper
// this is meant to be run before each command to be able to parse any override from
// the CLI/ENV
func InitContext(cmd *cobra.Command) (ctx *context.Context, err error) {

	c := &types.BhojpurConfig{}
	err = viper.Unmarshal(c)
	if err != nil {
		return
	}

	// Converts user-defined config into paths
	// and creates the required directory on the system if necessary
	c.Init()

	finalizerEnvs, _ := cmd.Flags().GetStringArray("finalizer-env")
	setCliFinalizerEnvs(c, finalizerEnvs)

	c.Solver.SolverOptions = types.SolverOptions{Type: types.SolverSingleCoreSimple, Concurrency: c.General.Concurrency}

	ctx = context.NewContext(
		context.WithConfig(c),
		context.WithGarbageCollector(gc.GarbageCollector(c.System.TmpDirBase)),
	)

	// Inits the context with the configurations loaded
	// It reads system repositories, sets logging, and all the
	// context which is required to perform Bhojpur ISO actions
	return ctx, initContext(cmd, ctx)
}

func setCliFinalizerEnvs(c *types.BhojpurConfig, finalizerEnvs []string) error {
	if len(finalizerEnvs) > 0 {
		for _, v := range finalizerEnvs {
			idx := strings.Index(v, "=")
			if idx < 0 {
				return errors.New("Found invalid runtime finalizer environment: " + v)
			}

			c.SetFinalizerEnv(v[0:idx], v[idx+1:])
		}
	}

	return nil
}

const (
	CommandProcessOutput = "command.process.output"
)

func initContext(cmd *cobra.Command, c *context.Context) (err error) {
	if logger.IsTerminal() {
		if !c.Config.Logging.Color {
			pterm.DisableColor()
		}
	} else {
		pterm.DisableColor()
		c.Debug("Not a terminal, colors disabled")
	}

	if c.Config.General.Quiet {
		pterm.DisableColor()
		pterm.DisableStyling()
	}

	level := c.Config.Logging.Level
	if c.Config.General.Debug {
		level = "debug"
	}

	if _, ok := cmd.Annotations[CommandProcessOutput]; ok {
		// Note: create-repo output is different, so we annotate in the cmd of create-repo CommandNoProcess
		// to avoid
		out, _ := cmd.Flags().GetString("output")
		if out != "terminal" {
			level = zapcore.Level(log.LevelFatal).String()
		}
	}

	// Init logging
	opts := []logger.LoggerOptions{
		logger.WithLevel(level),
	}

	if c.Config.Logging.NoSpinner {
		opts = append(opts, logger.NoSpinner)
	}

	if c.Config.Logging.EnableLogFile && c.Config.Logging.Path != "" {
		f := "console"
		if c.Config.Logging.JSONFormat {
			f = "json"
		}
		opts = append(opts, logger.WithFileLogging(c.Config.Logging.Path, f))
	}

	if c.Config.Logging.EnableEmoji {
		opts = append(opts, logger.EnableEmoji())
	}

	l, err := logger.New(opts...)

	c.Logger = l

	c.Debug("System rootfs:", c.Config.System.Rootfs)
	c.Debug("Colors", c.Config.Logging.Color)
	c.Debug("Logging level", c.Config.Logging.Level)
	c.Debug("Debug mode", c.Config.General.Debug)

	return
}

func setDefaults(viper *viper.Viper) {
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.enable_logfile", false)
	viper.SetDefault("logging.path", "/var/log/iso.log")
	viper.SetDefault("logging.json_format", false)
	viper.SetDefault("logging.enable_emoji", true)
	viper.SetDefault("logging.color", true)

	viper.SetDefault("general.concurrency", runtime.NumCPU())
	viper.SetDefault("general.debug", false)
	viper.SetDefault("general.quiet", false)
	viper.SetDefault("general.show_build_output", true)
	viper.SetDefault("general.fatal_warnings", false)
	viper.SetDefault("general.http_timeout", 360)

	u, err := user.Current()
	// os/user doesn't work in from scratch environments
	if err != nil || (u != nil && u.Uid == "0") {
		viper.SetDefault("general.same_owner", true)
	} else {
		viper.SetDefault("general.same_owner", false)
	}

	viper.SetDefault("system.database_engine", "boltdb")
	viper.SetDefault("system.database_path", "/var/cache/bhojpur")
	viper.SetDefault("system.rootfs", "/")
	viper.SetDefault("system.tmpdir_base", filepath.Join(os.TempDir(), "tmpiso"))
	viper.SetDefault("system.pkgs_cache_path", "packages")

	viper.SetDefault("repos_confdir", []string{"/etc/bhojpur/repos.conf.d"})
	viper.SetDefault("config_protect_confdir", []string{"/etc/bhojpur/config.protect.d"})
	viper.SetDefault("config_protect_skip", false)
	// TODO: Set default to false when we are ready for migration.
	viper.SetDefault("config_from_host", true)
	viper.SetDefault("cache_repositories", []string{})
	viper.SetDefault("system_repositories", []string{})
	viper.SetDefault("finalizer_envs", make(map[string]string))

	viper.SetDefault("solver.type", "")
	viper.SetDefault("solver.rate", 0.7)
	viper.SetDefault("solver.discount", 1.0)
	viper.SetDefault("solver.max_attempts", 9000)
}

// InitViper inits a new viper
// this is meant to be run just once at beginning to setup the root command
func InitViper(RootCmd *cobra.Command) {
	cobra.OnInitialize(initConfig)
	pflags := RootCmd.PersistentFlags()
	pflags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.iso.yaml)")
	pflags.BoolP("debug", "d", false, "debug output")
	pflags.BoolP("quiet", "q", false, "quiet output")
	pflags.Bool("fatal", false, "Enables Warnings to exit")
	pflags.Bool("enable-logfile", false, "Enable log to file")
	pflags.Bool("no-spinner", false, "Disable spinner.")
	pflags.Bool("color", true, "Enable/Disable color.")
	pflags.Bool("emoji", true, "Enable/Disable emoji.")
	pflags.Bool("skip-config-protect", true, "Disable config protect analysis.")
	pflags.StringP("logfile", "l", "", "Logfile path. Empty value disable log to file.")
	pflags.StringSlice("plugin", []string{}, "A list of runtime plugins to load")

	pflags.String("system-dbpath", "", "System db path")
	pflags.String("system-target", "", "System rootpath")
	pflags.String("system-engine", "", "System DB engine")

	pflags.String("solver-type", "", "Solver strategy ( Defaults none, available: "+solver.AvailableResolvers+" )")
	pflags.Float32("solver-rate", 0.7, "Solver learning rate")
	pflags.Float32("solver-discount", 1.0, "Solver discount rate")
	pflags.Int("solver-attempts", 9000, "Solver maximum attempts")
	pflags.Bool("live-output", true, "Show live output during build")

	pflags.Bool("same-owner", true, "Maintain same owner on uncompress.")
	pflags.Int("concurrency", runtime.NumCPU(), "Concurrency")
	pflags.Int("http-timeout", 360, "Default timeout for http(s) requests")

	viper.BindPFlag("system.database_path", pflags.Lookup("system-dbpath"))
	viper.BindPFlag("system.rootfs", pflags.Lookup("system-target"))
	viper.BindPFlag("system.database_engine", pflags.Lookup("system-engine"))
	viper.BindPFlag("solver.type", pflags.Lookup("solver-type"))
	viper.BindPFlag("solver.discount", pflags.Lookup("solver-discount"))
	viper.BindPFlag("solver.rate", pflags.Lookup("solver-rate"))
	viper.BindPFlag("solver.max_attempts", pflags.Lookup("solver-attempts"))

	viper.BindPFlag("logging.color", pflags.Lookup("color"))
	viper.BindPFlag("logging.enable_emoji", pflags.Lookup("emoji"))
	viper.BindPFlag("logging.enable_logfile", pflags.Lookup("enable-logfile"))
	viper.BindPFlag("logging.path", pflags.Lookup("logfile"))
	viper.BindPFlag("logging.no_spinner", pflags.Lookup("no-spinner"))
	viper.BindPFlag("general.concurrency", pflags.Lookup("concurrency"))
	viper.BindPFlag("general.debug", pflags.Lookup("debug"))
	viper.BindPFlag("general.quiet", pflags.Lookup("quiet"))
	viper.BindPFlag("general.fatal_warnings", pflags.Lookup("fatal"))
	viper.BindPFlag("general.same_owner", pflags.Lookup("same-owner"))
	viper.BindPFlag("plugin", pflags.Lookup("plugin"))
	viper.BindPFlag("general.http_timeout", pflags.Lookup("http-timeout"))
	viper.BindPFlag("general.show_build_output", pflags.Lookup("live-output"))

	// Currently I maintain this only from cli.
	viper.BindPFlag("no_spinner", pflags.Lookup("no-spinner"))
	viper.BindPFlag("config_protect_skip", pflags.Lookup("skip-config-protect"))

	// Extensions must be binary with the "bhojpur-" prefix to be able to be shown in the help.
	// we also accept extensions in the relative path where Bhojpur ISO is being started, "extensions/"
	exts := extensions.Discover("bhojpur", "extensions")
	for _, ex := range exts {
		cobraCmd := ex.CobraCommand()
		RootCmd.AddCommand(cobraCmd)
	}
}
