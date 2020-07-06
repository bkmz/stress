package main

import (
	"time"
	"fmt"
	"os"
	"net/http"
	"context"
	"os/signal"
	"syscall"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/justinas/alice"

	"github.com/bkmz/stress/utils"
	"github.com/bkmz/stress/config"
)

var (
	conf *config.Config
	chain alice.Chain
)

// func getCommands() [] cli.Command {
// 	// global level flags
// 	var cpuload float64
// 	var duration float64
// 	var cpucore int
// 	var context *cli.Context
// 	sampleInterval := 100 * time.Millisecond

// 	cpuLoadFlags := []cli.Flag{
// 		cli.Float64Flag{
// 			Name:  "cpuload",
// 			Usage: "Target CPU load 0<cpuload<1",
// 			Value: 0.1,
// 			Destination: &cpuload,
// 		},
// 		cli.Float64Flag{
// 			Name:  "duration",
// 			Usage: "Duration to run the stress app in Seconds",
// 			Value: 10,
// 			Destination: &duration,
// 		},
// 		cli.IntFlag{
// 			Name:  "cpucore",
// 			Usage: "Cpu core to stress ",
// 			Value: 0,
// 			Destination: &cpucore,
// 		},

// 	}
// 	commands :=[]cli.Command{
// 		{
// 			Name: "cpu",
// 			Action: func(c *cli.Context) {
// 				context = c
// 				runCpuLoader(sampleInterval, cpuload, duration, cpucore)
// 			},
// 			Usage: "load cpu , use --help for more options",
// 			Flags: cpuLoadFlags,
// 			Before: func(_ *cli.Context) error { return nil },
// 		},

// 	}
// 	return commands
// }

func runCpuLoader(sampleInterval time.Duration, cpuload float64, duration float64, cpu int) {
	log.Info().Msg("Start CPU load")
	controller := utils.NewCpuLoadController(sampleInterval, cpuload)
	monitor := utils.NewCpuLoadMonitor(float64(cpu), sampleInterval)

	actuator := utils.NewCpuLoadGenerator(controller, monitor, time.Duration(duration))
	utils.StartCpuLoadController(controller)
	utils.StartCpuMonitor(monitor)

	utils.RunCpuLoader(actuator)
	utils.StopCpuLoadController(controller)
	utils.StopCpuMonitor(monitor)
	log.Info().Msg("Stop CPU load")
}

func HelpInfo(w http.ResponseWriter, r *http.Request) {
	str := `
	Sample API for simulate cpu load

	Params:
	cpuload 	value   Target CPU load 0<cpuload<1 (default: 0.1)
	duration	value  Duration to run the stress app in Seconds (default: 10)
	cpucore		value   Cpu core to stress  (default: 0)
	`

	fmt.Fprintf(w, str)
} 


func CPULoad(w http.ResponseWriter, r *http.Request) {
	var (
		cpucore int
		err error
		cpuload float64
		duration float64
	)

	querys := r.URL.Query()

	core_key, ok := querys["cpucore"]
	if !ok || len(core_key[0]) < 1 {
		log.Error().Msg("Url Param 'cpucore' is missing, use default: 0")
		cpucore = 0
	} else {
		cpucore, err = strconv.Atoi(core_key[0])
		if err != nil {
			log.Error().Msgf("Error convert 'cpucore' params '%s' to int, use default: 0", core_key[0])
			cpucore = 0
		}
	}
	
	load_key, ok := querys["cpuload"]
	if !ok || len(load_key[0]) < 1 {
		log.Error().Msg("Url Param 'cpuload' is missing, use default: 0.1")
		cpuload = 0.1
	} else {
		cpuload, err = strconv.ParseFloat(load_key[0], 64)
		if err != nil {
			log.Error().Msgf("Error convert 'cpuload' params '%s' to float, use default: 0.1", load_key[0])
			cpuload = 0.1
		}
		if cpuload < 0 || cpuload > 1 {
			log.Error().Msgf("'cpuload' = '%s' is incorrect, use default", load_key[0])
			cpuload = 0.1
		}
	}

	duration_key, ok := querys["duration"]
	if !ok || len(duration_key[0]) < 1 {
		log.Error().Msg("Url Param 'duration' is missing, use default: 10")
		duration = 10
	} else {
		duration, err = strconv.ParseFloat(duration_key[0], 64)
		if err != nil {
			log.Error().Msgf("Error convert 'duration' params '%s' to float, use default: 10", duration_key[0])
		}
	}
	
	sampleInterval := 10 * time.Millisecond

	go runCpuLoader(sampleInterval, cpuload, duration, cpucore)
}

func init() {
	conf = config.Load()
}

func main(){
	mux := http.NewServeMux()
	listenStr := fmt.Sprintf("%s:%s", conf.ListenAddress, conf.ListenPort)

	server := http.Server{Addr: listenStr, Handler: mux}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM)

	mux.Handle("/help", chain.Then(http.HandlerFunc(HelpInfo)))
	mux.Handle("/cpu", chain.Then(http.HandlerFunc(CPULoad)))

	go func() {
		log.Info().Msgf("Start listen on ..%s", listenStr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msg(err.Error())
		}
	}()

	select {
	case <-signals:
		// Shutdown the server when the context is canceled
		server.Shutdown(ctx)
		log.Info().Msg("Recive shutdown signal")
	}
	
	log.Info().Msg("Finished")
}
