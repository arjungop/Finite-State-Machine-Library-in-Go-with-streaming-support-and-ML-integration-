package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fla/self-programming-ai/pkg/fsm"
)

// VendingMachine demonstrates an autonomous vending machine controller
// that exhibits self-programming behavior through formal state transitions
type VendingMachine struct {
	machine         fsm.Machine
	machineID       string
	inventory       map[string]int
	balance         float64
	selectedProduct string
	logger          *log.Logger
}

// VendingEvent represents events in the vending machine operation
type VendingEvent string

const (
	InsertCoin         VendingEvent = "insert_coin"
	SelectProduct      VendingEvent = "select_product"
	ConfirmPurchase    VendingEvent = "confirm_purchase"
	DispenseProduct    VendingEvent = "dispense_product"
	ReturnChange       VendingEvent = "return_change"
	CancelVendingOrder VendingEvent = "cancel_order"
	RefillMachine      VendingEvent = "refill_machine"
	ServiceMachine     VendingEvent = "service_machine"
)

// VendingState represents states in the vending machine system
type VendingState string

const (
	Idle            VendingState = "idle"
	CoinInserted    VendingState = "coin_inserted"
	ProductSelected VendingState = "product_selected"
	PaymentComplete VendingState = "payment_complete"
	Dispensing      VendingState = "dispensing"
	Dispensed       VendingState = "dispensed"
	ReturningChange VendingState = "returning_change"
	OutOfStock      VendingState = "out_of_stock"
	ServiceMode     VendingState = "service_mode"
	Error           VendingState = "error"
)

// Product represents a vending machine product
type Product struct {
	Name  string
	Price float64
	Code  string
}

// NewVendingMachine creates a new autonomous vending machine
func NewVendingMachine(machineID string) *VendingMachine {
	logger := log.New(log.Writer(), fmt.Sprintf("[Vending %s] ", machineID), log.LstdFlags)

	// Initialize inventory
	inventory := map[string]int{
		"A1": 10, // Soda - $1.50
		"A2": 8,  // Water - $1.00
		"B1": 5,  // Chips - $2.00
		"B2": 3,  // Candy - $1.25
		"C1": 0,  // Gum - $0.75 (out of stock)
	}

	vm := &VendingMachine{
		machineID: machineID,
		inventory: inventory,
		balance:   0.0,
		logger:    logger,
	}

	// Build the self-programming FSM using declarative rules
	machine, err := fsm.NewBuilderWithHooks().
		// Define vending machine states
		AddStates(
			fsm.State(Idle),
			fsm.State(CoinInserted),
			fsm.State(ProductSelected),
			fsm.State(PaymentComplete),
			fsm.State(Dispensing),
			fsm.State(Dispensed),
			fsm.State(ReturningChange),
			fsm.State(OutOfStock),
			fsm.State(ServiceMode),
			fsm.State(Error),
		).
		// Define vending machine events
		AddEvents(
			fsm.Event(InsertCoin),
			fsm.Event(SelectProduct),
			fsm.Event(ConfirmPurchase),
			fsm.Event(DispenseProduct),
			fsm.Event(ReturnChange),
			fsm.Event(CancelVendingOrder),
			fsm.Event(RefillMachine),
			fsm.Event(ServiceMachine),
		).
		// Define autonomous behavior through state transitions
		AddTransitionWithAction(
			fsm.State(Idle), fsm.Event(InsertCoin), fsm.State(CoinInserted),
			vm.createCoinInsertAction(),
		).
		AddTransitionWithCondition(
			fsm.State(CoinInserted), fsm.Event(SelectProduct), fsm.State(ProductSelected),
			vm.createProductSelectionCondition(),
		).
		AddTransitionWithCondition(
			fsm.State(CoinInserted), fsm.Event(SelectProduct), fsm.State(OutOfStock),
			vm.createOutOfStockCondition(),
		).
		AddTransitionWithCondition(
			fsm.State(ProductSelected), fsm.Event(ConfirmPurchase), fsm.State(PaymentComplete),
			vm.createPaymentCondition(),
		).
		AddTransitionWithAction(
			fsm.State(PaymentComplete), fsm.Event(DispenseProduct), fsm.State(Dispensing),
			vm.createDispenseAction(),
		).
		AddTransitionWithAction(
			fsm.State(Dispensing), fsm.Event(DispenseProduct), fsm.State(Dispensed),
			vm.createDispenseCompleteAction(),
		).
		AddTransitionWithCondition(
			fsm.State(Dispensed), fsm.Event(ReturnChange), fsm.State(ReturningChange),
			vm.createChangeRequiredCondition(),
		).
		AddTransitionWithAction(
			fsm.State(Dispensed), fsm.Event(ReturnChange), fsm.State(Idle),
			vm.createNoChangeAction(),
		).
		AddTransitionWithAction(
			fsm.State(ReturningChange), fsm.Event(ReturnChange), fsm.State(Idle),
			vm.createReturnChangeAction(),
		).
		// Add cancellation paths
		AddTransitionWithAction(
			fsm.State(CoinInserted), fsm.Event(CancelVendingOrder), fsm.State(ReturningChange),
			vm.createCancelAction(),
		).
		AddTransitionWithAction(
			fsm.State(ProductSelected), fsm.Event(CancelVendingOrder), fsm.State(ReturningChange),
			vm.createCancelAction(),
		).
		AddTransitionWithAction(
			fsm.State(OutOfStock), fsm.Event(CancelVendingOrder), fsm.State(ReturningChange),
			vm.createCancelAction(),
		).
		// Add service mode transitions
		AddTransition(
			fsm.State(Idle), fsm.Event(ServiceMachine), fsm.State(ServiceMode),
		).
		AddTransitionWithAction(
			fsm.State(ServiceMode), fsm.Event(RefillMachine), fsm.State(ServiceMode),
			vm.createRefillAction(),
		).
		AddTransition(
			fsm.State(ServiceMode), fsm.Event(ServiceMachine), fsm.State(Idle),
		).
		// Add autonomous monitoring hooks
		AddOnStateEnterHook(vm.createStateEnterHook()).
		AddAfterTransitionHook(vm.createTransitionHook()).
		AddOnTransitionErrorHook(vm.createErrorHook()).
		SetInitialState(fsm.State(Idle)).
		Build()

	if err != nil {
		panic(fmt.Sprintf("Failed to build vending machine FSM: %v", err))
	}

	vm.machine = machine

	// Initialize context with machine data
	context := vm.machine.GetContext()
	context.Set("machine_id", machineID)
	context.Set("total_sales", 0.0)
	context.Set("transaction_count", 0)
	context.Set("last_service", time.Now())

	return vm
}

// Product catalog
func (vm *VendingMachine) getProducts() map[string]Product {
	return map[string]Product{
		"A1": {"Soda", 1.50, "A1"},
		"A2": {"Water", 1.00, "A2"},
		"B1": {"Chips", 2.00, "B1"},
		"B2": {"Candy", 1.25, "B2"},
		"C1": {"Gum", 0.75, "C1"},
	}
}

// Autonomous behavior methods

// createCoinInsertAction creates an action for coin insertion
func (vm *VendingMachine) createCoinInsertAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		// Simulate coin value (random between quarters and dollars)
		coinValue := []float64{0.25, 0.50, 1.00}[time.Now().UnixNano()%3]
		vm.balance += coinValue

		context.Set("current_balance", vm.balance)
		vm.logger.Printf("Coin inserted: $%.2f (Total: $%.2f)", coinValue, vm.balance)
		return nil
	}
}

// createProductSelectionCondition creates a condition for valid product selection
func (vm *VendingMachine) createProductSelectionCondition() fsm.TransitionCondition {
	return func(context fsm.Context) bool {
		productCode := context.Get("selected_product")
		if productCode == nil {
			return false
		}

		code := productCode.(string)
		inventory := vm.inventory[code]

		// Product must be in stock
		if inventory <= 0 {
			return false
		}

		// Check if user has enough money
		products := vm.getProducts()
		if product, exists := products[code]; exists {
			return vm.balance >= product.Price
		}

		return false
	}
}

// createOutOfStockCondition creates a condition for out-of-stock products
func (vm *VendingMachine) createOutOfStockCondition() fsm.TransitionCondition {
	return func(context fsm.Context) bool {
		productCode := context.Get("selected_product")
		if productCode == nil {
			return false
		}

		code := productCode.(string)
		inventory := vm.inventory[code]
		return inventory <= 0
	}
}

// createPaymentCondition creates a condition for payment validation
func (vm *VendingMachine) createPaymentCondition() fsm.TransitionCondition {
	return func(context fsm.Context) bool {
		productCode := context.Get("selected_product").(string)
		products := vm.getProducts()

		if product, exists := products[productCode]; exists {
			sufficient := vm.balance >= product.Price
			if sufficient {
				context.Set("product_price", product.Price)
				context.Set("product_name", product.Name)
			}
			return sufficient
		}
		return false
	}
}

// createDispenseAction creates an action for product dispensing
func (vm *VendingMachine) createDispenseAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		productCode := context.Get("selected_product").(string)
		productPrice := context.Get("product_price").(float64)
		productName := context.Get("product_name").(string)

		vm.logger.Printf("Dispensing %s (Code: %s, Price: $%.2f)", productName, productCode, productPrice)

		// Deduct from balance
		vm.balance -= productPrice
		context.Set("current_balance", vm.balance)

		// Update sales tracking
		totalSales := context.Get("total_sales").(float64)
		context.Set("total_sales", totalSales+productPrice)

		transactionCount := context.Get("transaction_count").(int)
		context.Set("transaction_count", transactionCount+1)

		return nil
	}
}

// createDispenseCompleteAction creates an action for dispensing completion
func (vm *VendingMachine) createDispenseCompleteAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		productCode := context.Get("selected_product").(string)
		productName := context.Get("product_name").(string)

		// Update inventory
		vm.inventory[productCode]--
		context.Set("inventory_"+productCode, vm.inventory[productCode])

		vm.logger.Printf("Product dispensed: %s (Remaining: %d)", productName, vm.inventory[productCode])

		// Check if product is now out of stock
		if vm.inventory[productCode] == 0 {
			vm.logger.Printf("WARNING: Product %s is now out of stock", productCode)
		}

		return nil
	}
}

// createChangeRequiredCondition creates a condition to check if change is needed
func (vm *VendingMachine) createChangeRequiredCondition() fsm.TransitionCondition {
	return func(context fsm.Context) bool {
		return vm.balance > 0
	}
}

// createNoChangeAction creates an action when no change is needed
func (vm *VendingMachine) createNoChangeAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		vm.logger.Printf("Transaction complete - no change required")
		vm.resetTransaction(context)
		return nil
	}
}

// createReturnChangeAction creates an action for returning change
func (vm *VendingMachine) createReturnChangeAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		change := vm.balance
		vm.logger.Printf("Returning change: $%.2f", change)
		vm.balance = 0.0
		context.Set("current_balance", 0.0)
		vm.resetTransaction(context)
		return nil
	}
}

// createCancelAction creates an action for order cancellation
func (vm *VendingMachine) createCancelAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		vm.logger.Printf("Order cancelled - returning all money: $%.2f", vm.balance)
		vm.resetTransaction(context)
		return nil
	}
}

// createRefillAction creates an action for machine refilling
func (vm *VendingMachine) createRefillAction() fsm.TransitionAction {
	return func(from, to fsm.State, event fsm.Event, context fsm.Context) error {
		vm.logger.Printf("Refilling machine inventory")

		// Refill all products to maximum capacity
		for code := range vm.inventory {
			vm.inventory[code] = 10
			context.Set("inventory_"+code, 10)
		}

		context.Set("last_service", time.Now())
		vm.logger.Printf("Machine refilled successfully")
		return nil
	}
}

// resetTransaction resets transaction-specific data
func (vm *VendingMachine) resetTransaction(context fsm.Context) {
	vm.balance = 0.0
	vm.selectedProduct = ""
	context.Set("current_balance", 0.0)
	context.Set("selected_product", "")
	context.Set("product_price", 0.0)
	context.Set("product_name", "")
}

// Hook methods for autonomous monitoring

// createStateEnterHook creates a hook that monitors state entries
func (vm *VendingMachine) createStateEnterHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		vm.logger.Printf("State: %s", result.ToState)

		// Autonomous behavior based on state
		switch VendingState(result.ToState) {
		case Dispensing:
			// Auto-complete dispensing after delay
			go func() {
				time.Sleep(200 * time.Millisecond)
				vm.machine.SendEvent(fsm.Event(DispenseProduct))
			}()
		case Dispensed:
			// Auto-return change
			go func() {
				time.Sleep(100 * time.Millisecond)
				vm.machine.SendEvent(fsm.Event(ReturnChange))
			}()
		case OutOfStock:
			vm.logger.Printf("Product out of stock - please select another or cancel")
		}
	}
}

// createTransitionHook creates a hook that logs transitions
func (vm *VendingMachine) createTransitionHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		vm.logger.Printf("Transition: %s -> %s (Event: %s)",
			result.FromState, result.ToState, result.Event)
	}
}

// createErrorHook creates a hook that handles errors
func (vm *VendingMachine) createErrorHook() fsm.Hook {
	return func(result fsm.TransitionResult, context fsm.Context) {
		vm.logger.Printf("Error: %v", result.Error)

		// Autonomous error recovery
		if result.Error != nil {
			vm.logger.Printf("Attempting error recovery...")
		}
	}
}

// Public interface methods

// InsertCoin simulates coin insertion
func (vm *VendingMachine) InsertCoin() error {
	_, err := vm.machine.SendEvent(fsm.Event(InsertCoin))
	return err
}

// SelectProduct selects a product by code
func (vm *VendingMachine) SelectProduct(productCode string) error {
	context := vm.machine.GetContext()
	context.Set("selected_product", productCode)
	vm.selectedProduct = productCode

	products := vm.getProducts()
	if product, exists := products[productCode]; exists {
		vm.logger.Printf("Product selected: %s (Price: $%.2f)", product.Name, product.Price)
	}

	_, err := vm.machine.SendEvent(fsm.Event(SelectProduct))
	return err
}

// ConfirmPurchase confirms the purchase
func (vm *VendingMachine) ConfirmPurchase() error {
	_, err := vm.machine.SendEvent(fsm.Event(ConfirmPurchase))
	return err
}

// CancelTransaction cancels the current transaction
func (vm *VendingMachine) CancelTransaction() error {
	_, err := vm.machine.SendEvent(fsm.Event(CancelVendingOrder))
	return err
}

// ServiceMode enters service mode
func (vm *VendingMachine) ServiceMode() error {
	_, err := vm.machine.SendEvent(fsm.Event(ServiceMachine))
	return err
}

// RefillMachine refills the machine inventory
func (vm *VendingMachine) RefillMachine() error {
	_, err := vm.machine.SendEvent(fsm.Event(RefillMachine))
	return err
}

// GetCurrentState returns the current machine state
func (vm *VendingMachine) GetCurrentState() VendingState {
	return VendingState(vm.machine.CurrentState())
}

// GetStatus returns current machine status
func (vm *VendingMachine) GetStatus() map[string]interface{} {
	context := vm.machine.GetContext()
	status := context.GetAll()
	status["current_state"] = vm.GetCurrentState()
	status["current_balance"] = vm.balance
	status["inventory"] = vm.inventory
	return status
}

// GetInventory returns current inventory levels
func (vm *VendingMachine) GetInventory() map[string]int {
	inventory := make(map[string]int)
	for code, count := range vm.inventory {
		inventory[code] = count
	}
	return inventory
}

// Example usage function
func demonstrateVendingMachine() {
	fmt.Println("=== Autonomous Vending Machine Controller Demo ===")

	// Create autonomous vending machine
	vm := NewVendingMachine("VM-001")

	fmt.Println("Machine initialized. Available products:")
	products := vm.getProducts()
	for code, product := range products {
		inventory := vm.inventory[code]
		status := "Available"
		if inventory == 0 {
			status = "OUT OF STOCK"
		}
		fmt.Printf("  %s: %s - $%.2f (%s)\n", code, product.Name, product.Price, status)
	}

	fmt.Println("\n=== Transaction 1: Successful Purchase ===")
	// Insert coins
	vm.InsertCoin() // Random amount
	vm.InsertCoin() // Add more

	// Select and purchase product
	vm.SelectProduct("A2") // Water - $1.00
	time.Sleep(100 * time.Millisecond)
	vm.ConfirmPurchase()

	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Transaction 1 complete. State: %s\n", vm.GetCurrentState())

	fmt.Println("\n=== Transaction 2: Out of Stock ===")
	vm.InsertCoin()
	vm.SelectProduct("C1") // Gum - out of stock
	time.Sleep(100 * time.Millisecond)

	if vm.GetCurrentState() == OutOfStock {
		fmt.Println("Product out of stock, canceling...")
		vm.CancelTransaction()
	}

	time.Sleep(300 * time.Millisecond)
	fmt.Printf("Transaction 2 complete. State: %s\n", vm.GetCurrentState())

	fmt.Println("\n=== Service Mode Demo ===")
	vm.ServiceMode()
	vm.RefillMachine()
	vm.ServiceMode() // Exit service mode

	fmt.Printf("After service. Current inventory:\n")
	inventory := vm.GetInventory()
	for code, count := range inventory {
		fmt.Printf("  %s: %d units\n", code, count)
	}

	fmt.Println("\n=== Transaction 3: After Refill ===")
	vm.InsertCoin()
	vm.SelectProduct("C1") // Gum - now available
	time.Sleep(100 * time.Millisecond)
	vm.ConfirmPurchase()

	time.Sleep(500 * time.Millisecond)

	// Display final status
	fmt.Printf("\nFinal machine state: %s\n", vm.GetCurrentState())
	status := vm.GetStatus()
	fmt.Printf("Total sales: $%.2f\n", status["total_sales"])
	fmt.Printf("Transaction count: %d\n", status["transaction_count"])
}
