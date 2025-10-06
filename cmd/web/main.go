// Package main provides a web interface for FSM visualization and interaction
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fla/self-programming-ai/pkg/fsm"
	"github.com/fla/self-programming-ai/pkg/web"
)

// createDemoMachine builds a simple demo FSM and returns it
func createDemoMachine() fsm.Machine {
	builder := fsm.NewBuilderWithHooks()

	builder.AddStates(
		fsm.State("pending"),
		fsm.State("processing"),
		fsm.State("completed"),
		fsm.State("cancelled"),
	)

	builder.AddEvents(
		fsm.Event("submit"),
		fsm.Event("process"),
		fsm.Event("complete"),
		fsm.Event("cancel"),
	)

	builder.AddTransitionWithAction(
		fsm.State("pending"), fsm.Event("submit"), fsm.State("processing"),
		func(from, to fsm.State, event fsm.Event, ctx fsm.Context) error {
			ctx.Set("submitted_at", time.Now().Format(time.RFC3339))
			return nil
		},
	)

	builder.AddTransitionWithAction(
		fsm.State("processing"), fsm.Event("complete"), fsm.State("completed"),
		func(from, to fsm.State, event fsm.Event, ctx fsm.Context) error {
			ctx.Set("completed_at", time.Now().Format(time.RFC3339))
			return nil
		},
	)

	builder.AddTransition(fsm.State("processing"), fsm.Event("cancel"), fsm.State("cancelled"))

	builder.AddAfterTransitionHook(func(result fsm.TransitionResult, ctx fsm.Context) {
		fmt.Printf("✨ Transition: %s -> %s via %s (success=%v)\n", 
			result.FromState, result.ToState, result.Event, result.Success)
	})

	builder.SetInitialState(fsm.State("pending"))

	machine, err := builder.Build()
	if err != nil {
		log.Fatalf("Failed to build demo machine: %v", err)
	}
	return machine
}

func main() {
	// Get port from environment or default to 8080
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	fmt.Printf("� Starting FSM Web Server on port %d\n", port)

	server := web.NewAdvancedVisualizationServer(port)

	// Register demo machine
	demoMachine := createDemoMachine()
	server.RegisterMachine("demo-order", demoMachine)
	fmt.Println("📦 Demo machine registered")

	fmt.Printf("🌐 Open: http://localhost:%d\n", port)
	fmt.Printf("📊 API:  http://localhost:%d/api/machines\n", port)

	log.Fatal(server.Start())
}
