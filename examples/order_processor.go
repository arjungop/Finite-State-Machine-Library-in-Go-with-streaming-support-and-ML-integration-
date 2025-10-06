package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fla/self-programming-ai/pkg/fsm"
)

// OrderProcessor demonstrates an autonomous order processing system
// that exhibits self-programming behavior through state transitions
type OrderProcessor struct {
	machine      fsm.Machine
	orderID      string
	totalAmount  float64
	logger       *log.Logger
}

// OrderEvent represents events in the order processing lifecycle
type OrderEvent string

const (
	SubmitOrder    OrderEvent = "submit_order"
	ValidateOrder  OrderEvent = "validate_order"
	ProcessPayment OrderEvent = "process_payment"
	PackageOrder   OrderEvent = "package_order"
	ShipOrder      OrderEvent = "ship_order"
	DeliverOrder   OrderEvent = "deliver_order"
	CancelOrder    OrderEvent = "cancel_order"
	RefundOrder    OrderEvent = "refund_order"
)

// OrderState represents states in the order processing system
type OrderState string

const (
	Pending     OrderState = "pending"
	Validating  OrderState = "validating"
	Validated   OrderState = "validated"
	Processing  OrderState = "processing"
	Paid        OrderState = "paid"
	Packaging   OrderState = "packaging"
	Packaged    OrderState = "packaged"
	Shipping    OrderState = "shipping"
	Shipped     OrderState = "shipped"
	Delivered   OrderState = "delivered"
	Cancelled   OrderState = "cancelled"
	Refunded    OrderState = "refunded"
)

// NewOrderProcessor creates a new autonomous order processing system
func NewOrderProcessor(orderID string, totalAmount float64) *OrderProcessor {
	logger := log.New(log.Writer(), fmt.Sprintf("[Order %s] ", orderID), log.LstdFlags)
	
	processor := &OrderProcessor{
		orderID:     orderID,
		totalAmount: totalAmount,
		logger:      logger,
	}
	
	// Build the self-programming FSM using declarative rules
	machine, err := fsm.NewBuilderWithHooks().
		// Define the order lifecycle states
		AddStates(
			fsm.State(Pending),
			fsm.State(Validating),
			fsm.State(Validated),
			fsm.State(Processing),
			fsm.State(Paid),
			fsm.State(Packaging),
			fsm.State(Packaged),
			fsm.State(Shipping),
			fsm.State(Shipped),
			fsm.State(Delivered),
			fsm.State(Cancelled),
			fsm.State(Refunded),
		).
		// Define the order lifecycle events
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
		// Define the autonomous behavior through state transitions
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
		// Add cancellation paths (demonstrating adaptive behavior)
		AddTransitionWithAction(
			fsm.State(Pending), fsm.Event(CancelOrder), fsm.State(Cancelled),
			processor.createLogAction("Order cancelled before validation"),
		).
		AddTransitionWithAction(
			fsm.State(Validating), fsm.Event(CancelOrder), fsm.State(Cancelled),
			processor.createLogAction("Order cancelled during validation"),
		).
		AddTransitionWithAction(
			fsm.State(Validated), fsm.Event(CancelOrder), fsm.State(Cancelled),
			processor.createLogAction("Order cancelled after validation"),
		).
		// Add refund paths
		AddTransitionWithAction(
			fsm.State(Paid), fsm.Event(RefundOrder), fsm.State(Refunded),
			processor.createRefundAction(),
		).
		AddTransitionWithAction(
			fsm.State(Delivered), fsm.Event(RefundOrder), fsm.State(Refunded),
			processor.createRefundAction(),
		).
		// Add autonomous monitoring hooks
		AddOnStateEnterHook(processor.createStateEnterHook()).
		AddOnStateExitHook(processor.createStateExitHook()).
		AddOnTransitionErrorHook(processor.createErrorHook()).
		SetInitialState(fsm.State(Pending)).
		Build()
	
	if err != nil {
		panic(fmt.Sprintf("Failed to build order processor FSM: %v", err))
	}
	
	processor.machine = machine
	
	// Initialize context with order data
	context := processor.machine.GetContext()
	context.Set("order_id", orderID)
	context.Set("total_amount", totalAmount)
	context.Set("created_at", time.Now())
	context.Set("validation_attempts", 0)
	context.Set("payment_attempts", 0)
	
	return processor
}

// Autonomous behavior methods

// createValidationCondition creates a guard condition for order validation
func (op *OrderProcessor) createValidationCondition() fsm.TransitionCondition {
	return func(context fsm.Context) bool {
		attempts := context.Get("validation_attempts")
		if attempts == nil {
			attempts = 0
		}
		
		// Increment attempts
		context.Set("validation_attempts", attempts.(int)+1)
		
		// Simulate validation logic
		totalAmount := context.Get("total_amount").(float64)
		
		// Orders over $1000 require additional validation
		if totalAmount > 1000 {
			op.logger.Printf("High-value order requires additional validation (Amount: $%.2f)", totalAmount)
			return attempts.(int) >= 2 // Require 2 validation attempts for high-value orders
		}
		
		return true // Standard orders pass validation immediately
	}
}

// createPaymentCondition creates a guard condition for payment processing
func (op *OrderProcessor) createPaymentCondition() fsm.TransitionCondition {
	return func(context fsm.Context) bool {
		attempts := context.Get("payment_attempts")
		if attempts == nil {
			attempts = 0
		}
		
		// Increment attempts
		context.Set("payment_attempts", attempts.(int)+1)
		
		// Simulate payment processing (90% success rate)
		success := time.Now().UnixNano()%10 < 9
		
		if !success && attempts.(int) < 3 {
			op.logger.Printf("Payment attempt %d failed, retrying...", attempts.(int)+1)
			return false
		}
		
		if success {
			context.Set("payment_completed_at", time.Now())
			op.logger.Printf("Payment processed successfully")
		} else {
			op.logger.Printf("Payment failed after %d attempts", attempts.(int))
		}
		
		return success
	}
}

// createLogAction creates a transition action that logs a message
func (op *OrderProcessor) createLogAction(message string) fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		op.logger.Printf("%s (Transition: %s --%s--> %s)", message, from, event, to)
		return nil
	}
}

// createPackagingAction creates an action for the packaging process
func (op *OrderProcessor) createPackagingAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		op.logger.Printf("Packaging completed")
		context.Set("packaged_at", time.Now())
		context.Set("tracking_number", fmt.Sprintf("TRK%d", time.Now().UnixNano()%1000000))
		return nil
	}
}

// createDeliveryAction creates an action for delivery completion
func (op *OrderProcessor) createDeliveryAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		context.Set("delivered_at", time.Now())
		trackingNumber := context.Get("tracking_number")
		op.logger.Printf("Order delivered successfully (Tracking: %s)", trackingNumber)
		return nil
	}
}

// createRefundAction creates an action for processing refunds
func (op *OrderProcessor) createRefundAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		totalAmount := context.Get("total_amount").(float64)
		context.Set("refunded_at", time.Now())
		context.Set("refund_amount", totalAmount)
		op.logger.Printf("Refund processed: $%.2f", totalAmount)
		return nil
	}
}

// Hook methods for autonomous monitoring

// createStateEnterHook creates a hook that monitors state entries
func (op *OrderProcessor) createStateEnterHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		op.logger.Printf("Entered state: %s", result.ToState)
		
		// Autonomous behavior: automatically trigger next steps based on state
		switch OrderState(result.ToState) {
		case Validating:
			// Auto-trigger validation after brief delay
			go func() {
				time.Sleep(100 * time.Millisecond)
				op.machine.SendEvent(fsm.Event(ValidateOrder))
			}()
		case Validated:
			// Auto-trigger payment processing
			go func() {
				time.Sleep(50 * time.Millisecond)
				op.machine.SendEvent(fsm.Event(ProcessPayment))
			}()
		case Paid:
			// Auto-trigger packaging
			go func() {
				time.Sleep(200 * time.Millisecond)
				op.machine.SendEvent(fsm.Event(PackageOrder))
			}()
		case Packaged:
			// Auto-trigger shipping
			go func() {
				time.Sleep(100 * time.Millisecond)
				op.machine.SendEvent(fsm.Event(ShipOrder))
			}()
		}
	}
}

// createStateExitHook creates a hook that monitors state exits
func (op *OrderProcessor) createStateExitHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		op.logger.Printf("Exited state: %s", result.FromState)
	}
}

// createErrorHook creates a hook that handles transition errors
func (op *OrderProcessor) createErrorHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		op.logger.Printf("Transition error: %v", result.Error)
		
		// Autonomous error recovery
		if OrderState(result.FromState) == Processing {
			// Retry payment after delay
			go func() {
				time.Sleep(1 * time.Second)
				op.machine.SendEvent(fsm.Event(ProcessPayment))
			}()
		}
	}
}

// Public interface methods

// ProcessOrder starts the autonomous order processing
func (op *OrderProcessor) ProcessOrder() error {
	op.logger.Printf("Starting autonomous order processing")
	_, err := op.machine.SendEvent(fsm.Event(SubmitOrder))
	return err
}

// CancelOrder cancels the order if possible
func (op *OrderProcessor) CancelOrder() error {
	op.logger.Printf("Attempting to cancel order")
	_, err := op.machine.SendEvent(fsm.Event(CancelOrder))
	return err
}

// RefundOrder processes a refund if applicable
func (op *OrderProcessor) RefundOrder() error {
	op.logger.Printf("Attempting to process refund")
	_, err := op.machine.SendEvent(fsm.Event(RefundOrder))
	return err
}

// GetCurrentState returns the current order state
func (op *OrderProcessor) GetCurrentState() OrderState {
	return OrderState(op.machine.CurrentState())
}

// GetOrderDetails returns current order information
func (op *OrderProcessor) GetOrderDetails() map[string]interface{} {
	context := op.machine.GetContext()
	details := context.GetAll()
	details["current_state"] = op.GetCurrentState()
	return details
}

// SimulateDelivery simulates the delivery process
func (op *OrderProcessor) SimulateDelivery() error {
	// Wait for shipping state
	for op.GetCurrentState() != Shipping {
		time.Sleep(100 * time.Millisecond)
	}
	
	// Simulate delivery after shipping
	time.Sleep(500 * time.Millisecond)
	_, err := op.machine.SendEvent(fsm.Event(DeliverOrder))
	return err
}

// Example usage function
func demonstrateOrderProcessor() {
	fmt.Println("=== Autonomous Order Processing System Demo ===\n")
	
	// Create autonomous order processors
	order1 := NewOrderProcessor("ORD-001", 750.00)
	order2 := NewOrderProcessor("ORD-002", 1500.00) // High-value order
	
	// Start autonomous processing
	fmt.Println("Starting order processing...")
	order1.ProcessOrder()
	order2.ProcessOrder()
	
	// Simulate delivery for both orders
	go order1.SimulateDelivery()
	go order2.SimulateDelivery()
	
	// Wait for processing to complete
	time.Sleep(3 * time.Second)
	
	// Display final states
	fmt.Printf("\nFinal States:\n")
	fmt.Printf("Order 1: %s\n", order1.GetCurrentState())
	fmt.Printf("Order 2: %s\n", order2.GetCurrentState())
	
	// Demonstrate cancellation and refund
	fmt.Println("\n=== Testing Cancellation and Refund ===")
	order3 := NewOrderProcessor("ORD-003", 299.99)
	order3.ProcessOrder()
	
	time.Sleep(100 * time.Millisecond)
	if err := order3.CancelOrder(); err != nil {
		fmt.Printf("Cancel failed: %v\n", err)
	}
	
	// Try refund on a completed order
	time.Sleep(1 * time.Second)
	if order1.GetCurrentState() == Delivered {
		if err := order1.RefundOrder(); err != nil {
			fmt.Printf("Refund failed: %v\n", err)
		}
	}
	
	fmt.Printf("\nOrder 3 final state: %s\n", order3.GetCurrentState())
	fmt.Printf("Order 1 after refund: %s\n", order1.GetCurrentState())
}