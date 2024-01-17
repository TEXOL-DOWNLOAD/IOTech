#!/bin/sh

execute_command_until_success(){
  max_attempts="$1"
  shift
  expect_resp="$1"
  shift
  cmd="$@"
  attempts=0
  cmd_status=1
  cmd_resp=""
  until [ $cmd_status -eq 0 ] && [ "$cmd_resp" = "$expect_resp" ]
  do

    if [ ${attempts} -eq ${max_attempts} ];then
      echo "max attempts reached"
      break
    elif [ ${attempts} -ne 0 ]; then
      sleep 5s
    fi

    cmd_resp=$($cmd)
    cmd_status=$?
    attempts=$(($attempts+1))

	echo "   cmd_status: $cmd_status, cmd_resp: $cmd_resp, attempts: $attempts"

  done
  echo "   execute command successfully"
}

register_appservice() {
    local API_ENDPOINT="$1"
    local GUI_ACCESS_TOKEN="$2"
    local POST_DATA="$3"
    echo "Create AppService"
    curl -X POST \
         -H "Authorization: Bearer $GUI_ACCESS_TOKEN" \
         -H "Content-Type: application/json" \
         -d $POST_DATA \
         $API_ENDPOINT
}


# Set SENSOR_TYPE and TEST as environment variables
export SENSOR_TYPE=$1
export TEST=$2

# Accept two or three arguments: SENSOR_TYPE, and optional TEST
if [ -z "$1" ]; then
  echo "Error: Not enough arguments supplied."
  echo "Usage: $0 SENSOR_TYPE [TEST]"
  echo "Available SENSOR_TYPE options: modbus / mqtt / both"
  echo ""
  echo "The TEST variable can only have two values: --test or an empty string. It is used to enable the test mode when set to --test, while leaving it empty represents the product deployment."
  exit 1
fi


## Copy resource to file system
sudo mkdir -p /usr/share/texol/picture
sudo cp -rf config/picture/* /usr/share/texol/picture
sudo cp -rf config/GatewayIP.txt /usr/share/texol

if [ "$TEST" = "--test" ]; then
  echo "running test mode with eval compose file."
  export EDGEXPERT_PROJECT=eval
else
  export EDGEXPERT_PROJECT=texol
fi
echo "EDGEXPERT_PROJECT = $EDGEXPERT_PROJECT"

echo "Launching Edge Xpert with the required microservices..."
edgexpert pull core-keeper core-metadata core-data core-command mqtt-broker redis xpert-manager sys-mgmt influxdb grafana
edgexpert up xpert-manager sys-mgmt influxdb grafana

sleep 10s

## create buckets
./tools/influx_cmd.sh "bucket" "create -o texol -n Hourly_Bucket -r 730d"
sleep 1s
./tools/influx_cmd.sh "bucket" "create -o texol -n Daily_Bucket -r 730d"
sleep 1s

## create tasks
./tools/influx_cmd.sh "task" "./config/influxdb/daily_avg.json"
sleep 1s
./tools/influx_cmd.sh "task" "./config/influxdb/hourly_avg.json"
sleep 1s


##
## setup Grafana Dashboard
##
echo "Configuring the Grafana InfluxDB datasource..."
execute_command_until_success 2 200 curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type:application/json" -u admin:admin http://localhost:3000/api/datasources -d '{"name":"InfluxDB","type":"influxdb","isDefault":true,"url":"http://influxdb:8086","access":"proxy","basicAuth":false,"jsonData":{"organization":"texol","defaultBucket":"bucket","version":"Flux"},"secureJsonData":{"token":"texol-password"}}'



##
## setup App-service through Xpert Manager
##
API_ENDPOINT='http://localhost:9090/api/v2'
echo "Get Xpert Manager token by password."
GUI_TOKEN=$(curl -s --request POST --url ${API_ENDPOINT}/user/login --data '{"name":"admin","password":"admin"}')

GUI_ACCESS_TOKEN=$(echo $GUI_TOKEN | json_pp | grep access_token | cut -d'"' -f4)
# echo $GUI_ACCESS_TOKEN

INFLUXDB_APP_NAME='"influxdb-exporter"'
response=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $GUI_ACCESS_TOKEN" $API_ENDPOINT/appSvcConf/name/$INFLUXDB_APP_NAME)
if [ "$response" -eq 404 ]; then
  echo "No $INFLUXDB_APP_NAME found!, create it now"
  ## create Influxdb exporter
  POST_DATA='{"name":'${INFLUXDB_APP_NAME}',"logLevel":"INFO","destination":"InfluxDB","influxDBSyncWrite":{"influxDBServerURL":"http://influxdb:8086","influxDBOrganization":"texol","influxDBBucket":"bucket","token":"texol-password","influxDBMeasurement":"readings","fieldKeyPattern":"{resourceName}","influxDBValueType":"float","influxDBPrecision":"us","authMode":"token","secretPath":"influxdb","skipVerify":"true","persistOnError":"false","storeEventTags":"true","storeReadingTags":"false"}}'
  register_appservice "$API_ENDPOINT/appSvcConf" "$GUI_ACCESS_TOKEN" "$POST_DATA"

else
    echo "***The $INFLUXDB_APP_NAME exsit! Please remove it and try it again"
fi


echo "Uploading the device profile..."
echo "Sensor Type: $SENSOR_TYPE"
if [ "$SENSOR_TYPE" = "mqtt" ] || [ "$SENSOR_TYPE" = "both" ]; then
  # Start MQTT related services
  edgexpert up device-mqtt texol-broker texol-ble-driver

  # Upload MQTT device profile for BLE sensors
  execute_command_until_success 2 201 curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:59881/api/v2/deviceprofile/uploadfile -F "file=@config/deviceprofile/Texol_211HM1-B1_v2_10.yaml"
  sleep 1s
  execute_command_until_success 2 201 curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:59881/api/v2/deviceprofile/uploadfile -F "file=@config/deviceprofile/Texol_213MM1-B1_v2_10.yaml"
  sleep 1s
fi

if [ "$SENSOR_TYPE" = "modbus" ] || [ "$SENSOR_TYPE" = "both" ]; then
  # Start Modbus RTU related services
  edgexpert up device-modbus

  # Upload MQTT device profile for BLE sensors
  execute_command_until_success 2 201 curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:59881/api/v2/deviceprofile/uploadfile -F "file=@config/deviceprofile/Texol_213MM2-R1_v1_72.yaml"
  sleep 1s
fi


echo "Finish!"
exit 0
