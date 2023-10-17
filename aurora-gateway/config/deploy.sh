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


echo "Launching Edge Xpert with the required microservices..."
edgexpert up xpert-manager sys-mgmt device-modbus influxdb grafana

# Sleep
sleep 10s

echo "Uploading the device profile..."
##
## setup App-service through Xpert Manager
##
API_ENDPOINT='http://localhost:9090/api/v2'
GUI_TOKEN=$(curl --request POST --url ${API_ENDPOINT}/user/login --data '{"name":"admin","password":"admin"}')

GUI_ACCESS_TOKEN=$(echo $GUI_TOKEN | json_pp | grep access_token | cut -d'"' -f4)
echo $GUI_ACCESS_TOKEN

## create Influxdb exporter
POST_DATA='{"name":"influxdb-exporter","logLevel":"INFO","destination":"InfluxDB","influxDBSyncWrite":{"influxDBServerURL":"http://influxdb:8086","influxDBOrganization":"texol","influxDBBucket":"bucket","token":"texol-password","influxDBMeasurement":"readings","fieldKeyPattern":"{resourceName}","influxDBValueType":"float","influxDBPrecision":"us","authMode":"token","secretPath":"influxdb","skipVerify":"true","persistOnError":"false","storeEventTags":"false","storeReadingTags":"false"}}'
register_appservice "$API_ENDPOINT/appSvcConf" "$GUI_ACCESS_TOKEN" "$POST_DATA"


## Copy sensor picture to file system
sudo mkdir -p /usr/share/texol
sudo cp -rf picture/* /usr/share/texol/picture

echo "Finish!"
exit 0