#include <Arduino.h>
#include <ESP8266WiFi.h>
#include <AsyncMqttClient.h>

#define AIR_VALUE 1000.0 // REPLACE WITH YOUR VALUE
#define WATER_VALUE 0.0 // REPLACE WITH YOUR VALUE

#define WIFI_SSID "SSID_HERE"
#define WIFI_PASSWORD "PASSWD HERE"

#define MQTT_HOST IPAddress(192, 168, 0, 0)
#define MQTT_PORT 1883
#define MQTT_USER "USERNAME"
#define MQTT_PASSWD "PASSWORD"
#define MQTT_TOPIC "/plants/moisture/1"

WiFiEventHandler wifiConnectHandler;
WiFiEventHandler wifiDisconnectHandler;

AsyncMqttClient mqttClient;

bool readyToDisconnect = false;

void setReadyToDisconnect() {
  readyToDisconnect = true;
}

void disconnect() {
  ESP.deepSleep(300e6);
}

float getPercentage(int sensorReading) {
  return 1 - (sensorReading-WATER_VALUE)/(AIR_VALUE-WATER_VALUE);
}

void connectToWiFi() {
  Serial.println("Connecting to WiFi...");
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
}

void connectToMqtt() {
  Serial.println("Connecting to MQTT...");
  mqttClient.connect();
}

void onMqttConnect(bool sessionPresent) {
  Serial.println("Connected to MQTT Server");
  Serial.printf("Session present: %d\n", sessionPresent);

  int sensorReading = analogRead(A0); //put Sensor insert into soil
  float soilMoistureValue = getPercentage(sensorReading);
  char *string = (char*)malloc(13 * sizeof(char));
  sprintf(string, "%f", soilMoistureValue);
  uint16_t packetIdPub = mqttClient.publish(MQTT_TOPIC, 2, true, string); 
  Serial.printf("%s %d\n", string, packetIdPub);
  setReadyToDisconnect();
}

void onWifiConnect(const WiFiEventStationModeGotIP& event) {
  Serial.println("Connected to Wi-Fi.");
  connectToMqtt();
}

void onWifiDisconnect(const WiFiEventStationModeDisconnected& event) {
  Serial.println("Disconnected from Wi-Fi.");
}



void setup() {
  Serial.begin(9600); // open serial port, set the baud rate to 9600 bps

  wifiConnectHandler = WiFi.onStationModeGotIP(onWifiConnect);
  wifiDisconnectHandler = WiFi.onStationModeDisconnected(onWifiDisconnect);

  mqttClient.onConnect(onMqttConnect);
  mqttClient.setServer(MQTT_HOST, MQTT_PORT);
  mqttClient.setCredentials(MQTT_USER, MQTT_PASSWD);
  connectToWiFi();
}


void loop() {
  if (readyToDisconnect) {
    delay(1000);
    disconnect();
  }
}