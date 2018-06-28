var AWS = require('aws-sdk')
var kinesalite = require('kinesalite')

var kinesaliteServer = kinesalite()

const SETTINGS = {
  'port': process.env.MINIKINE_PORT || 4567,
  'region': process.env.MINIKINE_REGION || 'eu-west-2',

  // So far RDSS has defined four streams in their Messaging API
  'streamError': process.env.MINIKINE_STREAM_ERROR || 'error',
  'streamInput': process.env.MINIKINE_STREAM_INPUT || 'input',
  'streamInvalid': process.env.MINIKINE_STREAM_INVALID || 'invalid',
  'streamOutput': process.env.MINIKINE_STREAM_OUTPUT || 'output',

  // The number of streams that each shard is going to have
  'streamShards': process.env.MINIKINE_STREAM_SHARDS || 4
}

// Set up credentials in AWS-SDK
process.env.AWS_ACCESS_KEY_ID = 'XXXXXXXXXXXXXXXXXXX';
process.env.AWS_SECRET_ACCESS_KEY = 'XXXXXXXXXXXXXXXXXXXXXXXXXX';
process.env.AWS_REGION = SETTINGS.region;

// Start server
kinesaliteServer.listen(SETTINGS.port, function (err) {
  if (err) {
    throw err
  }
  console.log('Kinesalite started on port ' + SETTINGS.port)

  var kinesis = new AWS.Kinesis({endpoint: 'http://127.0.0.1:' + SETTINGS.port, region: SETTINGS.region})
  bootstrap(kinesis)

  console.log('Bootstrap finished!')
})


// This function creates the streams once the server has started.
function bootstrap(kinesis) {
  let streams = Array('streamError', 'streamInput', 'streamInvalid', 'streamOutput')
  for (let i in streams) {
    let stream = SETTINGS[streams[i]]
    let params = { ShardCount: SETTINGS.streamShards, StreamName: stream }
    kinesis.createStream(params, function (err, data) {
      if (err) {
        console.log('Error creating stream:', stream, err, err.stack)
      } else {
        console.log('Stream created:', stream)
      }
    })
  }
}
