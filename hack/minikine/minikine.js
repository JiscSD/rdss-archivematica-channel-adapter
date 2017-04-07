var kinesalite = require('kinesalite')
var AWS = require('aws-sdk')

var kinesaliteServer = kinesalite()

const PORT = 4567
const AWS_REGION = 'local-region'
const AWS_KINESIS_STREAM = 'foobar'
const AWS_KINESIS_SHARD_COUNT = 4

kinesaliteServer.listen(4567, function(err) {
  if (err) throw err
  console.log('Kinesalite started on port 4567')
  client()
})

function client() {
  var kinesis = new AWS.Kinesis({
    endpoint: 'http://127.0.0.1:' + PORT,
    region: AWS_REGION
  })
  var params = {
    ShardCount: AWS_KINESIS_SHARD_COUNT,
    StreamName: AWS_KINESIS_STREAM
  }
  kinesis.createStream(params, function(err, data) {
    if (err) {
      console.log(err, err.stack)
    } else {
      console.log('Stream created:', AWS_KINESIS_STREAM)
    }
  })
}
