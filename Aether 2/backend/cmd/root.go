package cmd

import (
	"aether-core/services/configstore"
	"aether-core/services/globals"
	"aether-core/services/logging"
	"fmt"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

// cmdRoot represents the base command when called without any subcommands
var cmdRoot = &cobra.Command{
	Use:   "mre",
	Short: "MRE communicates with other computers using Mim-based subprotocols, persist objects received over the network, and serves the objects to other requesters as specified by the Mim protocol spec.",
	Long: `Mim Runtime Environment is the standalone executable that handles any Mim-based app's communication with the external world. After spinning up a running instance of this executable, it will act as a database for your app that you can query through appropriate local APIs.

For example: The app 'Aether' uses MRE to communicate in c0 and dweb Mim subprotocols. For more information, please see https://getaether.net. `,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`You've attempted to run the Mim Runtime Environment without any commands. MRE requires some variables to be passed to it to be able to do what you want.

Please run "mre --help" to see all available options.
`)
	},
}

// This is called by main.main(). It only needs to happen once to the cmdRoot.
func Execute() {
	if err := cmdRoot.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Below are methods used in the cmd library. These are not specifically related to root cmd.

func showIntro() {
	// color.Cyan(`
	//                  1ttfffLLLLLLLLLffft
	//              11111ttfffLLLLLLLLLffftt111
	//           111ttfLLLCCGGG000000GGGCCLLLfft111
	//        1111ffLLCG00880000GGGG00008880GCCLLft111
	//       11tfLLCG0880GCCLLLLLCCLLLLLCCGG0880CLLft111
	//     111fLLC0880CCLLLLLLLLL08CLLLLLLLLLCG880GLLft11
	//    11tLLCG88GCLCCLLLLLLLLL08CLLLLLLLLCCLLG0@0CLLf11
	//   11tLLC0@0CLLL080CLLLLLLLG8LLLLLLLCG88CLLLG88GLLf11
	//   1tLLC8@GLLLLLCG880CLLLLLG8LLLLLCG880CLLLLLC88GLLf1
	//  11LLL0@GLLLLLLLLLG080CLLLG8LLLCG88GCLLLLLLLLC88CLLt1
	//  1tLLG@8LLLLLLLLLLLLC080CLG8LCG80GCLLLLLLLLLLLG@0LLL1
	//  1LLL0@GLLLLLLLLLLLLLLCG8008080GCLLLLLLLLLLLLLC8@CLLt
	//  1LLL0@CLLG000000000000G0@@@@800000000000000CLL0@CLLt
	//  1LLL0@GLLLCCCCCCCCCCCCG0088080CCCCCCCCCCCCCLLC8@CLLt
	//  1tLLG@8LLLLLLLLLLLLCG08GLG8LC080CLLLLLLLLLLLLG@0LLL1
	//  11LLC8@GLLLLLLLLLCG80GLLLG8LLLC080CLLLLLLLLLC8@GLLt1
	//   1tLLC8@GLLLLLLC080GLLLLLG8LLLLLC080GCLLLLLC8@GLLf1
	//   11tLLC8@0CLLLG80GLLLLLLLG8LLLLLLLC080CLLLG88GLLf11
	//    11tLLCG88GLLCCLLLLLLLLL08CLLLLLLLLCCLLC088CLLf11
	//     111fLLC0880CLLLLLLLLLL08CLLLLLLLLLCG080GLLft111
	//      111tfLLCG8880GCCLLLLLCCLLLLLLCCG0880CLLft1111
	//        1111ffLLCG0088000GGGGGGG0008800CCLLft111
	//           1111tffLLCCGG00000000GGGCLLLftt1111
	//              11111ttfffLLLLLLLLLffftt11111
	//                  1ttfffLLLLLLLLLffft
	//    `)

	colorSet := color.New(color.FgWhite)
	colorSet = colorSet.Add(color.Bold)
	colorSet = colorSet.Add(color.BgCyan)
	colorSet.Printf(`

  __    __     __     __    __
  /\ "-./  \   /\ \   /\ "-./  \
  \ \ \-./\ \  \ \ \  \ \ \-./\ \
   \ \_\ \ \_\  \ \_\  \ \_\ \ \_\
    \/_/  \/_/   \/_/   \/_/  \/_/

     Mim Runtime Environment.
     App version: %#s
     Protocol version: %#s

`, fmt.Sprintf(
		"v%d.%d.%d",
		globals.BackendConfig.GetClientVersionMajor(),
		globals.BackendConfig.GetClientVersionMinor(),
		globals.BackendConfig.GetClientVersionPatch()),
		fmt.Sprintf(
			"v%d.%d",
			globals.BackendConfig.GetProtocolVersionMajor(),
			globals.BackendConfig.GetProtocolVersionMinor()))
}

// Start methods for Mim.

// establishConfigs establishes the configs (both transient and permanent) based on the flags provided through the CLI. Some flags are available directly (if they're local variables) and some of them are saved into transient config, made available globally until the app quits. If you need to have the data that is provided by the flag used in multiple places, create a new item in the transient config and put it there, it will be made available to the whole app. If you need the value once (ex: inserting a value into a database) then you can just use the value from the flags struct.
func establishConfigs(cmd *cobra.Command) flags {
	flgs := renderFlags(cmd)
	// Transient configs are established before permanent (saved to file) configs because appname and orgname in the transient configs determine where permanent configs get saved to. This is useful when running swarm tests, because specifying these effectively makes a swarm spawn save configs and caches into a different location than what it would normally not save.
	globals.BackendTransientConfig = &configstore.Btc
	globals.BackendTransientConfig.SetDefaults()
	globals.FrontendTransientConfig = &configstore.Ftc
	globals.FrontendTransientConfig.SetDefaults()
	// Set the transient config data.
	globals.BackendTransientConfig.AppIdentifier = flgs.appName.value.(string)
	globals.BackendTransientConfig.OrgIdentifier = flgs.orgName.value.(string)
	// Establish permanent configs.
	becfg, err := configstore.EstablishBackendConfig()
	if err != nil {
		logging.LogCrash(err)
	}
	becfg.Cycle()
	globals.BackendConfig = becfg
	fecfg, err := configstore.EstablishFrontendConfig()
	if err != nil {
		logging.LogCrash(err)
	}
	fecfg.Cycle()
	globals.FrontendConfig = fecfg
	// Determine whether the configs have been manipulated by flags. If so, disable editing of permanent configs for this session.
	if flagsChanged(cmd) {
		globals.BackendTransientConfig.PermConfigReadOnly = true
		globals.FrontendTransientConfig.PermConfigReadOnly = true
	}
	// Start setting permanent configs. These are NO-OPs if the permament config is read only.
	globals.BackendConfig.SetLoggingLevel(flgs.loggingLevel.value.(int))
	// If the permanent config is read only, we probably should tell.
	if globals.BackendTransientConfig.PermConfigReadOnly {
		// This is double because we want to both print on the console, and have it in the logs. Also providing even the default value explicitly through the command line triggers a changed=true, so even if you do logginglevel=0 (whose default is also 0), the configs will end up read only.
		fmt.Println("Configuration read only. Configuration for both backend and the frontend will be treated as read only because command line flags were provided for this run.")
		logging.Log(1, fmt.Sprint("Configuration read only. Configuration for both backend and the frontend will be treated as read only because command line flags were provided for this run."))
	}
	if flgs.port.changed {
		globals.BackendConfig.SetExternalPort(flgs.port.value.(int))
	}
	if flgs.externalIp.changed {
		globals.BackendConfig.SetExternalIp(flgs.externalIp.value.(string))
	}
	globals.BackendTransientConfig.PrintToStdout = flgs.printToStdout.value.(bool)
	// Set up the DB Instance so that we get access to the database.
	if globals.BackendConfig.GetDbEngine() == "sqlite" {
		globals.DbInstance = sqlx.MustConnect(
			"sqlite3",
			fmt.Sprintf(
				"%s/AetherDB.db", globals.BackendConfig.GetUserDirectory()))
	} else if globals.BackendConfig.GetDbEngine() == "mysql" {
		// If you want to use the MySQL, create a 'AetherDB' in your MySQL instance and insert the username / password here.
		/*
			MySQL connection string syntax:
			root:PASSWORD@tcp(l:3306)/sqlx_test
		*/
		mysqlConnectionString := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/aetherdb",
			globals.BackendConfig.GetDbUsername(),
			globals.BackendConfig.GetDbPassword(),
			globals.BackendConfig.GetDbIp(),
			globals.BackendConfig.GetDbPort())
		fmt.Println(mysqlConnectionString)
		globals.DbInstance = sqlx.MustConnect("mysql", mysqlConnectionString)
	} else {
		logging.LogCrash(fmt.Sprintf("Storage engine you've inputted is not supported. Please change it from the backend user config into something that is supported. You've provided: %s", globals.BackendConfig.GetDbEngine()))
	}
	return flgs
}

// Instructions: If you're adding a flag anywhere in the app, please add it to flags struct, render flags, and flagschanged.

type flag struct {
	value   interface{}
	changed bool
}

// Struct for flags. When there's a new flag, add it here.
type flags struct {
	loggingLevel  flag // int
	orgName       flag // string
	appName       flag // string
	port          flag // int
	externalIp    flag // string
	bootstrapIp   flag
	bootstrapPort flag
	bootstrapType flag
	syncAndQuit   flag
	printToStdout flag
	// Flags will be all lowercase in terminal input, heads up.
}

// When there's a new flag, add the parsing logic here underneath.
// I'm aware that this one sets the changed field, and yet, there's another method to check changed fields underneath that doesn't use this. It's because without using reflect package Go doesn't allow iteration over struct fields, and reflect, when used, does slow things down.
func renderFlags(cmd *cobra.Command) flags {
	var fl flags
	ll, err := cmd.Flags().GetInt("logginglevel")
	if err != nil && !strings.Contains(err.Error(), "flag accessed but not defined") {
		logging.LogCrash(err)
	}
	fl.loggingLevel.value = ll
	fl.loggingLevel.changed = cmd.Flags().Changed("logginglevel")
	on, err2 := cmd.Flags().GetString("orgname")
	if err2 != nil && !strings.Contains(
		err2.Error(), "flag accessed but not defined") {
		logging.LogCrash(err2)
	}
	fl.orgName.value = on
	fl.orgName.changed = cmd.Flags().Changed("orgname")
	an, err3 := cmd.Flags().GetString("appname")
	if err3 != nil && !strings.Contains(
		err3.Error(), "flag accessed but not defined") {
		logging.LogCrash(err3)
	}
	fl.appName.value = an
	fl.appName.changed = cmd.Flags().Changed("appname")

	p, err4 := cmd.Flags().GetInt("port")
	if err4 != nil && !strings.Contains(
		err4.Error(), "flag accessed but not defined") {
		logging.LogCrash(err4)
	}
	fl.port.value = p
	fl.port.changed = cmd.Flags().Changed("port")

	ei, err5 := cmd.Flags().GetString("externalip")
	if err5 != nil && !strings.Contains(
		err5.Error(), "flag accessed but not defined") {
		logging.LogCrash(err5)
	}
	fl.externalIp.value = ei
	fl.externalIp.changed = cmd.Flags().Changed("externalip")

	bi, err6 := cmd.Flags().GetString("bootstrapip")
	if err6 != nil && !strings.Contains(
		err6.Error(), "flag accessed but not defined") {
		logging.LogCrash(err6)
	}
	fl.bootstrapIp.value = bi
	fl.bootstrapIp.changed = cmd.Flags().Changed("bootstrapip")

	bp, err7 := cmd.Flags().GetInt("bootstrapport")
	if err7 != nil && !strings.Contains(
		err7.Error(), "flag accessed but not defined") {
		logging.LogCrash(err7)
	}
	fl.bootstrapPort.value = bp
	fl.bootstrapPort.changed = cmd.Flags().Changed("bootstrapport")

	bt, err8 := cmd.Flags().GetInt("bootstraptype")
	if err8 != nil && !strings.Contains(
		err8.Error(), "flag accessed but not defined") {
		logging.LogCrash(err8)
	}
	fl.bootstrapType.value = bt
	fl.bootstrapType.changed = cmd.Flags().Changed("bootstraptype")

	se, err9 := cmd.Flags().GetBool("syncandquit")
	if err9 != nil && !strings.Contains(
		err9.Error(), "flag accessed but not defined") {
		logging.LogCrash(err9)
	}
	fl.syncAndQuit.value = se
	fl.syncAndQuit.changed = cmd.Flags().Changed("syncandquit")

	so, err10 := cmd.Flags().GetBool("printtostdout")
	if err10 != nil && !strings.Contains(
		err10.Error(), "flag accessed but not defined") {
		logging.LogCrash(err10)
	}
	fl.printToStdout.value = so
	fl.printToStdout.changed = cmd.Flags().Changed("printtostdout")

	return fl
}

// When there's a new flag, add it underneath so that it'll be checked if a value was provided. If it is, we want to disable the writes.
func flagsChanged(cmd *cobra.Command) bool {
	var result bool
	changeChecker := func(flag *pflag.Flag) {
		if flag.Changed {
			result = true
		}
	}
	cmd.Flags().VisitAll(changeChecker)
	return result

	// ... For every flag, we need this, because if a flag is given we need to stop writing to config store file, and only keep the config store object in memory.

	// What that means is that if you provide ANY flags, the app won't commit ANYTHING to the file - not just the flag you set, but anything else, too. It will effectively operate in read-only mode in terms of configuration. This read-only mode will activate only after the init of the configstore is complete, so it does not prevent initial creation or fixing of missing values.
	return false
}
