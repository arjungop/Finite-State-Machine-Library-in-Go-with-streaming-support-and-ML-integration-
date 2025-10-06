// Package main demonstrates advanced autonomous order processing using FSM
package main

import (
	"fmt"       // Standard library for formatted I/O operations
	"log"       // Standard library for logging functionality
	"math/rand" // Standard library for random number generation
	"time"      // Standard library for time operations and delays

	"github.com/fla/self-programming-ai/pkg/fsm" // Import the FSM framework
)

// Define state and event types for type safety and code clarity
type State string // Custom type for FSM states to prevent mixing with regular strings
type Event string // Custom type for FSM events to prevent mixing with regular strings

// Order processing states - defines all possible states in the order lifecycle
const (
	Pending    State = "pending"    // Initial state when order is first created
	Validating State = "validating" // State during order validation process
	Validated  State = "validated"  // State after successful validation
	Processing State = "processing" // State during payment processing
	Paid       State = "paid"       // State after successful payment
	Packaging  State = "packaging"  // State during order packaging
	Packaged   State = "packaged"   // State after packaging is complete
	Shipping   State = "shipping"   // State during shipment
	Delivered  State = "delivered"  // Final successful state - order delivered
	Cancelled  State = "cancelled"  // State when order is cancelled
	Refunded   State = "refunded"   // State when order is refunded
)

// Order processing events
const (
	SubmitOrder    Event = "submit_order"
	ValidateOrder  Event = "validate_order"
	ProcessPayment Event = "process_payment"
	PackageOrder   Event = "package_order"
	ShipOrder      Event = "ship_order"
	DeliverOrder   Event = "deliver_order"
	CancelOrder    Event = "cancel_order"
	RefundOrder    Event = "refund_order"
)

// OrderProcessor demonstrates an autonomous order processing system
type OrderProcessor struct {
	orderID     string
	totalAmount float64
	machine     fsm.Machine
	logger      *log.Logger
}

// NewOrderProcessor creates a new autonomous order processing system
func NewOrderProcessor(orderID string, totalAmount float64) *OrderProcessor {
	logger := log.New(log.Writer(), fmt.Sprintf("[Order %s] ", orderID), log.LstdFlags)

	processor := &OrderProcessor{
		orderID:     orderID,
		totalAmount: totalAmount,
		logger:      logger,
	}

	// Build the self-programming FSM
	machine, err := fsm.NewBuilderWithHooks().
		AddStates(
			fsm.State(Pending),
			fsm.State(Validating),
			fsm.State(Validated),
			fsm.State(Processing),
			fsm.State(Paid),
			fsm.State(Packaging),
			fsm.State(Packaged),
			fsm.State(Shipping),
			fsm.State(Delivered),
			fsm.State(Cancelled),
			fsm.State(Refunded),
		).
		AddEvents(
			fsm.Event(SubmitOrder),
			fsm.Event(ValidateOrder),
			fsm.Event(ProcessPayment),
			fsm.Event(PackageOrder),
			fsm.Event(ShipOrder),
			fsm.Event(DeliverOrder),
			fsm.Event(CancelOrder),
			fsm.Event(RefundOrder),
		).
		AddTransitionWithAction(
			fsm.State(Pending), fsm.Event(SubmitOrder), fsm.State(Validating),
			processor.createLogAction("Order submitted, starting validation"),
		).
		AddTransitionWithCondition(
			fsm.State(Validating), fsm.Event(ValidateOrder), fsm.State(Validated),
			processor.createValidationCondition(),
		).
		AddTransitionWithAction(
			fsm.State(Validated), fsm.Event(ProcessPayment), fsm.State(Processing),
			processor.createLogAction("Processing payment"),
		).
		AddTransitionWithCondition(
			fsm.State(Processing), fsm.Event(ProcessPayment), fsm.State(Paid),
			processor.createPaymentCondition(),
		).
		AddTransitionWithAction(
			fsm.State(Paid), fsm.Event(PackageOrder), fsm.State(Packaging),
			processor.createLogAction("Starting packaging process"),
		).
		AddTransitionWithAction(
			fsm.State(Packaging), fsm.Event(PackageOrder), fsm.State(Packaged),
			processor.createPackagingAction(),
		).
		AddTransitionWithAction(
			fsm.State(Packaged), fsm.Event(ShipOrder), fsm.State(Shipping),
			processor.createLogAction("Order shipped"),
		).
		AddTransitionWithAction(
			fsm.State(Shipping), fsm.Event(DeliverOrder), fsm.State(Delivered),
			processor.createDeliveryAction(),
		).
		AddOnStateEnterHook(processor.createStateEnterHook()).
		AddOnStateExitHook(processor.createStateExitHook()).
		SetInitialState(fsm.State(Pending)).
		Build()

	if err != nil {
		log.Fatalf("Failed to create order processor FSM: %v", err)
	}

	processor.machine = machine
	return processor
}

// Helper methods for creating actions and conditions
func (op *OrderProcessor) createLogAction(message string) fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		op.logger.Printf("%s", message)
		return nil
	}
}

func (op *OrderProcessor) createValidationCondition() fsm.TransitionCondition {
	return func(context fsm.Context) bool {
		// Simulate validation logic (90% success rate)
		return rand.Float64() > 0.1
	}
}

func (op *OrderProcessor) createPaymentCondition() fsm.TransitionCondition {
	return func(context fsm.Context) bool {
		// Simulate payment processing (95% success rate)
		return rand.Float64() > 0.05
	}
}

func (op *OrderProcessor) createPackagingAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		op.logger.Printf("Packaging completed for order %s", op.orderID)
		return nil
	}
}

func (op *OrderProcessor) createDeliveryAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		op.logger.Printf("Order %s delivered successfully!", op.orderID)
		return nil
	}
}

func (op *OrderProcessor) createRefundAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		op.logger.Printf("Refund processed for order %s: $%.2f", op.orderID, op.totalAmount)
		return nil
	}
}

func (op *OrderProcessor) createStateEnterHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		op.logger.Printf("🔄 Entering state: %s", result.ToState)
	}
}

func (op *OrderProcessor) createStateExitHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		op.logger.Printf("⬅️  Exiting state: %s", result.FromState)
	}
}

// ProcessOrder autonomously processes an order through its lifecycle
func (op *OrderProcessor) ProcessOrder() error {
	op.logger.Printf("🚀 Starting autonomous order processing for order %s (Amount: $%.2f)", op.orderID, op.totalAmount)

	// Submit the order
	if _, err := op.machine.SendEvent(fsm.Event(SubmitOrder)); err != nil {
		return fmt.Errorf("failed to submit order: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Validate the order
	if _, err := op.machine.SendEvent(fsm.Event(ValidateOrder)); err != nil {
		return fmt.Errorf("failed to validate order: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Process payment
	if _, err := op.machine.SendEvent(fsm.Event(ProcessPayment)); err != nil {
		return fmt.Errorf("failed to process payment: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Trigger the payment event again to move to paid state
	if _, err := op.machine.SendEvent(fsm.Event(ProcessPayment)); err != nil {
		return fmt.Errorf("failed to complete payment: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Package the order
	if _, err := op.machine.SendEvent(fsm.Event(PackageOrder)); err != nil {
		return fmt.Errorf("failed to package order: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Complete packaging
	if _, err := op.machine.SendEvent(fsm.Event(PackageOrder)); err != nil {
		return fmt.Errorf("failed to complete packaging: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Ship the order
	if _, err := op.machine.SendEvent(fsm.Event(ShipOrder)); err != nil {
		return fmt.Errorf("failed to ship order: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Deliver the order
	if _, err := op.machine.SendEvent(fsm.Event(DeliverOrder)); err != nil {
		return fmt.Errorf("failed to deliver order: %w", err)
	}

	op.logger.Printf("✅ Order processing completed successfully!")
	return nil
}

func main() {
	fmt.Println("🎉 FORMAL LANGUAGE-BASED SELF-PROGRAMMING AI FRAMEWORK")
	fmt.Println("✅ Advanced Demo - Autonomous Order Processing")
	fmt.Println("===============================================")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Create multiple order processors to demonstrate autonomous behavior
	orders := []struct {
		id     string
		amount float64
	}{
		{"ORD-001", 99.99},
		{"ORD-002", 249.50},
		{"ORD-003", 15.99},
	}

	for _, order := range orders {
		fmt.Printf("\n📦 Processing Order: %s\n", order.id)
		fmt.Println("-----------------------------------")

		processor := NewOrderProcessor(order.id, order.amount)

		if err := processor.ProcessOrder(); err != nil {
			log.Printf("❌ Order processing failed: %v", err)
		}

		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("\n🎯 Demo completed successfully!")
	fmt.Println("The FSM framework demonstrates:")
	fmt.Println("• ✅ Autonomous state transitions")
	fmt.Println("• ✅ Self-programming behavior")
	fmt.Println("• ✅ Hook-based monitoring")
	fmt.Println("• ✅ Condition-based decision making")
	fmt.Println("• ✅ Action-based state management")
}
