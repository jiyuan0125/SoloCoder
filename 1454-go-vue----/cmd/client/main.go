package main

import (
	"bufio"
	"canteen/internal/common"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	serverURL := os.Getenv("CANTEEN_SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	client := NewClient(serverURL)

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "help":
		printHelp()

	case "employee":
		err = handleEmployee(client, args)
	case "recharge":
		err = handleRecharge(client, args)

	case "dish":
		err = handleDish(client, args)

	case "consume":
		err = handleConsume(client, args)

	case "nutrition":
		err = handleNutrition(client, args)

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printHelp()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Canteen Management CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  canteen-cli <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  employee list                         List all employees")
	fmt.Println("  employee get <id>                     Get employee details")
	fmt.Println("  employee create                       Create a new employee (interactive)")
	fmt.Println()
	fmt.Println("  recharge <employee_id> <amount>       Recharge employee card (amount in cents)")
	fmt.Println("  recharge list <employee_id>           List recharge records")
	fmt.Println()
	fmt.Println("  dish list [--all]                     List dishes (only on-sale by default)")
	fmt.Println("  dish get <id>                         Get dish details")
	fmt.Println("  dish create                           Create a new dish (interactive)")
	fmt.Println("  dish update <id>                      Update dish (interactive)")
	fmt.Println("  dish onsale <id> <true|false>         Set dish on/off sale")
	fmt.Println()
	fmt.Println("  consume                               Start consumption session (interactive)")
	fmt.Println("  consume list <employee_id>            List consumption records")
	fmt.Println()
	fmt.Println("  nutrition <employee_id> [days]        Get nutrition summary (7 or 30 days)")
	fmt.Println()
}

func handleEmployee(client *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("employee subcommand required: list, get, create")
	}

	subcmd := args[0]
	switch subcmd {
	case "list":
		employees, err := client.ListEmployees()
		if err != nil {
			return err
		}
		fmt.Println("Employees:")
		for _, e := range employees {
			fmt.Printf("  ID: %s\n    Name: %s\n    Gender: %s\n    Weight: %dkg\n    Balance: %.2f\n    Credit: %.2f\n\n",
				e.ID, e.Name, e.Gender, e.Weight, float64(e.Balance)/100, float64(e.Credit)/100)
		}

	case "get":
		if len(args) < 2 {
			return fmt.Errorf("employee id required")
		}
		emp, err := client.GetEmployee(args[1])
		if err != nil {
			return err
		}
		fmt.Printf("ID: %s\n", emp.ID)
		fmt.Printf("Name: %s\n", emp.Name)
		fmt.Printf("Gender: %s\n", emp.Gender)
		fmt.Printf("Weight: %dkg\n", emp.Weight)
		fmt.Printf("Balance: %.2f\n", float64(emp.Balance)/100)
		fmt.Printf("Credit: %.2f\n", float64(emp.Credit)/100)
		fmt.Printf("Active: %v\n", emp.IsActive)

	case "create":
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		fmt.Print("Gender (male/female): ")
		genderStr, _ := reader.ReadString('\n')
		genderStr = strings.TrimSpace(genderStr)
		gender := common.Gender(strings.ToLower(genderStr))
		if gender != common.GenderMale && gender != common.GenderFemale {
			return fmt.Errorf("invalid gender")
		}

		fmt.Print("Weight (kg): ")
		weightStr, _ := reader.ReadString('\n')
		weight, err := strconv.Atoi(strings.TrimSpace(weightStr))
		if err != nil {
			return err
		}

		emp, err := client.CreateEmployee(common.CreateEmployeeRequest{
			Name:   name,
			Gender: gender,
			Weight: weight,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created employee: %s (ID: %s)\n", emp.Name, emp.ID)

	default:
		return fmt.Errorf("unknown employee subcommand: %s", subcmd)
	}
	return nil
}

func handleRecharge(client *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("recharge subcommand required: <employee_id> <amount> or list <employee_id>")
	}

	if args[0] == "list" {
		if len(args) < 2 {
			return fmt.Errorf("employee id required")
		}
		records, err := client.ListRechargeRecords(args[1])
		if err != nil {
			return err
		}
		fmt.Println("Recharge Records:")
		for _, r := range records {
			fmt.Printf("  ID: %s\n    Amount: %.2f\n    Before: %.2f\n    After: %.2f\n    Time: %s\n\n",
				r.ID, float64(r.Amount)/100, float64(r.Before)/100, float64(r.After)/100, r.CreatedAt)
		}
		return nil
	}

	if len(args) < 2 {
		return fmt.Errorf("amount required (in cents)")
	}
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	record, err := client.Recharge(common.RechargeRequest{
		EmployeeID: args[0],
		Amount:     amount,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Recharge successful!\n")
	fmt.Printf("Amount: %.2f\n", float64(record.Amount)/100)
	fmt.Printf("Before: %.2f\n", float64(record.Before)/100)
	fmt.Printf("After:  %.2f\n", float64(record.After)/100)
	return nil
}

func handleDish(client *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("dish subcommand required: list, get, create, update, onsale")
	}

	subcmd := args[0]
	switch subcmd {
	case "list":
		onlyOnSale := true
		if len(args) > 1 && args[1] == "--all" {
			onlyOnSale = false
		}
		dishes, err := client.ListDishes(onlyOnSale)
		if err != nil {
			return err
		}
		fmt.Println("Dishes:")
		for _, d := range dishes {
			status := "ON SALE"
			if !d.IsOnSale {
				status = "OFF SALE"
			}
			fmt.Printf("  ID: %s\n    Name: %s\n    Price: %.2f\n    Calories: %.1f kcal\n    Protein: %.1fg\n    Carbs: %.1fg\n    Fat: %.1fg\n    Status: %s\n\n",
				d.ID, d.Name, float64(d.Price)/100, d.Nutrition.Calories, d.Nutrition.Protein, d.Nutrition.Carbs, d.Nutrition.Fat, status)
		}

	case "get":
		if len(args) < 2 {
			return fmt.Errorf("dish id required")
		}
		dish, err := client.GetDish(args[1])
		if err != nil {
			return err
		}
		status := "ON SALE"
		if !dish.IsOnSale {
			status = "OFF SALE"
		}
		fmt.Printf("ID: %s\n", dish.ID)
		fmt.Printf("Name: %s\n", dish.Name)
		fmt.Printf("Price: %.2f\n", float64(dish.Price)/100)
		fmt.Printf("Calories: %.1f kcal\n", dish.Nutrition.Calories)
		fmt.Printf("Protein: %.1fg\n", dish.Nutrition.Protein)
		fmt.Printf("Carbs: %.1fg\n", dish.Nutrition.Carbs)
		fmt.Printf("Fat: %.1fg\n", dish.Nutrition.Fat)
		fmt.Printf("Status: %s\n", status)

	case "create":
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		fmt.Print("Price (in cents): ")
		priceStr, _ := reader.ReadString('\n')
		price, err := strconv.Atoi(strings.TrimSpace(priceStr))
		if err != nil {
			return err
		}

		fmt.Print("Calories (kcal): ")
		calStr, _ := reader.ReadString('\n')
		cal, _ := strconv.ParseFloat(strings.TrimSpace(calStr), 64)

		fmt.Print("Protein (g): ")
		protStr, _ := reader.ReadString('\n')
		prot, _ := strconv.ParseFloat(strings.TrimSpace(protStr), 64)

		fmt.Print("Carbs (g): ")
		carbsStr, _ := reader.ReadString('\n')
		carbs, _ := strconv.ParseFloat(strings.TrimSpace(carbsStr), 64)

		fmt.Print("Fat (g): ")
		fatStr, _ := reader.ReadString('\n')
		fat, _ := strconv.ParseFloat(strings.TrimSpace(fatStr), 64)

		dish, err := client.CreateDish(common.CreateDishRequest{
			Name:  name,
			Price: price,
			Nutrition: common.NutritionInfo{
				Calories: cal,
				Protein:  prot,
				Carbs:    carbs,
				Fat:      fat,
			},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created dish: %s (ID: %s)\n", dish.Name, dish.ID)

	case "update":
		if len(args) < 2 {
			return fmt.Errorf("dish id required")
		}
		reader := bufio.NewReader(os.Stdin)
		var req common.UpdateDishRequest

		fmt.Print("Name (leave empty to skip): ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name != "" {
			req.Name = &name
		}

		fmt.Print("Price in cents (leave empty to skip): ")
		priceStr, _ := reader.ReadString('\n')
		priceStr = strings.TrimSpace(priceStr)
		if priceStr != "" {
			price, err := strconv.Atoi(priceStr)
			if err != nil {
				return err
			}
			req.Price = &price
		}

		fmt.Println("Nutrition info (leave empty to skip):")
		var nutrition common.NutritionInfo
		hasNutrition := false

		fmt.Print("  Calories (kcal): ")
		calStr, _ := reader.ReadString('\n')
		calStr = strings.TrimSpace(calStr)
		if calStr != "" {
			cal, _ := strconv.ParseFloat(calStr, 64)
			nutrition.Calories = cal
			hasNutrition = true
		}

		fmt.Print("  Protein (g): ")
		protStr, _ := reader.ReadString('\n')
		protStr = strings.TrimSpace(protStr)
		if protStr != "" {
			prot, _ := strconv.ParseFloat(protStr, 64)
			nutrition.Protein = prot
			hasNutrition = true
		}

		fmt.Print("  Carbs (g): ")
		carbsStr, _ := reader.ReadString('\n')
		carbsStr = strings.TrimSpace(carbsStr)
		if carbsStr != "" {
			carbs, _ := strconv.ParseFloat(carbsStr, 64)
			nutrition.Carbs = carbs
			hasNutrition = true
		}

		fmt.Print("  Fat (g): ")
		fatStr, _ := reader.ReadString('\n')
		fatStr = strings.TrimSpace(fatStr)
		if fatStr != "" {
			fat, _ := strconv.ParseFloat(fatStr, 64)
			nutrition.Fat = fat
			hasNutrition = true
		}

		if hasNutrition {
			req.Nutrition = &nutrition
		}

		dish, err := client.UpdateDish(args[1], req)
		if err != nil {
			return err
		}
		fmt.Printf("Updated dish: %s\n", dish.Name)

	case "onsale":
		if len(args) < 3 {
			return fmt.Errorf("dish id and on_sale (true/false) required")
		}
		onSale := strings.ToLower(args[2]) == "true"
		if err := client.SetDishOnSale(args[1], onSale); err != nil {
			return err
		}
		status := "ON SALE"
		if !onSale {
			status = "OFF SALE"
		}
		fmt.Printf("Dish is now %s\n", status)

	default:
		return fmt.Errorf("unknown dish subcommand: %s", subcmd)
	}
	return nil
}

func handleConsume(client *Client, args []string) error {
	if len(args) > 0 && args[0] == "list" {
		if len(args) < 2 {
			return fmt.Errorf("employee id required")
		}
		records, err := client.ListConsumptionRecords(args[1])
		if err != nil {
			return err
		}
		fmt.Println("Consumption Records:")
		for _, r := range records {
			credit := ""
			if r.IsCredit {
				credit = " (CREDIT)"
			}
			fmt.Printf("  ID: %s\n    Meal: %s\n    Total: %.2f%s\n    Time: %s\n    Items:\n",
				r.ID, r.MealType, float64(r.TotalAmount)/100, credit, r.CreatedAt)
			for _, item := range r.Items {
				fmt.Printf("      %s x%d (%.2f)\n", item.Dish.Name, item.Quantity, float64(item.Dish.Price*item.Quantity)/100)
			}
			fmt.Println()
		}
		return nil
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Employee ID: ")
	empID, _ := reader.ReadString('\n')
	empID = strings.TrimSpace(empID)

	emp, err := client.GetEmployee(empID)
	if err != nil {
		return err
	}
	fmt.Printf("Employee: %s\n", emp.Name)
	fmt.Printf("Current Balance: %.2f\n", float64(emp.Balance)/100)
	fmt.Printf("Credit Used: %.2f / 100.00\n", float64(emp.Credit)/100)

	dishes, err := client.ListDishes(true)
	if err != nil {
		return err
	}
	fmt.Println("\nAvailable Dishes:")
	for i, d := range dishes {
		fmt.Printf("  [%d] %s - %.2f\n", i+1, d.Name, float64(d.Price)/100)
	}

	var items []common.DishItem
	total := 0

	for {
		fmt.Print("\nSelect dish number (0 to checkout): ")
		selStr, _ := reader.ReadString('\n')
		selStr = strings.TrimSpace(selStr)
		sel, _ := strconv.Atoi(selStr)

		if sel == 0 {
			break
		}
		if sel < 1 || sel > len(dishes) {
			fmt.Println("Invalid selection")
			continue
		}

		dish := dishes[sel-1]
		fmt.Print("Quantity: ")
		qtyStr, _ := reader.ReadString('\n')
		qty, err := strconv.Atoi(strings.TrimSpace(qtyStr))
		if err != nil || qty <= 0 {
			fmt.Println("Invalid quantity")
			continue
		}

		items = append(items, common.DishItem{DishID: dish.ID, Quantity: qty})
		total += dish.Price * qty
		fmt.Printf("Added: %s x%d\n", dish.Name, qty)
		fmt.Printf("Current Total: %.2f\n", float64(total)/100)
	}

	if len(items) == 0 {
		fmt.Println("No items selected")
		return nil
	}

	fmt.Println("\nSelect Meal Type:")
	fmt.Println("  [1] Breakfast")
	fmt.Println("  [2] Lunch")
	fmt.Println("  [3] Dinner")
	fmt.Println("  [4] Supper")
	fmt.Print("Selection: ")
	mealStr, _ := reader.ReadString('\n')
	mealSel, _ := strconv.Atoi(strings.TrimSpace(mealStr))

	var mealType common.MealType
	switch mealSel {
	case 1:
		mealType = common.MealBreakfast
	case 2:
		mealType = common.MealLunch
	case 3:
		mealType = common.MealDinner
	case 4:
		mealType = common.MealSupper
	default:
		mealType = common.MealLunch
	}

	record, err := client.Checkout(common.CheckoutRequest{
		EmployeeID: empID,
		Items:      items,
		MealType:   mealType,
	})
	if err != nil {
		return err
	}

	credit := ""
	if record.IsCredit {
		credit = " (on credit)"
	}
	fmt.Printf("\nCheckout successful!%s\n", credit)
	fmt.Printf("Total: %.2f\n", float64(record.TotalAmount)/100)
	return nil
}

func handleNutrition(client *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("employee id required")
	}

	days := 7
	if len(args) > 1 {
		d, err := strconv.Atoi(args[1])
		if err == nil {
			days = d
		}
	}

	summary, err := client.GetNutritionSummary(args[0], days)
	if err != nil {
		return err
	}

	fmt.Printf("Nutrition Summary (last %d days)\n", days)
	fmt.Printf("Period: %s to %s\n", summary.StartDate.Format("2006-01-02"), summary.EndDate.Format("2006-01-02"))
	fmt.Println()
	fmt.Printf("Calories:  %.1f kcal (Recommended: %.1f kcal)\n", summary.Calories, summary.Recommended)
	fmt.Printf("Protein:   %.1f g\n", summary.Protein)
	fmt.Printf("Carbs:     %.1f g\n", summary.Carbs)
	fmt.Printf("Fat:       %.1f g\n", summary.Fat)

	if summary.Recommended > 0 {
		ratio := summary.Calories / summary.Recommended * 100
		fmt.Printf("\nCalorie intake: %.1f%% of recommended\n", ratio)
	}
	return nil
}
