#!/bin/sh
export EDGEXPERT_PROJECT=texol

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
      exit 1
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


## Copy resource to file system
sudo mkdir -p /usr/share/texol/picture
sudo cp -rf picture/* /usr/share/texol/picture
sudo cp -rf GatewayIP.txt /usr/share/texol

echo "Launching Edge Xpert with the required microservices..."
edgexpert up xpert-manager sys-mgmt device-modbus device-mqtt influxdb grafana texol-broker texol-ble-driver

# Sleep
sleep 10s

echo "Uploading the device profile..."
## TBD


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
  POST_DATA='{"name":'${INFLUXDB_APP_NAME}',"logLevel":"INFO","destination":"InfluxDB","influxDBSyncWrite":{"influxDBServerURL":"http://influxdb:8086","influxDBOrganization":"texol","influxDBBucket":"bucket","token":"texol-password","influxDBMeasurement":"readings","fieldKeyPattern":"{resourceName}","influxDBValueType":"float","influxDBPrecision":"us","authMode":"token","secretPath":"influxdb","skipVerify":"true","persistOnError":"false","storeEventTags":"false","storeReadingTags":"false"}}'
  register_appservice "$API_ENDPOINT/appSvcConf" "$GUI_ACCESS_TOKEN" "$POST_DATA"
else
    echo "***The $INFLUXDB_APP_NAME exsit! Please remove it and try it again"
fi

echo "Finish!"
exit 0
