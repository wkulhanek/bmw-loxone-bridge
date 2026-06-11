package names

import "strings"

// Known maps BMW REST API data point names to friendly keys.
// The BMW API returns names like "vehicle.drivetrain.electricEngine.charging.status"
// which autoFriendly converts to "drivetrain_electric_engine_charging_status".
// This map provides shorter aliases for the most useful ones.
var known = map[string]string{
	// Battery / Charging
	"vehicle.drivetrain.electricEngine.charging.level":                "battery_soc",
	"vehicle.drivetrain.electricEngine.charging.status":               "charging_status",
	"vehicle.powertrain.electric.battery.charging.power":              "charging_power_kw",
	"vehicle.drivetrain.electricEngine.charging.chargingMode":         "charging_mode",
	"vehicle.drivetrain.electricEngine.charging.connectionType":       "charging_connection_type",
	"vehicle.drivetrain.electricEngine.charging.connectorStatus":      "charging_connector_status",
	"vehicle.drivetrain.electricEngine.charging.timeRemaining":        "charging_time_remaining",
	"vehicle.drivetrain.electricEngine.charging.timeToFullyCharged":   "charging_time_to_full",
	"vehicle.drivetrain.electricEngine.remainingElectricRange":        "electric_range_km",
	"vehicle.drivetrain.electricEngine.kombiRemainingElectricRange":   "electric_range_display_km",
	"vehicle.drivetrain.electricEngine.hvsMaxEnergyAbsolute":          "battery_max_energy_kwh",
	"vehicle.drivetrain.batteryManagement.maxEnergy":                  "battery_max_energy",
	"vehicle.drivetrain.batteryManagement.batterySizeMax":             "battery_size_max",
	"vehicle.powertrain.electric.battery.stateOfChargeTarget":         "battery_soc_target",
	"vehicle.powertrain.electric.battery.stateOfHealthDisplayed":      "battery_health_percent",
	"vehicle.body.chargingPort.status":                                "charging_cable_status",
	"vehicle.body.chargingPort.lockedStatus":                          "charging_cable_lock",

	// Location
	"vehicle.location.lat":     "location_lat",
	"vehicle.location.lon":     "location_lon",
	"vehicle.location.heading": "location_heading",

	// Doors
	"vehicle.cabin.door.row1.driver.isOpen":     "door_driver_front",
	"vehicle.cabin.door.row1.passenger.isOpen":  "door_passenger_front",
	"vehicle.cabin.door.row2.driver.isOpen":     "door_driver_rear",
	"vehicle.cabin.door.row2.passenger.isOpen":  "door_passenger_rear",
	"vehicle.cabin.door.lockStatus":             "lock_state",
	"vehicle.cabin.door.status":                 "door_status",

	// Body
	"vehicle.body.hood.isOpen":     "hood",
	"vehicle.body.trunk.isOpen":    "tailgate",
	"vehicle.body.trunk.isLocked":  "tailgate_locked",

	// Windows
	"vehicle.cabin.window.row1.driver.isOpen":    "window_driver_front",
	"vehicle.cabin.window.row1.passenger.isOpen": "window_passenger_front",
	"vehicle.cabin.window.row2.driver.isOpen":    "window_driver_rear",
	"vehicle.cabin.window.row2.passenger.isOpen": "window_passenger_rear",

	// Fuel (PHEV)
	"vehicle.drivetrain.fuelSystem.level":          "fuel_level_percent",
	"vehicle.drivetrain.fuelSystem.remainingFuel":  "fuel_remaining_liters",
	"vehicle.drivetrain.totalRemainingRange":       "total_range_km",
	"vehicle.drivetrain.lastRemainingRange":        "last_remaining_range_km",

	// Odometer / Motion
	"vehicle.chassis.odometer":   "odometer_km",
	"vehicle.motion.state":       "motion_state",

	// 12V battery
	"vehicle.electricalSystem.battery.stateOfCharge":  "battery_12v_soc",
	"vehicle.electricalSystem.battery.voltage":        "battery_12v_voltage",
}

// Category prefixes for grouping endpoints.
// These match the auto-generated friendly names (after autoFriendly conversion).
var categories = map[string][]string{
	"battery": {
		"battery_",
		"charging_",
		"electric_range",
		"drivetrain_electric_engine_charging_",
		"drivetrain_electric_engine_remaining",
		"drivetrain_electric_engine_kombi",
		"drivetrain_electric_engine_hvs",
		"drivetrain_battery_management_",
		"powertrain_electric_battery_",
		"body_charging_port_",
	},
	"location": {
		"location_",
		"cabin_infotainment_navigation_current_location_",
	},
	"status": {
		"door_",
		"lock_",
		"hood",
		"tailgate",
		"window_",
		"cabin_door_",
		"cabin_window_",
		"body_hood_",
		"body_trunk_",
		"body_flap_",
	},
	"fuel": {
		"fuel_",
		"drivetrain_fuel_system_",
		"total_range",
		"last_remaining_range",
		"drivetrain_total_remaining_range",
		"drivetrain_last_remaining_range",
	},
}

func CategoryPrefixes(category string) []string {
	return categories[category]
}

func ToFriendly(bmwName string) string {
	if friendly, ok := known[bmwName]; ok {
		return friendly
	}
	return autoFriendly(bmwName)
}

func Category(bmwName string) string {
	for cat, prefixes := range categories {
		for _, prefix := range prefixes {
			if strings.HasPrefix(bmwName, prefix) {
				return cat
			}
		}
	}
	return ""
}

func autoFriendly(bmwName string) string {
	name := bmwName
	name = strings.TrimPrefix(name, "vehicle.")
	name = strings.ReplaceAll(name, ".", "_")
	return toSnakeCase(name)
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
