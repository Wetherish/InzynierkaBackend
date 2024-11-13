package main

import (
	"app/Commands"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	configFile = "config.json"
	mqttBroker = "mqtt:1883"
)

type Config struct {
	ID                int     `json:"ID"`
	Topic             string  `json:"Topic"`
	Message           string  `json:"Message"`
	AdditionalCommand bool    `json:"AdditionalCommand"`
	Time              string  `json:"Time"` // Format: "HH:MM"
	Longitude         float64 `json:"Longitude"`
	Latitude          float64 `json:"Latitude"`
	Command           string  `json:"Command"` //sunSet or sunRise
}

type Configurations struct {
	Configs []Config `json:"configs"`
}

var (
	mqttClient mqtt.Client
	configs    Configurations
)

func loadConfigs() error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	var tempcon Configurations
	if err := json.Unmarshal(data, &tempcon); err != nil {
		return err
	}
	for _, config := range tempcon.Configs {
		processConfig(&config)
	}
	configs.Configs = tempcon.Configs
	return nil
}

func processConfig(config *Config) error {
	var err error
	if config.AdditionalCommand {
		if config.Command == "sunSet" {
			config.Time, err = Commands.GetSunset(config.Latitude, config.Longitude)
			if err != nil {
				return err
			}
		}
		if config.Command == "sunRise" {
			config.Time, err = Commands.GetSunrise(config.Latitude, config.Longitude)
			if err != nil {
				return err
			}
		}
	}
	return nil

}

func saveConfigs() error {
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}

func publishMessage(config Config) error {
	token := mqttClient.Publish(config.Topic, 0, false, config.Message)
	token.Wait()
	return token.Error()
}

func getConfigHandler(c echo.Context) error {
	if err := loadConfigs(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load configurations"})
	}
	return c.JSON(http.StatusOK, configs.Configs)
}

func updateConfigHandler(c echo.Context) error {
	var newConfig Config
	if err := c.Bind(&newConfig); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}
	processConfig(&newConfig)

	for i, config := range configs.Configs {
		if config.ID == newConfig.ID {
			configs.Configs[i] = newConfig
			if err := saveConfigs(); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save configuration"})
			}
			return c.JSON(http.StatusOK, map[string]string{"message": "Configuration updated successfully"})
		}
	}

	configs.Configs = append(configs.Configs, newConfig)
	if err := saveConfigs(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save configuration"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Configuration added successfully"})
}

func clearConfigHandler(c echo.Context) error {
	configs.Configs = []Config{}
	if err := saveConfigs(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to clear configurations"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "All configurations cleared successfully"})
}

func updateConfig(c echo.Context) error {
	id := c.Param("id")

	var config Config
	if err := c.Bind(&config); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid configuration data"})
	}

	for i, cfg := range configs.Configs {
		if fmt.Sprintf("%d", cfg.ID) == id {
			configs.Configs[i] = config
			return c.JSON(http.StatusOK, config)
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"message": "Configuration not found"})
}

func initMQTT() mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.SetUsername("mobile")
	opts.SetPassword("lokomotywa")
	opts.AddBroker(mqttBroker)
	opts.SetClientID("go_mqtt_client")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}

	return client
}

func scheduleMessages() {
	for {
		now := time.Now().Format("15:04")
		log.Println(now)
		for _, config := range configs.Configs {
			if config.Time == now {
				log.Printf("Publishing message to topic %s at %s", config.Topic, now)
				if err := publishMessage(config); err != nil {
					log.Printf("Failed to publish message: %v", err)
				}
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func main() {
	if err := loadConfigs(); err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}
	fmt.Println("Check")
	mqttClient = initMQTT()

	go scheduleMessages()

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodDelete, http.MethodPut},
	}))

	e.GET("/config", getConfigHandler)
	e.POST("/config", updateConfigHandler)
	e.DELETE("/config", clearConfigHandler)
	e.PUT("/config/:id", updateConfig)

	e.Logger.Fatal(e.Start(":8080"))
}
