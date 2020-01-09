#
# These env should be provided by CI runtime env variable.
# Only source this file for local testing purpose.
#
export PULSAR_TOKEN="eyJhbG"
export PULSAR_URI="pulsar+ssl://useast1.gcp.kafkaesque.io:6651"
export WEBHOOK_TOPIC="persistent://"
export REST_API_TOKEN="eyJhbG"
export WEBHOOK2_URL="http://localhost:8080/wh"
export FN_SINK_TOPIC="persistent://"

# a compacted Pulsar Topic can be used as Database table for RESTful API
export REST_DB_TABLE_TOPIC="persistent://<tenant>/<namespace>/<compacted topic>"
